package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/insight"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
)

type insightStub struct {
	received insight.Facts
	card     insight.Card
	err      error
}

func (s *insightStub) Generate(_ context.Context, facts insight.Facts) (insight.Card, error) {
	s.received = facts
	if s.err != nil {
		return insight.Card{}, s.err
	}
	return s.card, nil
}

func eventAt(profileID uuid.UUID, occurredAt time.Time, activityType model.ActivityType) model.Event {
	return model.Event{ID: uuid.New(), ProfileID: profileID, Type: activityType, OccurredAt: occurredAt}
}

func newBehaviorInsightService(t *testing.T, options ...application.Option) (*application.Service, *testkit.EventRangeStorage, *insightStub) {
	t.Helper()
	events := &testkit.EventRangeStorage{}
	generator := &insightStub{card: insight.Card{Title: "Заголовок", Description: "Описание поведения."}}
	allOptions := append([]application.Option{
		application.WithClock(testkit.Clock),
		application.WithEventRangeStorage(events),
		application.WithInsightGenerator(generator),
	}, options...)
	service, err := application.NewService(
		&testkit.ProfileStorage{Profile: testkit.Profile()},
		&testkit.AnalyticsStorage{Metrics: testkit.Metrics()},
		&testkit.ActionStateStorage{},
		&testkit.RecapStorage{},
		allOptions...,
	)
	if err != nil {
		t.Fatal(err)
	}
	return service, events, generator
}

func TestAnalyzeBehaviorReturnsCardWithGroundingFacts(t *testing.T) {
	start := testkit.Clock().AddDate(0, 0, -7)
	end := testkit.Clock()
	service, events, generator := newBehaviorInsightService(t)
	events.Events = []model.Event{
		eventAt(testkit.ProfileID, start.Add(time.Hour), model.ActivitySearch),
		eventAt(testkit.ProfileID, start.Add(2*time.Hour), model.ActivityListingView),
		eventAt(testkit.ProfileID, end.Add(time.Hour), model.ActivitySearch), // outside range, must be excluded
	}

	result, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if result.Card.Title != "Заголовок" {
		t.Fatalf("card = %+v", result.Card)
	}
	if result.ProfileCode != testkit.Profile().Code {
		t.Fatalf("profile code = %q", result.ProfileCode)
	}
	if result.Facts.Metrics.TotalEvents != 2 {
		t.Fatalf("facts should only reflect in-range events: %+v", result.Facts.Metrics)
	}
	if generator.received.Metrics.TotalEvents != 2 {
		t.Fatalf("generator did not receive the same facts returned to caller: %+v", generator.received)
	}
}

func TestAnalyzeBehaviorRejectsInvalidPeriod(t *testing.T) {
	service, _, _ := newBehaviorInsightService(t)
	now := testkit.Clock()

	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, now, now); !errors.Is(err, application.ErrInvalidPeriod) {
		t.Fatalf("start == end: err = %v, want ErrInvalidPeriod", err)
	}
	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, now, now.AddDate(0, 0, -1)); !errors.Is(err, application.ErrInvalidPeriod) {
		t.Fatalf("start after end: err = %v, want ErrInvalidPeriod", err)
	}
	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, now.AddDate(0, 0, -1), now.Add(time.Hour)); !errors.Is(err, application.ErrInvalidPeriod) {
		t.Fatalf("end in the future: err = %v, want ErrInvalidPeriod", err)
	}
}

func TestAnalyzeBehaviorRejectsPeriodLongerThanMax(t *testing.T) {
	service, _, _ := newBehaviorInsightService(t)
	now := testkit.Clock()
	start := now.AddDate(0, 0, -(application.MaxInsightRangeDays + 1))

	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, start, now); !errors.Is(err, application.ErrPeriodTooLong) {
		t.Fatalf("err = %v, want ErrPeriodTooLong", err)
	}
}

func TestAnalyzeBehaviorRequiresEventRangeStorage(t *testing.T) {
	service, err := application.NewService(
		&testkit.ProfileStorage{Profile: testkit.Profile()},
		&testkit.AnalyticsStorage{Metrics: testkit.Metrics()},
		&testkit.ActionStateStorage{},
		&testkit.RecapStorage{},
		application.WithClock(testkit.Clock),
		application.WithInsightGenerator(&insightStub{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	now := testkit.Clock()
	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, now.AddDate(0, 0, -1), now); !errors.Is(err, application.ErrEventRangeUnsupported) {
		t.Fatalf("err = %v, want ErrEventRangeUnsupported", err)
	}
}

func TestAnalyzeBehaviorRequiresInsightGenerator(t *testing.T) {
	service, err := application.NewService(
		&testkit.ProfileStorage{Profile: testkit.Profile()},
		&testkit.AnalyticsStorage{Metrics: testkit.Metrics()},
		&testkit.ActionStateStorage{},
		&testkit.RecapStorage{},
		application.WithClock(testkit.Clock),
		application.WithEventRangeStorage(&testkit.EventRangeStorage{}),
	)
	if err != nil {
		t.Fatal(err)
	}
	now := testkit.Clock()
	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, now.AddDate(0, 0, -1), now); !errors.Is(err, application.ErrInsightUnavailable) {
		t.Fatalf("err = %v, want ErrInsightUnavailable", err)
	}
}

func TestAnalyzeBehaviorRejectsEmptyPeriod(t *testing.T) {
	service, events, _ := newBehaviorInsightService(t)
	events.Events = nil
	now := testkit.Clock()

	if _, err := service.AnalyzeBehavior(context.Background(), testkit.ProfileID, now.AddDate(0, 0, -1), now); !errors.Is(err, application.ErrNoActivityInPeriod) {
		t.Fatalf("err = %v, want ErrNoActivityInPeriod", err)
	}
}
