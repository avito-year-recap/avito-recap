//go:build integration

package clickhouse_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
	storage "github.com/year-recap/internal/storage/clickhouse"
)

func TestRepoImplementsApplicationStorages(t *testing.T) {
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
	profile.Code = "integration-" + profile.ID.String()
	if err := repo.UpsertProfiles(ctx, []model.Profile{profile}); err != nil {
		t.Fatal(err)
	}
	storedProfile, err := repo.GetProfile(ctx, profile.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedProfile.ID != profile.ID || storedProfile.Code != profile.Code {
		t.Fatalf("profile mismatch: got %+v want %+v", storedProfile, profile)
	}

	metrics := testkit.Metrics()
	if err := repo.UpsertAnnualMetrics(ctx, profile.ID, 2025, metrics); err != nil {
		t.Fatal(err)
	}
	period := testkit.Period()
	storedMetrics, err := repo.CalculateMetrics(ctx, profile.ID, period)
	if err != nil {
		t.Fatal(err)
	}
	if storedMetrics.TotalEvents != metrics.TotalEvents || storedMetrics.TopCategoryCode != metrics.TopCategoryCode {
		t.Fatalf("metrics mismatch: got %+v want %+v", storedMetrics, metrics)
	}

	state := model.ActionableState{FavoritesCount: 7, HasEverPublishedListing: true}
	asOf := testkit.Clock()
	if err := repo.PutActionableState(ctx, profile.ID, asOf.Add(-time.Minute), state); err != nil {
		t.Fatal(err)
	}
	storedState, err := repo.GetActionableState(ctx, profile.ID, asOf)
	if err != nil {
		t.Fatal(err)
	}
	if storedState.FavoritesCount != 7 || !storedState.CapturedAt.Equal(asOf) {
		t.Fatalf("actionable state mismatch: %+v", storedState)
	}

	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	index := 0
	service, err := application.NewService(repo, repo, repo, repo,
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
	first, err := service.Generate(ctx, profile.ID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Generate(ctx, profile.ID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || first.ShareID != second.ShareID {
		t.Fatalf("recap is not idempotent: first=%s/%s second=%s/%s", first.ID, first.ShareID, second.ID, second.ShareID)
	}
	byID, err := repo.GetRecap(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	byShare, err := repo.GetRecapByShareID(ctx, first.ShareID)
	if err != nil {
		t.Fatal(err)
	}
	if byID.ID != first.ID || byShare.ID != first.ID {
		t.Fatalf("lookup mismatch: byID=%s byShare=%s want=%s", byID.ID, byShare.ID, first.ID)
	}
	if !reflect.DeepEqual(model.NormalizeRecap(byID), model.NormalizeRecap(first)) {
		t.Fatalf("immutable recap round trip changed the domain object")
	}
}

func testDSN() string {
	if value := os.Getenv("CLICKHOUSE_TEST_DSN"); value != "" {
		return value
	}
	return "clickhouse://recap:recap@localhost:9000/recap"
}

func TestRepoSerializesConcurrentRecapCreationAcrossConnections(t *testing.T) {
	ctx := context.Background()
	repoA, err := storage.Connect(ctx, testDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer repoA.Close()
	repoB, err := storage.Connect(ctx, testDSN())
	if err != nil {
		t.Fatal(err)
	}
	defer repoB.Close()
	if err := repoA.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}

	profile := testkit.Profile()
	profile.ID = uuid.New()
	profile.Code = "concurrent-" + profile.ID.String()
	if err := repoA.UpsertProfiles(ctx, []model.Profile{profile}); err != nil {
		t.Fatal(err)
	}
	if err := repoA.UpsertAnnualMetrics(ctx, profile.ID, 2025, testkit.Metrics()); err != nil {
		t.Fatal(err)
	}
	if err := repoA.PutActionableState(ctx, profile.ID, testkit.Clock().Add(-time.Minute), model.ActionableState{
		FavoritesCount: 5, HasEverPublishedListing: true,
	}); err != nil {
		t.Fatal(err)
	}

	services := make([]*application.Service, 0, 2)
	for _, repo := range []*storage.Repo{repoA, repoB} {
		service, err := application.NewService(repo, repo, repo, repo, application.WithClock(testkit.Clock))
		if err != nil {
			t.Fatal(err)
		}
		services = append(services, service)
	}

	const workers = 8
	results := make(chan uuid.UUID, workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		service := services[i%len(services)]
		go func() {
			value, err := service.Generate(ctx, profile.ID, 2025)
			if err != nil {
				errs <- err
				return
			}
			results <- value.ID
		}()
	}
	var winner uuid.UUID
	for i := 0; i < workers; i++ {
		select {
		case err := <-errs:
			t.Fatal(err)
		case id := <-results:
			if winner == uuid.Nil {
				winner = id
			} else if id != winner {
				t.Fatalf("concurrent generation returned multiple ids: %s and %s", winner, id)
			}
		}
	}
}
