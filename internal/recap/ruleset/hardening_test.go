package ruleset

import (
	"errors"
	"testing"

	"github.com/year-recap/internal/recap/model"
)

func TestRulesDigestBindsMaterialConfiguration(t *testing.T) {
	base := DefaultRuleset()
	baseDigest := base.Digest()
	mutations := []func(*Ruleset){
		func(r *Ruleset) { r.Algorithm += "-changed" },
		func(r *Ruleset) { r.Thresholds.FindHunterMinViews++ },
		func(r *Ruleset) { r.AchievementPolicy.Rules[0].Priority++ },
		func(r *Ruleset) { r.AchievementPolicy.Rules[0].Category = model.AchievementCategoryBuying },
		func(r *Ruleset) { r.AchievementPolicy.MaxAwarded-- },
		func(r *Ruleset) { r.RecommendationThresholds.ImproveListingsMinActive++ },
		func(r *Ruleset) { r.RecommendationPriorities.OpenFavorites++ },
		func(r *Ruleset) { r.SharePolicy.MaxPublicTextRunes++ },
	}
	for index, mutate := range mutations {
		value := base
		value.AchievementPolicy.Rules = append([]AchievementRuleConfig(nil), base.AchievementPolicy.Rules...)
		value.SharePolicy.AllowedAchievementCodes = append([]model.AchievementCode(nil), base.SharePolicy.AllowedAchievementCodes...)
		mutate(&value)
		if value.Digest() == baseDigest {
			t.Fatalf("mutation %d did not change digest", index)
		}
	}
}

func TestRulesetRejectsLabelsWithoutImplementedContract(t *testing.T) {
	invalidAlgorithm := DefaultRuleset()
	invalidAlgorithm.Algorithm = "unimplemented"
	if err := invalidAlgorithm.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("unknown algorithm error = %v", err)
	}
	invalidThresholds := DefaultRuleset()
	invalidThresholds.Thresholds.StartingSellerMaxPublished = invalidThresholds.Thresholds.StartingSellerMinCreated
	if err := invalidThresholds.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("impossible thresholds error = %v", err)
	}
	invalidPriority := DefaultRuleset()
	invalidPriority.RecommendationPriorities.OpenTopCategory = invalidPriority.RecommendationPriorities.FinishDraft + 1
	if err := invalidPriority.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("unsafe priority error = %v", err)
	}
	invalidAchievementLimit := DefaultRuleset()
	invalidAchievementLimit.AchievementPolicy.MaxAwarded = MaxAchievements + 1
	if err := invalidAchievementLimit.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("achievement limit error = %v", err)
	}
	duplicateRule := DefaultRuleset()
	duplicateRule.AchievementPolicy.Rules[1].Code = duplicateRule.AchievementPolicy.Rules[0].Code
	if err := duplicateRule.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("duplicate achievement policy error = %v", err)
	}
	unknownCategory := DefaultRuleset()
	unknownCategory.AchievementPolicy.Rules[0].Category = "UNKNOWN"
	if err := unknownCategory.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("unknown achievement category error = %v", err)
	}
}
