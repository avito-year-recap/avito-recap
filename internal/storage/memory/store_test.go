package memory_test

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/seed"
	"github.com/year-recap/internal/storage/memory"
)

func TestSeedStoreSupportsCompleteRecapFlow(t *testing.T) {
	store := loadStore(t)
	profiles, err := store.ListProfiles(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 17 {
		t.Fatalf("profile count = %d, want 17", len(profiles))
	}
	if profiles[0].Code != "active-buyer" {
		t.Fatalf("catalogue order was not preserved: first = %q", profiles[0].Code)
	}

	now := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)
	ids := []uuid.UUID{
		uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
	}
	idIndex := 0
	service, err := application.NewService(
		store,
		store,
		store,
		store,
		application.WithClock(func() time.Time { return now }),
		application.WithIDGenerator(func() (uuid.UUID, error) {
			value := ids[idIndex]
			idIndex++
			return value, nil
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	profileID := uuid.MustParse("26a3f4e0-1ae7-5b46-b2b6-2ae9fc180ba2")
	first, err := service.Generate(context.Background(), profileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Generate(context.Background(), profileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("idempotent recap ids differ: %s and %s", first.ID, second.ID)
	}
	if first.ActionableState.CapturedAt != now {
		t.Fatalf("snapshot time = %s, want %s", first.ActionableState.CapturedAt, now)
	}
	if first.Behavior.Code != model.BehaviorFindHunter {
		t.Fatalf("behavior = %s, want %s", first.Behavior.Code, model.BehaviorFindHunter)
	}

	stored, err := service.Get(context.Background(), first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ID != first.ID {
		t.Fatalf("stored recap id = %s, want %s", stored.ID, first.ID)
	}
	share, err := service.GetShareCard(context.Background(), first.ShareID)
	if err != nil {
		t.Fatal(err)
	}
	if share.ShareID != first.ShareID || share.BehaviorTitle != first.Behavior.Title {
		t.Fatalf("unexpected share projection: %+v", share)
	}
}

func TestStoreReturnsNotFoundForUnknownValues(t *testing.T) {
	store := loadStore(t)
	unknown := uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")
	if _, err := store.GetProfile(context.Background(), unknown); !errors.Is(err, application.ErrProfileNotFound) {
		t.Fatalf("profile error = %v", err)
	}
	if _, err := store.GetRecap(context.Background(), unknown); !errors.Is(err, application.ErrRecapNotFound) {
		t.Fatalf("recap error = %v", err)
	}
	if _, err := store.GetRecapByShareID(context.Background(), unknown); !errors.Is(err, application.ErrRecapNotFound) {
		t.Fatalf("share error = %v", err)
	}
}

func TestNewRejectsDuplicateProfileCode(t *testing.T) {
	profiles := []model.Profile{
		{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Code: "same", DisplayName: "One"},
		{ID: uuid.MustParse("22222222-2222-4222-8222-222222222222"), Code: "same", DisplayName: "Two"},
	}
	if _, err := memory.New(profiles, nil); !errors.Is(err, memory.ErrInvalidSeedData) {
		t.Fatalf("error = %v, want invalid seed data", err)
	}
}

func TestNewRejectsScenarioWithoutProfile(t *testing.T) {
	profiles := []model.Profile{{
		ID:          uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		Code:        "known",
		DisplayName: "Known",
	}}
	scenarios := []seed.Scenario{{ProfileCode: "missing", Year: 2025}}
	if _, err := memory.New(profiles, scenarios); !errors.Is(err, memory.ErrInvalidSeedData) {
		t.Fatalf("error = %v, want invalid seed data", err)
	}
}

func TestStoreHonorsCancelledContext(t *testing.T) {
	store := loadStore(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.ListProfiles(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context canceled", err)
	}
}

func loadStore(t *testing.T) *memory.Store {
	t.Helper()
	root := projectRoot(t)
	store, err := memory.Load(
		filepath.Join(root, "seeds", "profiles.json"),
		filepath.Join(root, "seeds", "scenarios.json"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(current), "..", "..", ".."))
}
