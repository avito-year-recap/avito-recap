package application

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	// Fast path: most repeated reads avoid singleflight bookkeeping entirely.
	if existing, err := s.recaps.GetRecapByKey(ctx, key); err == nil {
		return s.validateStoredByKey(existing, key)
	} else if !errors.Is(err, ErrRecapNotFound) {
		return model.Recap{}, fmt.Errorf("get recap by idempotency key: %w", err)
	}

	// Slow path: concurrent requests for the same profile/year/ruleset share one
	// expensive generation (metrics + rules + optional AI). Followers wait for
	// the leader instead of issuing duplicate work or duplicate AI requests.
	return s.generateFlights.Do(ctx, key, func(sharedCtx context.Context) (model.Recap, error) {
		// Re-check after entering the flight: another generation may have completed
		// between the fast-path lookup and this call.
		if existing, err := s.recaps.GetRecapByKey(sharedCtx, key); err == nil {
			return s.validateStoredByKey(existing, key)
		} else if !errors.Is(err, ErrRecapNotFound) {
			return model.Recap{}, fmt.Errorf("get recap by idempotency key after singleflight: %w", err)
		}

		return s.generateOnce(sharedCtx, profileID, year, period, now, key)
	})
}

func (s *Service) generateOnce(
	ctx context.Context,
	profileID uuid.UUID,
	year uint32,
	period model.RecapPeriod,
	now time.Time,
	key model.RecapKey,
) (model.Recap, error) {
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

	if s.narrative != nil {
		candidate, err = s.narrative.Enrich(ctx, candidate)
		if err != nil {
			return model.Recap{}, err
		}
	}

	// singleflight only removes duplicate work inside one application instance.
	// The current ClickHouse adapter is intentionally single-writer; a multi-replica
	// production deployment needs a distributed lock or atomic uniqueness elsewhere.
	stored, err := s.recaps.CreateRecapIfAbsent(ctx, key, candidate)
	if err != nil {
		return model.Recap{}, fmt.Errorf("create recap if absent: %w", err)
	}
	return s.validateStoredByKey(stored, key)
}
