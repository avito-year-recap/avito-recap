package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/testkit"
)

func TestGenerateRejectsStorageThatFalsifiesSnapshotTime(t *testing.T) {
	state := testkit.ActionableState()
	state.CapturedAt = testkit.Clock().Add(-time.Minute)
	service, err := application.NewService(
		&testkit.ProfileStorage{Profile: testkit.Profile()},
		&testkit.AnalyticsStorage{Metrics: testkit.Metrics()},
		&testkit.ActionStateStorage{State: state},
		&testkit.RecapStorage{},
		application.WithClock(testkit.Clock),
		application.WithIDGenerator(sequenceIDGenerator(testkit.RecapID, testkit.ShareID)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Generate(context.Background(), testkit.ProfileID, 2025); !errors.Is(err, application.ErrInvalidActionableState) {
		t.Fatalf("stale snapshot error = %v", err)
	}
}
