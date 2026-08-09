package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/engine"
	"github.com/year-recap/internal/recap/model"
)

func (s *Service) Generate(ctx context.Context, profileID uuid.UUID, year uint32) (model.Recap, error) {
	if profileID == uuid.Nil {
		return model.Recap{}, ErrInvalidProfileID
	}
	now := s.now().UTC()
	period, err := analytics.CompletedYearPeriod(year, now)
	if err != nil {
		return model.Recap{}, err
	}
	key := s.engine.RecapKey(profileID, year)

	if existing, err := s.recaps.GetRecapByKey(ctx, key); err == nil {
		return s.validateStoredByKey(existing, key)
	} else if !errors.Is(err, ErrRecapNotFound) {
		return model.Recap{}, fmt.Errorf("get recap by idempotency key: %w", err)
	}

	profile, err := s.profiles.GetProfile(ctx, profileID)
	if err != nil {
		return model.Recap{}, fmt.Errorf("get profile: %w", err)
	}
	profile = model.NormalizeProfile(profile)
	if profile.ID != profileID {
		return model.Recap{}, fmt.Errorf("%w: requested %s, got %s", ErrProfileIDMismatch, profileID, profile.ID)
	}

	metrics, err := s.analytics.CalculateMetrics(ctx, profileID, period)
	if err != nil {
		return model.Recap{}, fmt.Errorf("calculate metrics: %w", err)
	}

	state, err := s.actionStates.GetActionableState(ctx, profileID, now)
	if err != nil {
		return model.Recap{}, fmt.Errorf("get actionable state: %w", err)
	}

	recapID, err := s.generateNonNilID("internal recap")
	if err != nil {
		return model.Recap{}, err
	}
	shareID, err := s.generateNonNilID("public share")
	if err != nil {
		return model.Recap{}, err
	}
	if recapID == shareID {
		return model.Recap{}, fmt.Errorf("%w: internal and public ids must differ", ErrGenerateID)
	}

	candidate, err := s.engine.Build(engine.BuildInput{
		RecapID:         recapID,
		ShareID:         shareID,
		Profile:         profile,
		Year:            year,
		Period:          period,
		Metrics:         metrics,
		ActionableState: state,
		GeneratedAt:     now,
	})
	if err != nil {
		return model.Recap{}, err
	}

	// Persistence remains the concurrency/idempotency boundary. Business
	// derivation is already complete before the storage call.
	stored, err := s.recaps.CreateRecapIfAbsent(ctx, key, candidate)
	if err != nil {
		return model.Recap{}, fmt.Errorf("create recap if absent: %w", err)
	}
	return s.validateStoredByKey(stored, key)
}
