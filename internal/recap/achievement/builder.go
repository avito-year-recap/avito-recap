package achievement

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

type achievementDefinition struct {
	evaluate func(model.Metrics, ruleset.Ruleset) (model.Achievement, bool)
}

// Build evaluates the complete catalogue and then assembles a
// three-slot portfolio. Balanced seller-buyers receive the versatility persona
// plus one selling and one buying award. Seller-only profiles receive exactly
// one strongest seller persona. Other profiles receive the strongest award in
// each category, up to the global limit.
func Build(r ruleset.Ruleset, metrics model.Metrics) []model.Achievement {
	candidates := make([]model.Achievement, 0, len(r.AchievementPolicy.Rules))

	for _, configured := range r.AchievementPolicy.Rules {
		definition, ok := achievementDefinitionFor(configured.Code)
		if !ok {
			continue
		}
		candidate, earned := definition.evaluate(metrics, r)
		if !earned {
			continue
		}
		candidate.Category = configured.Category
		candidate.Priority = configured.Priority
		candidates = append(candidates, candidate)
	}

	sort.Slice(candidates, func(i, j int) bool { return achievementLess(candidates[i], candidates[j]) })
	limit := r.AchievementPolicy.MaxAwarded
	if limit > ruleset.MaxAchievements {
		limit = ruleset.MaxAchievements
	}
	if limit < 0 {
		limit = 0
	}
	return selectAchievementPortfolio(metrics, r.AchievementThresholds, candidates, limit)
}
