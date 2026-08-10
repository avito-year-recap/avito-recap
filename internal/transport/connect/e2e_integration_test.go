//go:build integration

package connect_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	connectrpc "connectrpc.com/connect"
	"github.com/google/uuid"
	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
	storage "github.com/year-recap/internal/storage/clickhouse"
	transportconnect "github.com/year-recap/internal/transport/connect"
)

func TestConnectAPIEndToEndWithClickHouse(t *testing.T) {
	ctx := context.Background()
	repo, err := storage.Connect(ctx, testDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if err := repo.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}

	profile := testkit.Profile()
	profile.ID = uuid.New()
	profile.Code = "e2e-" + profile.ID.String()
	if err := repo.UpsertProfiles(ctx, []model.Profile{profile}); err != nil {
		t.Fatal(err)
	}
	events := testEvents(profile.ID)

	if err := repo.InsertEvents(ctx, events); err != nil {
		t.Fatal(err)
	}
	if err := repo.PutActionableState(ctx, profile.ID, testkit.Clock().Add(-time.Minute), model.ActionableState{
		FavoritesCount: 5, HasEverPublishedListing: true,
	}); err != nil {
		t.Fatal(err)
	}

	ids := []uuid.UUID{uuid.New(), uuid.New()}
	index := 0
	app, err := application.NewService(repo, repo, repo, repo,
		application.WithClock(testkit.Clock),
		application.WithIDGenerator(func() (uuid.UUID, error) {
			value := ids[index]
			index++
			return value, nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := transportconnect.NewHandler(app)
	if err != nil {
		t.Fatal(err)
	}
	path, rpcHandler := recapv1connect.NewRecapServiceHandler(handler)
	mux := http.NewServeMux()
	mux.Handle(path, rpcHandler)
	server := httptest.NewServer(mux)
	defer server.Close()

	client := recapv1connect.NewRecapServiceClient(server.Client(), server.URL)
	listed, err := client.ListProfiles(ctx, connectrpc.NewRequest(&recapv1.ListProfilesRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	foundProfile := false
	for _, item := range listed.Msg.Profiles {
		if item.ProfileCode == profile.Code {
			foundProfile = true
			break
		}
	}
	if !foundProfile {
		t.Fatalf("ListProfiles did not return seeded profile code %q", profile.Code)
	}

	generated, err := client.GenerateRecap(ctx, connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
		ProfileCode: profile.Code,
		Year:        2025,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if generated.Msg.Recap == nil || generated.Msg.Recap.Id == "" {
		t.Fatalf("incomplete generated recap: %+v", generated.Msg.Recap)
	}

	repeated, err := client.GenerateRecap(ctx, connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
		ProfileCode: profile.Code,
		Year:        2025,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if repeated.Msg.Recap == nil || repeated.Msg.Recap.Id != generated.Msg.Recap.Id {
		t.Fatalf("GenerateRecap is not idempotent through HTTP: first=%q second=%q",
			generated.Msg.Recap.Id, repeated.Msg.Recap.GetId())
	}

	got, err := client.GetRecap(ctx, connectrpc.NewRequest(&recapv1.GetRecapRequest{
		ProfileCode: profile.Code,
		Year:        2025,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Msg.Recap == nil || got.Msg.Recap.Id != generated.Msg.Recap.Id {
		t.Fatalf("GetRecap id = %q, want %q", got.Msg.Recap.GetId(), generated.Msg.Recap.Id)
	}

	sharePayload := findSharePayload(t, generated.Msg.Recap.Cards)
	shared, err := client.GetPublicShare(ctx, connectrpc.NewRequest(&recapv1.GetPublicShareRequest{
		ShareId: sharePayload.ShareId,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if shared.Msg.Share == nil || shared.Msg.Share.ShareId != sharePayload.ShareId {
		t.Fatalf("share response mismatch: %+v", shared.Msg.Share)
	}
}

func findSharePayload(t *testing.T, cards []*recapv1.RecapCard) *recapv1.SharePayload {
	t.Helper()
	for _, card := range cards {
		if card != nil && card.Type == recapv1.CardType_CARD_TYPE_SHARE {
			payload := card.GetShare()
			if payload == nil || payload.ShareId == "" {
				t.Fatalf("share card has no share payload: %+v", card)
			}
			return payload
		}
	}
	t.Fatal("generated recap has no share card")
	return nil
}

func testEvents(profileID uuid.UUID) []model.Event {
	times := []time.Time{
		time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 2, 10, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 3, 10, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 4, 10, 12, 0, 0, 0, time.UTC),
		time.Date(2025, 5, 10, 12, 0, 0, 0, time.UTC),
	}

	events := make([]model.Event, 0, len(times))

	for _, occurredAt := range times {
		events = append(events, model.Event{
			ID:         uuid.New(),
			ProfileID:  profileID,
			Type:       model.ActivitySearch,
			OccurredAt: occurredAt,
		})
	}

	return events
}

func testDSN() string {
	if value := os.Getenv("CLICKHOUSE_TEST_DSN"); value != "" {
		return value
	}
	return "clickhouse://recap:recap@localhost:9000/recap"
}
