package recap

import "testing"

func TestDetectBehavior(t *testing.T) {
	thresholds := DefaultRuleset().Thresholds
	tests := []struct {
		name     string
		metrics  Metrics
		expected BehaviorCode
	}{
		{
			name: "active seller at boundary",
			metrics: Metrics{
				ListingsPublished: thresholds.ActiveSellerMinListings,
				SalesCompleted:    thresholds.ActiveSellerMinDeals,
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
				ListingsCreated:   thresholds.StartingSellerMinCreated,
				ListingsPublished: thresholds.StartingSellerMaxPublished,
			},
			expected: BehaviorStartingSeller,
		},
		{
			name: "decisive buyer uses linked chats",
			metrics: Metrics{
				ChatsStarted:       thresholds.DecisiveBuyerMinChats,
				ChatsWithPurchase:  thresholds.DecisiveBuyerMinLinkedChats,
				PurchasesCompleted: thresholds.DecisiveBuyerMinPurchases,
			},
			expected: BehaviorDecisiveBuyer,
		},
		{
			name: "purchases without linked chats are not decisive buyer",
			metrics: Metrics{
				ChatsStarted:       thresholds.DecisiveBuyerMinChats,
				PurchasesCompleted: thresholds.DecisiveBuyerMinPurchases,
			},
			expected: BehaviorUniversal,
		},
		{
			name: "find hunter at boundary",
			metrics: Metrics{
				TotalViews:     thresholds.FindHunterMinViews,
				UniqueListings: 16,
				RepeatedViews:  4,
				FavoritesAdded: thresholds.FindHunterMinFavorites,
			},
			expected: BehaviorFindHunter,
		},
		{
			name: "find hunter does not divide favorites by annual views",
			metrics: Metrics{
				TotalViews:     200,
				UniqueListings: 160,
				RepeatedViews:  40,
				FavoritesAdded: thresholds.FindHunterMinFavorites,
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
				TotalViews:      thresholds.ResearcherMinViews,
				UniqueListings:  thresholds.ResearcherMinViews,
				ChatsStarted:    thresholds.ResearcherMaxChats,
				CategoriesCount: thresholds.ResearcherMinCategories,
			},
			expected: BehaviorResearcher,
		},
		{
			name: "researcher chat boundary is absolute",
			metrics: Metrics{
				TotalViews:      1000,
				UniqueListings:  1000,
				ChatsStarted:    thresholds.ResearcherMaxChats + 1,
				CategoriesCount: thresholds.ResearcherMinCategories,
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

func TestBehaviorSelectionUsesScoreNotSliceOrder(t *testing.T) {
	low := behaviorCandidate{behavior: Behavior{Code: BehaviorResearcher, Score: 80}, eligible: true, tieBreak: 10}
	high := behaviorCandidate{behavior: Behavior{Code: BehaviorActiveSeller, Score: 120}, eligible: true, tieBreak: 50}
	first := selectBehaviorCandidate([]behaviorCandidate{low, high})
	second := selectBehaviorCandidate([]behaviorCandidate{high, low})
	if first.Code != BehaviorActiveSeller || second.Code != BehaviorActiveSeller {
		t.Fatalf("selection depends on order: %s / %s", first.Code, second.Code)
	}
}

func TestDetectedBehaviorContainsAuditableEvidence(t *testing.T) {
	behavior := DetectBehavior(Metrics{ListingsPublished: 5, SalesCompleted: 3})
	if behavior.Score == 0 || len(behavior.Evidence) == 0 {
		t.Fatalf("missing score/evidence: %+v", behavior)
	}
	var sum uint32
	for _, evidence := range behavior.Evidence {
		sum += evidence.Points
	}
	if sum != behavior.Score {
		t.Fatalf("evidence sum %d != score %d", sum, behavior.Score)
	}
}
