package achievement

import (
	"testing"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
)

func TestBuildBalancedPortfolio(t *testing.T) {
	values := Build(model.Metrics{ListingsPublished: 5, SalesCompleted: 5, PurchasesCompleted: 5, ChatsStarted: 5, ChatsWithPurchase: 3})
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
	values := Build(model.Metrics{TotalViews: 30, CategoriesCount: 1, CategoryActivities: []model.CategoryActivity{{CategoryCode: analytics.CategoryBooks, Category: "Книги", Views: 30}}})
	if len(values) != 1 || values[0].Code != model.AchievementBookworm {
		t.Fatalf("unexpected: %+v", values)
	}
}
