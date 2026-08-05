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
			name: "active seller has highest priority",
			metrics: Metrics{
				TotalViews:         200,
				UniqueListings:     100,
				RepeatedViews:      100,
				FavoritesAdded:     50,
				ChatsStarted:       20,
				ChatsWithPurchase:  4,
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
				ListingsPublished: startingSellerMaxPublished,
			},
			expected: BehaviorStartingSeller,
		},
		{
			name: "decisive buyer uses linked chats",
			metrics: Metrics{
				ChatsStarted:       decisiveBuyerMinChats,
				ChatsWithPurchase:  decisiveBuyerMinLinkedChats,
				PurchasesCompleted: decisiveBuyerMinPurchases,
			},
			expected: BehaviorDecisiveBuyer,
		},
		{
			name: "purchases without linked chats are not decisive buyer",
			metrics: Metrics{
				ChatsStarted:       decisiveBuyerMinChats,
				PurchasesCompleted: decisiveBuyerMinPurchases,
			},
			expected: BehaviorUniversal,
		},
		{
			name: "find hunter at boundary",
			metrics: Metrics{
				TotalViews:     findHunterMinViews,
				UniqueListings: 16,
				RepeatedViews:  4,
				FavoritesAdded: findHunterMinFavorites,
			},
			expected: BehaviorFindHunter,
		},
		{
			name: "find hunter does not divide favorites by annual views",
			metrics: Metrics{
				TotalViews:     200,
				UniqueListings: 160,
				RepeatedViews:  40,
				FavoritesAdded: findHunterMinFavorites,
			},
			expected: BehaviorFindHunter,
		},
		{
			name: "favorite from another period is not treated as a view conversion",
			metrics: Metrics{
				TotalViews:     1,
				UniqueListings: 1,
				FavoritesAdded: 10,
			},
			expected: BehaviorUniversal,
		},
		{
			name: "researcher at lower boundaries",
			metrics: Metrics{
				TotalViews:      researcherMinViews,
				UniqueListings:  researcherMinViews,
				ChatsStarted:    researcherMaxChats,
				CategoriesCount: researcherMinCategories,
			},
			expected: BehaviorResearcher,
		},
		{
			name: "researcher chat boundary is absolute",
			metrics: Metrics{
				TotalViews:      1000,
				UniqueListings:  1000,
				ChatsStarted:    researcherMaxChats + 1,
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
		{
			name: "publications from an earlier creation cohort do not imply drafts",
			metrics: Metrics{
				ListingsCreated:   1,
				ListingsPublished: 5,
			},
			expected: BehaviorUniversal,
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
