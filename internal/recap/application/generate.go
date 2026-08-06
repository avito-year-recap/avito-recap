package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/integrity"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
	"github.com/year-recap/internal/recap/presentation/cards"
	"github.com/year-recap/internal/recap/validation/structural"
)

const MinEventsForRecap uint64 = 5

func (s *Service) Generate(ctx context.Context, profileID uuid.UUID, year uint32) (model.Recap, error) {
	if profileID == uuid.Nil {
		return model.Recap{}, ErrInvalidProfileID
	}
	now := s.now().UTC()
	period, err := analytics.CompletedYearPeriod(year, now)
	if err != nil {
		return model.Recap{}, err
	}
	key := model.RecapKey{ProfileID: profileID, Year: year, RulesVersion: s.ruleset.Version, RulesDigest: s.ruleset.Digest()}

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
	if err := structural.ValidateProfile(profile); err != nil {
		return model.Recap{}, err
	}
	if profile.ID != profileID {
		return model.Recap{}, fmt.Errorf("%w: requested %s, got %s", ErrProfileIDMismatch, profileID, profile.ID)
	}

	metrics, err := s.analytics.CalculateMetrics(ctx, profileID, period)
	if err != nil {
		return model.Recap{}, fmt.Errorf("calculate metrics: %w", err)
	}
	metrics = model.NormalizeMetrics(metrics)
	if err := structural.ValidateMetricsForPeriod(metrics, period); err != nil {
		return model.Recap{}, err
	}
	if metrics.TotalEvents < MinEventsForRecap {
		return model.Recap{}, ErrNotEnoughActivity
	}
	metrics = analytics.EnrichMetrics(metrics)

	state, err := s.actionStates.GetActionableState(ctx, profileID, now)
	if err != nil {
		return model.Recap{}, fmt.Errorf("get actionable state: %w", err)
	}
	state = model.NormalizeActionableState(state)
	if err := structural.ValidateActionableState(state); err != nil {
		return model.Recap{}, err
	}
	if !state.CapturedAt.Equal(now) {
		return model.Recap{}, fmt.Errorf("%w: snapshot captured at %s, requested %s", structural.ErrInvalidActionableState, state.CapturedAt, now)
	}

	detectedBehavior := behavior.DetectWithRuleset(s.ruleset, metrics)
	achievements := achievement.BuildWithRuleset(s.ruleset, metrics)
	nextAction := nextaction.BuildWithRuleset(s.ruleset, metrics, state, detectedBehavior)

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
	cards := cards.BuildWithRuleset(s.ruleset, profile, year, shareID, metrics, detectedBehavior, achievements, nextAction)

	candidate := model.Recap{
		ID: recapID, ShareID: shareID, Profile: profile, Year: year, Period: period,
		RulesVersion: s.ruleset.Version, RulesDigest: s.ruleset.Digest(), Metrics: metrics, ActionableState: state,
		Behavior: detectedBehavior, Achievements: achievements, Cards: cards, NextAction: nextAction,
		GeneratedAt: now,
	}
	if err := integrity.ValidateRecapAgainstRuleset(candidate, s.ruleset, now); err != nil {
		return model.Recap{}, fmt.Errorf("validate generated recap: %w", err)
	}

	// The storage operation is the concurrency boundary. It must insert candidate or
	// atomically return the already stored value for the same unique key.
	stored, err := s.recaps.CreateRecapIfAbsent(ctx, key, candidate)
	if err != nil {
		return model.Recap{}, fmt.Errorf("create recap if absent: %w", err)
	}
	return s.validateStoredByKey(stored, key)
}
