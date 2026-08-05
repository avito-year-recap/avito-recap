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
				ListingsCreated:   activeSellerMinListings,
				ListingsPublished: activeSellerMinListings,
				SalesCompleted:    activeSellerMinDeals,
			},
			expected: BehaviorActiveSeller,
		},
		{
			name: "active seller has highest priority",
			metrics: Metrics{
				TotalViews:         200,
				UniqueListings:     100,
				RepeatedViews:      100,
				FavoritesAdded:     50,
				ChatsStarted:       20,
				PurchasesCompleted: 4,
				ListingsCreated:    5,
				ListingsPublished:  5,
				SalesCompleted:     3,
				CategoriesCount:    8,
			},
			expected: BehaviorActiveSeller,
		},
		{
			name: "starting seller at boundary",
			metrics: Metrics{
				ListingsCreated:   startingSellerMinCreated,
				ListingsPublished: 1,
			},
			expected: BehaviorStartingSeller,
		},
		{
			name: "decisive buyer at boundary",
			metrics: Metrics{
				TotalViews:         20,
				UniqueListings:     20,
				ChatsStarted:       decisiveBuyerMinChats,
				PurchasesCompleted: decisiveBuyerMinPurchases,
			},
			expected: BehaviorDecisiveBuyer,
		},
		{
			name: "find hunter at boundary",
			metrics: Metrics{
				TotalViews:     findHunterMinViews,
				UniqueListings: 16,
				RepeatedViews:  4,

			},
			expected: BehaviorFindHunter,
		},
		{
			name: "researcher at lower boundaries",
			metrics: Metrics{
				TotalViews:      researcherMinViews,
				UniqueListings:  researcherMinViews,
				ChatsStarted:    4,
				CategoriesCount: researcherMinCategories,
			},
			expected: BehaviorResearcher,
		},
		{
			name: "researcher upper chat boundary is exclusive",
			metrics: Metrics{
				TotalViews:      researcherMinViews,
				UniqueListings:  researcherMinViews,
				ChatsStarted:    5,
				CategoriesCount: researcherMinCategories,
			},
			expected: BehaviorUniversal,
		},
		{
			name: "find hunter more specific than researcher",
			metrics: Metrics{
				TotalViews:      100,
				UniqueListings:  80,
				RepeatedViews:   20,
				FavoritesAdded:  15,
				CategoriesCount: 6,
			},
			expected: BehaviorFindHunter,
		},
		{name: "fallback", metrics: Metrics{}, expected: BehaviorUniversal},
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
