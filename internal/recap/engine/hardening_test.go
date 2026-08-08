package engine_test

import (
	"errors"
	"testing"
	"time"

	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/engine"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/recap/validation"
)

func testEngine(t testing.TB) *engine.Engine {
	t.Helper()
	core, err := engine.New(ruleset.DefaultRuleset())
	if err != nil {
		t.Fatalf("new engine: %v", err)
	}
	return core
}

func TestValidGeneratedRecapPassesIntegrityChecks(t *testing.T) {
	if _, err := testEngine(t).ValidateStored(testkit.Recap(), testkit.Clock()); err != nil {
		t.Fatalf("valid recap rejected: %v", err)
	}
}

func TestEngineDetectsForgedDerivedData(t *testing.T) {
	value := testkit.Recap()
	value.Behavior.Title = "Подменено"
	if _, err := testEngine(t).ValidateStored(value, testkit.Clock()); !errors.Is(err, validation.ErrInvalidRecap) {
		t.Fatalf("expected invalid recap error, got %v", err)
	}
}

func TestEngineRejectsPlausibleButSemanticallyForgedRecap(t *testing.T) {
	value := testkit.Recap()
	value.NextAction = model.NextAction{
		Code: model.ActionCreateListing, Title: "Продолжить продажи", Description: "Создай новое объявление.",
		ButtonText: "Создать объявление", Reason: "Формально валидная, но не вычисленная рекомендация.",
		Target: model.ActionTarget{Route: &model.RouteTarget{Route: "/listings/new"}},
	}
	if err := validation.ValidateRecap(value); err != nil {
		t.Fatalf("fixture must be structurally valid: %v", err)
	}
	if _, err := testEngine(t).ValidateStored(value, testkit.Clock()); !errors.Is(err, validation.ErrInvalidRecap) {
		t.Fatalf("forged recap error = %v", err)
	}
}

func TestEngineRejectsFutureDatedStoredRecap(t *testing.T) {
	value := testkit.Recap()
	value.GeneratedAt = testkit.Clock().Add(24 * time.Hour)
	value.ActionableState.CapturedAt = value.GeneratedAt
	if _, err := testEngine(t).ValidateStored(value, testkit.Clock()); !errors.Is(err, validation.ErrInvalidRecap) {
		t.Fatalf("future recap error = %v", err)
	}
}

func TestEngineRejectsStoredAchievementsFromSameCategory(t *testing.T) {
	value := testkit.Recap()
	configured := ruleset.DefaultRuleset()
	first := achievement.Build(configured, analytics.EnrichMetrics(model.Metrics{TotalEvents: 5, SalesCompleted: 5, MostActiveMonth: 1}))
	second := achievement.Build(configured, analytics.EnrichMetrics(model.Metrics{TotalEvents: 6, ListingsPublished: 5, SalesCompleted: 1, MostActiveMonth: 1}))
	if len(first) == 0 || len(second) == 0 {
		t.Fatal("seller achievement fixtures are empty")
	}
	value.Achievements = []model.Achievement{first[0], second[0]}
	if err := validation.ValidateRecap(value); err != nil {
		t.Fatalf("fixture must be structurally valid: %v", err)
	}
	if _, err := testEngine(t).ValidateStored(value, testkit.Clock()); !errors.Is(err, validation.ErrInvalidRecap) {
		t.Fatalf("same-category achievements error = %v", err)
	}
}

func TestEngineRejectsLowerGradeWhenHigherGradeWasEarned(t *testing.T) {
	value := testkit.Recap()
	value.Metrics.SalesCompleted = 5
	value.Metrics.ListingsPublished = 5
	value.Metrics.TotalEvents += 10
	value.Metrics = analytics.EnrichMetrics(value.Metrics)
	configured := ruleset.DefaultRuleset()
	expected := achievement.Build(configured, value.Metrics)
	lower := achievement.Build(configured, analytics.EnrichMetrics(model.Metrics{TotalEvents: 5, SalesCompleted: 1, MostActiveMonth: 1}))
	if len(expected) == 0 || len(lower) == 0 {
		t.Fatal("achievement fixtures are empty")
	}
	value.Achievements = append([]model.Achievement(nil), expected...)
	value.Achievements[0] = lower[0]
	if _, err := testEngine(t).ValidateStored(value, testkit.Clock()); !errors.Is(err, validation.ErrInvalidRecap) {
		t.Fatalf("lower-grade achievement error = %v", err)
	}
}

func TestEngineFreezesRulesetAtConstruction(t *testing.T) {
	configured := ruleset.DefaultRuleset()
	core, err := engine.New(configured)
	if err != nil {
		t.Fatal(err)
	}
	before := core.RecapKey(testkit.ProfileID, 2025)
	configured.Eligibility.MinEvents++
	configured.AchievementPolicy.Rules[0].Priority++
	configured.SharePolicy.AllowedAchievementCodes[0] = model.AchievementBookworm
	after := core.RecapKey(testkit.ProfileID, 2025)
	if before != after {
		t.Fatalf("engine rules changed after external mutation: before=%+v after=%+v", before, after)
	}
}

func TestBuildUsesDigestBoundEligibilityPolicy(t *testing.T) {
	configured := ruleset.DefaultRuleset()
	configured.Eligibility.MinEvents = testkit.Metrics().TotalEvents + 1
	core, err := engine.New(configured)
	if err != nil {
		t.Fatal(err)
	}
	_, err = core.Build(engine.BuildInput{
		RecapID: testkit.RecapID, ShareID: testkit.ShareID, Profile: testkit.Profile(), Year: 2025,
		Period: testkit.Period(), Metrics: testkit.Metrics(), ActionableState: testkit.ActionableState(), GeneratedAt: testkit.Clock(),
	})
	if !errors.Is(err, engine.ErrNotEnoughActivity) {
		t.Fatalf("eligibility error = %v", err)
	}
}
