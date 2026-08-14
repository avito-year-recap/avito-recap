package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/insight"
)

const MaxInsightRangeDays = 366

func (s *Service) AnalyzeBehavior(ctx context.Context, profileID uuid.UUID, start, end time.Time) (insight.Result, error) {
	if profileID == uuid.Nil {
		return insight.Result{}, ErrInvalidProfileID
	}
	if s.events == nil {
		return insight.Result{}, ErrEventRangeUnsupported
	}
	if s.insight == nil {
		return insight.Result{}, ErrInsightUnavailable
	}

	start = start.UTC()
	end = end.UTC()
	now := s.now().UTC()
	if !start.Before(end) {
		return insight.Result{}, fmt.Errorf("%w: start must be before end", ErrInvalidPeriod)
	}
	if end.After(now) {
		return insight.Result{}, fmt.Errorf("%w: end must not be in the future", ErrInvalidPeriod)
	}
	if end.Sub(start) > MaxInsightRangeDays*24*time.Hour {
		return insight.Result{}, fmt.Errorf("%w: maximum is %d days", ErrPeriodTooLong, MaxInsightRangeDays)
	}

	profile, err := s.profiles.GetProfile(ctx, profileID)
	if err != nil {
		return insight.Result{}, fmt.Errorf("get profile: %w", err)
	}

	events, err := s.events.QueryEventsByRange(ctx, profileID, start, end)
	if err != nil {
		return insight.Result{}, fmt.Errorf("query events by range: %w", err)
	}
	if len(events) == 0 {
		return insight.Result{}, ErrNoActivityInPeriod
	}

	metrics := analytics.AggregateEvents(events)
	facts := insight.BuildFacts(profile.Code, start, end, metrics)

	card, err := s.insight.Generate(ctx, facts)
	if err != nil {
		return insight.Result{}, fmt.Errorf("generate behavior insight: %w", err)
	}
	if err := insight.ValidateCard(card); err != nil {
		return insight.Result{}, fmt.Errorf("validate behavior insight: %w", err)
	}

	return insight.Result{
		ProfileCode: profile.Code,
		StartAt:     start,
		EndAt:       end,
		Card:        card,
		Facts:       facts,
	}, nil
}
