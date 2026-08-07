package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
)

func TestServiceGeneratesAndReusesFinalRecap(t *testing.T) {
	profiles := &testkit.ProfileStorage{Profile: testkit.Profile()}
	analytics := &testkit.AnalyticsStorage{Metrics: testkit.Metrics()}
	states := &testkit.ActionStateStorage{State: model.ActionableState{FavoritesCount: 5, HasEverPublishedListing: true}}
	recaps := &testkit.RecapStorage{}
	ids := []uuid.UUID{testkit.RecapID, testkit.ShareID}
	index := 0
	service, err := application.NewService(profiles, analytics, states, recaps,
		application.WithClock(testkit.Clock),
		application.WithIDGenerator(func() (uuid.UUID, error) { value := ids[index]; index++; return value, nil }),
	)
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || recaps.Creates != 1 || analytics.Calls != 1 {
		t.Fatalf("not idempotent: first=%s second=%s creates=%d analytics=%d", first.ID, second.ID, recaps.Creates, analytics.Calls)
	}
	if first.NextAction.Title != "Вернись к своим находкам" {
		t.Fatalf("unexpected action: %+v", first.NextAction)
	}
}

func TestServiceRejectsCurrentYearBeforeStorageCalls(t *testing.T) {
	service, err := application.NewService(&testkit.ProfileStorage{Profile: testkit.Profile()}, &testkit.AnalyticsStorage{Metrics: testkit.Metrics()}, &testkit.ActionStateStorage{}, &testkit.RecapStorage{}, application.WithClock(testkit.Clock))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Generate(context.Background(), testkit.ProfileID, uint32(testkit.Clock().Year())); err == nil {
		t.Fatal("expected incomplete year error")
	}
}
