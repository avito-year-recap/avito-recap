package server_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	connectrpc "connectrpc.com/connect"
	"github.com/google/uuid"
	recapv1 "github.com/year-recap/gen/go/recap/v1"
	"github.com/year-recap/gen/go/recap/v1/recapv1connect"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/server"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPHandlerSupportsCompleteConnectFlow(t *testing.T) {
	httpServer := newTestServer(t)
	client := recapv1connect.NewRecapServiceClient(httpServer.Client(), httpServer.URL)
	ctx := context.Background()

	profiles, err := client.ListProfiles(
		ctx,
		connectrpc.NewRequest(&recapv1.ListProfilesRequest{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles.Msg.Profiles) != 1 || profiles.Msg.Profiles[0].ProfileCode != testkit.Profile().Code {
		t.Fatalf("unexpected profiles: %+v", profiles.Msg.Profiles)
	}

	generated, err := client.GenerateRecap(
		ctx,
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
			ProfileCode: testkit.Profile().Code,
			Year:        2025,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if generated.Msg.Recap.Id != testkit.RecapID.String() {
		t.Fatalf("recap id = %q, want %q", generated.Msg.Recap.Id, testkit.RecapID)
	}

	fetched, err := client.GetRecap(
		ctx,
		connectrpc.NewRequest(&recapv1.GetRecapRequest{
			ProfileCode: testkit.Profile().Code,
			Year:        2025,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if fetched.Msg.Recap.Id != generated.Msg.Recap.Id {
		t.Fatalf("fetched recap id = %q", fetched.Msg.Recap.Id)
	}

	shareCard := findCard(t, generated.Msg.Recap.Cards, recapv1.CardType_CARD_TYPE_SHARE).GetShare()
	shared, err := client.GetPublicShare(
		ctx,
		connectrpc.NewRequest(&recapv1.GetPublicShareRequest{
			ShareId: shareCard.ShareId,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	data, err := protojson.Marshal(shared.Msg.Share)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"profile", "metrics", "nextAction"} {
		if strings.Contains(string(data), forbidden) {
			t.Fatalf("public share contains %q: %s", forbidden, data)
		}
	}
}

func TestHTTPHandlerMapsInvalidRequestToConnectError(t *testing.T) {
	httpServer := newTestServer(t)
	client := recapv1connect.NewRecapServiceClient(httpServer.Client(), httpServer.URL)
	_, err := client.GenerateRecap(
		context.Background(),
		connectrpc.NewRequest(&recapv1.GenerateRecapRequest{
			ProfileCode: "",
			Year:        2025,
		}),
	)
	if code := connectrpc.CodeOf(err); code != connectrpc.CodeInvalidArgument {
		t.Fatalf("code = %s, want invalid_argument: %v", code, err)
	}
}

func TestHTTPHandlerHealthCORSAndAvatars(t *testing.T) {
	httpServer := newTestServer(t)

	response, err := httpServer.Client().Get(httpServer.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", response.StatusCode)
	}
	closeResponseBody(t, response)

	request, err := http.NewRequest(
		http.MethodOptions,
		httpServer.URL+recapv1connect.RecapServiceListProfilesProcedure,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	response, err = httpServer.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("preflight status = %d", response.StatusCode)
	}
	if actual := response.Header.Get("Access-Control-Allow-Origin"); actual != "http://localhost:5173" {
		t.Fatalf("allow origin = %q", actual)
	}
	closeResponseBody(t, response)

	request, err = http.NewRequest(http.MethodGet, httpServer.URL+"/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", "https://untrusted.example")
	response, err = httpServer.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("untrusted origin status = %d", response.StatusCode)
	}
	closeResponseBody(t, response)

	response, err = httpServer.Client().Get(httpServer.URL + "/avatars/active-buyer.png")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	closeResponseBody(t, response)
	if response.StatusCode != http.StatusOK || len(data) == 0 {
		t.Fatalf("avatar status=%d size=%d", response.StatusCode, len(data))
	}
}

func closeResponseBody(t *testing.T, response *http.Response) {
	t.Helper()
	if err := response.Body.Close(); err != nil {
		t.Errorf("close response body: %v", err)
	}
}

func findCard(t *testing.T, cards []*recapv1.RecapCard, cardType recapv1.CardType) *recapv1.RecapCard {
	t.Helper()
	for _, card := range cards {
		if card.Type == cardType {
			return card
		}
	}
	t.Fatalf("no card of type %v", cardType)
	return nil
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	recaps := &testkit.RecapStorage{}
	service, err := application.NewService(
		&testkit.ProfileStorage{Profile: testkit.Profile()},
		&testkit.AnalyticsStorage{Metrics: testkit.Metrics()},
		&testkit.ActionStateStorage{State: model.ActionableState{
			FavoritesCount:          5,
			HasEverPublishedListing: true,
		}},
		recaps,
		application.WithClock(testkit.Clock),
		application.WithIDGenerator(sequenceIDGenerator(testkit.RecapID, testkit.ShareID)),
	)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := server.NewHandler(service, server.Options{
		StaticDir:      filepath.Join(projectRoot(t), "frontend", "public"),
		AllowedOrigins: []string{"http://localhost:5173"},
	})
	if err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)
	return httpServer
}

func sequenceIDGenerator(values ...uuid.UUID) application.IDGenerator {
	index := 0
	return func() (uuid.UUID, error) {
		value := values[index]
		index++
		return value, nil
	}
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(current), "..", ".."))
}
