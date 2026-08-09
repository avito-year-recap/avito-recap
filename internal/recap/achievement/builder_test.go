package achievement

import (
	"testing"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestBuildBalancedPortfolio(t *testing.T) {
	values := buildAchievements(model.Metrics{ListingsPublished: 5, SalesCompleted: 5, PurchasesCompleted: 5, ChatsStarted: 5, ChatsWithPurchase: 3})
	if len(values) != 3 {
		t.Fatalf("got %d: %+v", len(values), values)
	}
	seen := map[model.AchievementCode]bool{}
	for _, value := range values {
		seen[value.Code] = true
	}
	if !seen[model.AchievementAllRounder] || !seen[model.AchievementSuccessfulSeller] || !seen[model.AchievementQuickDecision] {
		t.Fatalf("unexpected portfolio: %+v", values)
	}
}

func TestBuildThematicAchievement(t *testing.T) {
	values := buildAchievements(model.Metrics{TotalViews: 30, CategoriesCount: 1, CategoryActivities: []model.CategoryActivity{{CategoryCode: analytics.CategoryBooks, Category: "Книги", Views: 30}}})
	if len(values) != 1 || values[0].Code != model.AchievementBookworm {
		t.Fatalf("unexpected: %+v", values)
	}
}

func TestThematicSignalDoesNotOverflow(t *testing.T) {
	metrics := model.Metrics{
		FavoritesAdded:  ^uint64(0),
		CategoriesCount: 2,
		CategoryActivities: []model.CategoryActivity{
			{
				CategoryCode:   analytics.CategoryWomensFashion,
				Category:       "Женская одежда и аксессуары",
				FavoritesAdded: ^uint64(0)/4 + 1,
			},
			{
				CategoryCode: analytics.CategoryBooks,
				Category:     "Книги",
				Views:        100,
			},
		},
	}

	values := buildAchievements(metrics)
	for _, value := range values {
		if value.Code == model.AchievementStyleIcon {
			return
		}
	}
	t.Fatalf("style achievement was lost after uint64 overflow: %+v", values)
}

func buildAchievements(metrics model.Metrics) []model.Achievement {
	return Build(ruleset.DefaultRuleset(), analytics.EnrichMetrics(metrics))
}
