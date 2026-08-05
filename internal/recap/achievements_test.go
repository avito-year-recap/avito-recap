package recap

import "testing"

func TestBuildAchievementsForMVPProfiles(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected []AchievementCode
	}{
		{
			name:     "active buyer",
			metrics:  Metrics{TotalViews: 180, FavoritesAdded: 28},
			expected: []AchievementCode{AchievementAttentiveResearcher, AchievementMasterOfFavorites},
		},
		{
			name:     "active seller",
			metrics:  Metrics{ListingsPublished: 9, SalesCompleted: 6},
			expected: []AchievementCode{AchievementSuccessfulSeller, AchievementConsistentPublisher},
		},
		{
			name:     "researcher",
			metrics:  Metrics{TotalViews: 260, CategoriesCount: 7},
			expected: []AchievementCode{AchievementBroadInterests, AchievementAttentiveResearcher},
		},
		{
			name:     "universal",
			metrics:  Metrics{PurchasesCompleted: 1, SalesCompleted: 2, ListingsPublished: 4, ChatsStarted: 9},
			expected: []AchievementCode{AchievementAllRounder},
		},
		{
			name:     "draft seller",
			metrics:  Metrics{ListingsCreated: 7, ListingsPublished: 2},
			expected: []AchievementCode{AchievementFirstSellingSteps},
		},
		{
			name:     "decisive buyer",
			metrics:  Metrics{ChatsStarted: 15, PurchasesCompleted: 4},
			expected: []AchievementCode{AchievementDealCloser, AchievementQuickDecision},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := BuildAchievements(test.metrics)
			assertAchievementCodes(t, actual, test.expected)
			for _, achievement := range actual {
				if achievement.Priority == 0 || achievement.Title == "" || achievement.Reason == "" {
					t.Fatalf("incomplete achievement: %+v", achievement)
				}
			}
		})
	}
}

func TestBuildAchievementsLimitsSortsAndHasUniqueCodes(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews:         300,
		FavoritesAdded:     40,
		ChatsStarted:       20,
		PurchasesCompleted: 5,
		ListingsCreated:    10,
		ListingsPublished:  10,
		SalesCompleted:     7,
		CategoriesCount:    8,
	})

	if len(result) != maxAchievements {
		t.Fatalf("expected %d achievements, got %d", maxAchievements, len(result))
	}
	seen := make(map[AchievementCode]struct{}, len(result))
	for index, achievement := range result {
		if _, exists := seen[achievement.Code]; exists {
			t.Fatalf("duplicate achievement code: %s", achievement.Code)
		}
		seen[achievement.Code] = struct{}{}
		if index > 0 && result[index-1].Priority <= achievement.Priority {
			t.Fatalf("achievements are not strictly ordered by priority: %+v", result)
		}
	}
}

func TestBuildAchievementsBelowThresholds(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews:        149,
		FavoritesAdded:    19,
		ListingsCreated:   2,
		ListingsPublished: 2,
		CategoriesCount:   5,
	})
	if len(result) != 0 {
		t.Fatalf("expected no achievements, got %+v", result)
	}
}

func assertAchievementCodes(t *testing.T, actual []Achievement, expected []AchievementCode) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("achievement count = %d, want %d: %+v", len(actual), len(expected), actual)
	}
	for index, code := range expected {
		if actual[index].Code != code {
			t.Fatalf("achievement %d = %s, want %s", index, actual[index].Code, code)
		}
	}
}
