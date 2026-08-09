package engine

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
	"github.com/year-recap/internal/recap/presentation/cards"
	"github.com/year-recap/internal/recap/presentation/share"
	"github.com/year-recap/internal/recap/validation/structural"
)

type BuildInput struct {
	RecapID         uuid.UUID
	ShareID         uuid.UUID
	Profile         model.Profile
	Year            uint32
	Period          model.RecapPeriod
	Metrics         model.Metrics
	ActionableState model.ActionableState
	GeneratedAt     time.Time
}

type derived struct {
	Behavior     model.Behavior
	Achievements []model.Achievement
	NextAction   model.NextAction
	ShareCard    model.ShareCard
	Cards        []model.Card
}

func (e *Engine) Build(input BuildInput) (model.Recap, error) {
	input.GeneratedAt = input.GeneratedAt.UTC()
	input.Profile = model.NormalizeProfile(input.Profile)
	input.Metrics = model.NormalizeMetrics(input.Metrics)
	input.ActionableState = model.NormalizeActionableState(input.ActionableState)

	if input.RecapID == uuid.Nil {
		return model.Recap{}, fmt.Errorf("%w: internal id is required", structural.ErrInvalidRecap)
	}
	if input.ShareID == uuid.Nil {
		return model.Recap{}, fmt.Errorf("%w: public share id is required", structural.ErrInvalidRecap)
	}
	if input.RecapID == input.ShareID {
		return model.Recap{}, fmt.Errorf("%w: internal and public ids must differ", structural.ErrInvalidRecap)
	}
	if err := structural.ValidateProfile(input.Profile); err != nil {
		return model.Recap{}, err
	}
	if input.Year == 0 || input.Period.Year != input.Year {
		return model.Recap{}, fmt.Errorf("%w: year and period are inconsistent", structural.ErrInvalidRecap)
	}
	if err := structural.ValidatePeriod(input.Period); err != nil {
		return model.Recap{}, fmt.Errorf("%w: period: %w", structural.ErrInvalidRecap, err)
	}
	if err := structural.ValidateMetricsForPeriod(input.Metrics, input.Period); err != nil {
		return model.Recap{}, err
	}
	if err := e.ensureEligible(input.Metrics); err != nil {
		return model.Recap{}, err
	}
	if err := structural.ValidateActionableState(input.ActionableState); err != nil {
		return model.Recap{}, err
	}
	if !input.ActionableState.CapturedAt.Equal(input.GeneratedAt) {
		return model.Recap{}, fmt.Errorf("%w: snapshot captured at %s, generated at %s", structural.ErrInvalidActionableState, input.ActionableState.CapturedAt, input.GeneratedAt)
	}

	// Canonicalization is deliberately centralized here. Domain modules receive
	// normalized/enriched values and never repeat this work on their own.
	input.Metrics = analytics.EnrichMetrics(input.Metrics)
	result := e.derive(input.Profile, input.Year, input.ShareID, input.Metrics, input.ActionableState)

	recap := model.Recap{
		ID:              input.RecapID,
		ShareID:         input.ShareID,
		Profile:         input.Profile,
		Year:            input.Year,
		Period:          input.Period,
		RulesVersion:    e.rules.Version,
		RulesDigest:     e.digest,
		Metrics:         input.Metrics,
		ActionableState: input.ActionableState,
		Behavior:        result.Behavior,
		Achievements:    result.Achievements,
		Cards:           result.Cards,
		NextAction:      result.NextAction,
		GeneratedAt:     input.GeneratedAt,
	}
	if err := structural.ValidateRecap(recap); err != nil {
		return model.Recap{}, err
	}
	if err := e.validateAchievementSelection(recap.Achievements); err != nil {
		return model.Recap{}, fmt.Errorf("%w: achievement selection: %w", structural.ErrInvalidRecap, err)
	}
	return recap, nil
}

func (e *Engine) derive(
	profile model.Profile,
	year uint32,
	shareID uuid.UUID,
	metrics model.Metrics,
	state model.ActionableState,
) derived {
	detected := behavior.Detect(e.rules, metrics)
	achievements := achievement.Build(e.rules, metrics)
	action := nextaction.Build(e.rules, metrics, state, detected)
	shareCard := share.Build(e.rules.SharePolicy, shareID, year, metrics, detected, achievements)
	story := cards.Build(profile, year, metrics, detected, achievements, action, shareCard)
	return derived{
		Behavior: detected, Achievements: achievements, NextAction: action,
		ShareCard: shareCard, Cards: story,
	}
}

func (e *Engine) PublicProjection(value model.Recap) model.ShareCard {
	value = model.NormalizeRecap(value)
	return share.Build(
		e.rules.SharePolicy,
		value.ShareID,
		value.Year,
		value.Metrics,
		value.Behavior,
		value.Achievements,
	)
}
