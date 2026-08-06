package achievement

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func selectAchievementPortfolio(metrics model.Metrics, thresholds ruleset.AchievementThresholds, candidates []model.Achievement, limit int) []model.Achievement {
	if limit == 0 || len(candidates) == 0 {
		return nil
	}

	if isBalancedSellerBuyer(metrics, thresholds) {
		selected := make([]model.Achievement, 0, limit)
		selected = appendFirstCode(selected, candidates, model.AchievementAllRounder, limit)
		selected = appendBestCategory(selected, candidates, model.AchievementCategorySelling, limit)
		selected = appendBestCategory(selected, candidates, model.AchievementCategoryBuying, limit)
		selected = fillFromBestCategories(selected, candidates, limit)
		sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
		return selected
	}

	// A profile that only sold should get one clear seller identity rather than
	// an overloaded list of unrelated awards.
	if metrics.SalesCompleted > 0 && metrics.PurchasesCompleted == 0 {
		selected := appendBestCategory(nil, candidates, model.AchievementCategorySelling, 1)
		return selected
	}

	// Buyer-only profiles may receive up to three distinct category personas.
	// This makes the recap feel personal without forcing a seller identity onto
	// someone who only used buying scenarios.
	if metrics.PurchasesCompleted > 0 && metrics.SalesCompleted == 0 {
		selected := make([]model.Achievement, 0, limit)
		for _, candidate := range candidates {
			if len(selected) >= limit {
				break
			}
			if candidate.Category == model.AchievementCategoryInterest {
				selected = append(selected, candidate)
			}
		}
		selected = fillFromBestCategories(selected, candidates, limit)
		sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
		return selected
	}

	if metrics.SalesCompleted > metrics.PurchasesCompleted {
		selected := appendBestCategory(nil, candidates, model.AchievementCategorySelling, limit)
		// For seller-dominant mixed profiles, additional earned selling awards are
		// preferred before unrelated categories.
		for _, candidate := range candidates {
			if len(selected) >= limit {
				break
			}
			if candidate.Category == model.AchievementCategorySelling && !containsAchievement(selected, candidate.Code) {
				selected = append(selected, candidate)
			}
		}
		selected = fillFromBestCategories(selected, candidates, limit)
		sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
		return selected
	}

	selected := fillFromBestCategories(nil, candidates, limit)
	sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
	return selected
}

func isBalancedSellerBuyer(metrics model.Metrics, thresholds ruleset.AchievementThresholds) bool {
	if metrics.PurchasesCompleted < thresholds.BalancedMinPurchases ||
		metrics.SalesCompleted < thresholds.BalancedMinSales {
		return false
	}
	maximum := metrics.PurchasesCompleted
	minimum := metrics.SalesCompleted
	if maximum < minimum {
		maximum, minimum = minimum, maximum
	}
	if maximum == 0 {
		return false
	}
	return float64(maximum-minimum)/float64(maximum) <= thresholds.BalancedMaxDifferenceRate
}

func appendFirstCode(selected, candidates []model.Achievement, code model.AchievementCode, limit int) []model.Achievement {
	if len(selected) >= limit {
		return selected
	}
	for _, candidate := range candidates {
		if candidate.Code == code && !containsAchievement(selected, code) {
			return append(selected, candidate)
		}
	}
	return selected
}

func appendBestCategory(selected, candidates []model.Achievement, category model.AchievementCategory, limit int) []model.Achievement {
	if len(selected) >= limit {
		return selected
	}
	for _, candidate := range candidates {
		if candidate.Category == category && !containsAchievement(selected, candidate.Code) {
			return append(selected, candidate)
		}
	}
	return selected
}

func fillFromBestCategories(selected, candidates []model.Achievement, limit int) []model.Achievement {
	seenCategories := make(map[model.AchievementCategory]struct{}, len(selected))
	for _, achievement := range selected {
		seenCategories[achievement.Category] = struct{}{}
	}
	for _, candidate := range candidates {
		if len(selected) >= limit {
			break
		}
		if containsAchievement(selected, candidate.Code) {
			continue
		}
		if _, exists := seenCategories[candidate.Category]; exists {
			continue
		}
		selected = append(selected, candidate)
		seenCategories[candidate.Category] = struct{}{}
	}
	return selected
}

func containsAchievement(values []model.Achievement, code model.AchievementCode) bool {
	for _, value := range values {
		if value.Code == code {
			return true
		}
	}
	return false
}

// achievementLess defines a deterministic total ordering: higher product
// priority wins, then stronger measured evidence, then stable code.
func achievementLess(left, right model.Achievement) bool {
	if left.Priority != right.Priority {
		return left.Priority > right.Priority
	}
	if left.Strength != right.Strength {
		return left.Strength > right.Strength
	}
	return left.Code < right.Code
}
