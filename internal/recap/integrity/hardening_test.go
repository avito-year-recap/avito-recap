package integrity_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/integrity"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
	"github.com/year-recap/internal/recap/presentation/cards"
	"github.com/year-recap/internal/recap/presentation/share"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/recap/validation/structural"
)

// buildCards mirrors the engine's own derive step so hardening tests can
// recompute cards for a forged recap without duplicating engine internals.
func buildCards(
	configured ruleset.Ruleset,
	profile model.Profile,
	year uint32,
	shareID uuid.UUID,
	metrics model.Metrics,
	detected model.Behavior,
	achievements []model.Achievement,
	action model.NextAction,
) []model.Card {
	shareCard := share.Build(configured.SharePolicy, shareID, year, metrics, detected, achievements)
	return cards.Build(profile, year, metrics, detected, achievements, action, shareCard)
}

func TestValidGeneratedRecapPassesIntegrityChecks(t *testing.T) {
	if err := integrity.ValidateRecapAgainstRuleset(testkit.Recap(), ruleset.DefaultRuleset(), testkit.Clock()); err != nil {
		t.Fatalf("valid recap rejected: %v", err)
	}
}

func TestRecapIntegrityDetectsForgedDerivedData(t *testing.T) {
	value := testkit.Recap()
	value.Behavior.Title = "Подменено"
	if err := integrity.ValidateRecapAgainstRuleset(value, ruleset.DefaultRuleset(), testkit.Clock()); !errors.Is(err, structural.ErrInvalidRecap) {
		t.Fatalf("expected invalid recap error, got %v", err)
	}
}

func TestIntegrityRejectsPlausibleButSemanticallyForgedRecap(t *testing.T) {
	value := testkit.Recap()
	value.NextAction = nextaction.CreateListingAction("Формально валидная, но не вычисленная рекомендация.")
	value.Cards = buildCards(ruleset.DefaultRuleset(), value.Profile, value.Year, value.ShareID, value.Metrics, value.Behavior, value.Achievements, value.NextAction)
	if err := structural.ValidateRecap(value); err != nil {
		t.Fatalf("fixture must be structurally valid: %v", err)
	}
	if err := integrity.ValidateRecapAgainstRuleset(value, ruleset.DefaultRuleset(), testkit.Clock()); !errors.Is(err, structural.ErrInvalidRecap) {
		t.Fatalf("forged recap error = %v", err)
	}
}

func TestIntegrityRejectsFutureDatedStoredRecap(t *testing.T) {
	value := testkit.Recap()
	value.GeneratedAt = testkit.Clock().Add(24 * time.Hour)
	value.ActionableState.CapturedAt = value.GeneratedAt
	if err := integrity.ValidateRecapAgainstRuleset(value, ruleset.DefaultRuleset(), testkit.Clock()); !errors.Is(err, structural.ErrInvalidRecap) {
		t.Fatalf("future recap error = %v", err)
	}
}

func TestIntegrityRejectsStoredAchievementsFromSameCategory(t *testing.T) {
	value := testkit.Recap()
	first := achievement.Build(ruleset.DefaultRuleset(), model.Metrics{SalesCompleted: 5})
	second := achievement.Build(ruleset.DefaultRuleset(), model.Metrics{ListingsPublished: 5, SalesCompleted: 1})
	if len(first) == 0 || len(second) == 0 {
		t.Fatal("seller achievement fixtures are empty")
	}
	value.Achievements = []model.Achievement{first[0], second[0]}
	value.Cards = buildCards(ruleset.DefaultRuleset(), value.Profile, value.Year, value.ShareID, value.Metrics, value.Behavior, value.Achievements, value.NextAction)
	if err := structural.ValidateRecap(value); err != nil {
		t.Fatalf("fixture must be structurally valid: %v", err)
	}
	if err := integrity.ValidateRecapAgainstRuleset(value, ruleset.DefaultRuleset(), testkit.Clock()); !errors.Is(err, structural.ErrInvalidRecap) {
		t.Fatalf("same-category achievements error = %v", err)
	}
}

func TestIntegrityRejectsLowerGradeWhenHigherGradeWasEarned(t *testing.T) {
	value := testkit.Recap()
	value.Metrics.SalesCompleted = 5
	value.Metrics.ListingsPublished = 5
	value.Metrics.TotalEvents += 10
	value.Metrics = analytics.EnrichMetrics(value.Metrics)
	value.Behavior = behavior.Detect(ruleset.DefaultRuleset(), value.Metrics)
	expected := achievement.Build(ruleset.DefaultRuleset(), value.Metrics)
	lower := achievement.Build(ruleset.DefaultRuleset(), model.Metrics{SalesCompleted: 1})
	if len(expected) == 0 || len(lower) == 0 {
		t.Fatal("achievement fixtures are empty")
	}
	value.Achievements = append([]model.Achievement(nil), expected...)
	value.Achievements[0] = lower[0]
	value.NextAction = nextaction.Build(ruleset.DefaultRuleset(), value.Metrics, value.ActionableState, value.Behavior)
	value.Cards = buildCards(ruleset.DefaultRuleset(), value.Profile, value.Year, value.ShareID, value.Metrics, value.Behavior, value.Achievements, value.NextAction)
	if err := structural.ValidateRecap(value); err != nil {
		t.Fatalf("fixture must be structurally valid: %v", err)
	}
	if err := integrity.ValidateRecapAgainstRuleset(value, ruleset.DefaultRuleset(), testkit.Clock()); !errors.Is(err, structural.ErrInvalidRecap) {
		t.Fatalf("lower-grade achievement error = %v", err)
	}
}
