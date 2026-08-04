package recap

import "testing"

func TestDetectBehavior(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected BehaviorCode
	}{
		{
			name: "active seller at boundary",
			metrics: Metrics{
				ListingsPublished: activeSellerMinListings,
				SalesCompleted:    activeSellerMinDeals,
			},
			expected: BehaviorActiveSeller,
		},
		{
			name: "active seller has priority over buyer patterns",
			metrics: Metrics{
				TotalViews:        200,
				ListingsPublished: activeSellerMinListings,
				SalesCompleted:    activeSellerMinDeals,
				FavoriteRate:      0.5,
				RepeatRate:        0.5,
				CategoriesCount:   8,
				ChatRate:          0,
			},
			expected: BehaviorActiveSeller,
		},
		{
			name: "find hunter at boundary",
			metrics: Metrics{
				TotalViews:   findHunterMinViews,
				FavoriteRate: findHunterMinFavorite,
				RepeatRate:   findHunterMinRepeat,
			},
			expected: BehaviorFindHunter,
		},
		{
			name: "researcher at lower boundaries",
			metrics: Metrics{
				TotalViews:      researcherMinViews,
				CategoriesCount: researcherMinCategories,
				ChatRate:        researcherMaxChatRate - 0.001,
			},
			expected: BehaviorResearcher,
		},
		{
			name: "researcher upper chat boundary is exclusive",
			metrics: Metrics{
				TotalViews:      researcherMinViews,
				CategoriesCount: researcherMinCategories,
				ChatRate:        researcherMaxChatRate,
			},
			expected: BehaviorUniversal,
		},
		{
			name: "find hunter more specific than researcher",
			metrics: Metrics{
				TotalViews:      150,
				CategoriesCount: 8,
				ChatRate:        0,
				FavoriteRate:    findHunterMinFavorite,
				RepeatRate:      findHunterMinRepeat,
			},
			expected: BehaviorFindHunter,
		},
		{
			name:     "fallback",
			metrics:  Metrics{},
			expected: BehaviorUniversal,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := DetectBehavior(test.metrics)
			if actual.Code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, actual.Code)
			}
			if actual.Title == "" || actual.Description == "" || actual.Reason == "" {
				t.Fatalf("behavior contains empty user-facing text: %+v", actual)
			}
		})
	}
}
