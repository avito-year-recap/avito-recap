package recap

import "testing"

func TestBuildAchievementsEachRuleBoundary(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected AchievementCode
	}{
		{name: "successful seller", metrics: Metrics{SalesCompleted: 5}, expected: AchievementSuccessfulSeller},
		{name: "active publisher", metrics: Metrics{ListingsPublished: 5}, expected: AchievementActivePublisher},
		{name: "attentive researcher", metrics: Metrics{TotalViews: 150}, expected: AchievementAttentiveResearcher},
		{name: "favorites curator", metrics: Metrics{FavoritesAdded: 20}, expected: AchievementFavoritesCurator},
		{name: "category explorer", metrics: Metrics{CategoriesCount: 6}, expected: AchievementCategoryExplorer},
		{name: "consistent user", metrics: Metrics{ActiveDays: 30}, expected: AchievementConsistentUser},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := BuildAchievements(test.metrics)
			if len(result) != 1 {
				t.Fatalf("expected one achievement, got %d: %+v", len(result), result)
			}
			if result[0].Code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, result[0].Code)
			}
			if result[0].Priority == 0 || result[0].Title == "" || result[0].Reason == "" {
				t.Fatalf("incomplete achievement: %+v", result[0])
			}
		})
	}
}

func TestBuildAchievementsBelowThresholds(t *testing.T) {
	result := BuildAchievements(Metrics{
		SalesCompleted:    4,
		ListingsPublished: 4,
		TotalViews:        149,
		FavoritesAdded:    19,
		CategoriesCount:   5,
		ActiveDays:        29,
	})
	if len(result) != 0 {
		t.Fatalf("expected no achievements, got %+v", result)
	}
}

func TestBuildAchievementsLimitsSortsAndHasUniqueCodes(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews:        300,
		FavoritesAdded:    40,
		ListingsPublished: 10,
		SalesCompleted:    7,
		CategoriesCount:   8,
		ActiveDays:        120,
	})

	if len(result) != maxAchievements {
		t.Fatalf("expected %d achievements, got %d", maxAchievements, len(result))
	}

	expected := []AchievementCode{
		AchievementSuccessfulSeller,
		AchievementActivePublisher,
		AchievementAttentiveResearcher,
	}
	seen := make(map[AchievementCode]struct{}, len(result))

	for index, code := range expected {
		if result[index].Code != code {
			t.Fatalf("achievement %d: expected %s, got %s", index, code, result[index].Code)
		}
		if _, exists := seen[result[index].Code]; exists {
			t.Fatalf("duplicate achievement code: %s", result[index].Code)
		}
		seen[result[index].Code] = struct{}{}
		if index > 0 && result[index-1].Priority <= result[index].Priority {
			t.Fatalf("achievements are not strictly ordered by priority: %+v", result)
		}
	}
}
