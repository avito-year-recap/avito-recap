package recap

import "testing"

func TestBuildNextActionAllBranches(t *testing.T) {
	tests := []struct {
		name     string
		metrics  Metrics
		expected ActionCode
	}{
		{
			name:     "finish draft",
			metrics:  Metrics{ListingsCreated: 6, ListingsPublished: 2},
			expected: ActionFinishDraft,
		},
		{
			name:     "create listing for active seller",
			metrics:  Metrics{ListingsCreated: 5, ListingsPublished: 5, SalesCompleted: 3},
			expected: ActionCreateListing,
		},
		{
			name:     "view similar for decisive buyer",
			metrics:  Metrics{TotalViews: 60, ChatsStarted: 12, ChatsWithPurchase: 3, PurchasesCompleted: 3},
			expected: ActionViewSimilarListings,
		},
		{
			name:     "save search for researcher",
			metrics:  Metrics{TotalViews: 100, UniqueListings: 100, ChatsStarted: 4, CategoriesCount: 5},
			expected: ActionSaveSearch,
		},
		{
			name:     "open favorites for find hunter",
			metrics:  Metrics{TotalViews: 20, UniqueListings: 16, RepeatedViews: 4, FavoritesAdded: 3},
			expected: ActionOpenFavorites,
		},
		{
			name:     "continue dialogs for universal user",
			metrics:  Metrics{ChatsStarted: chatsForContinue},
			expected: ActionContinueDialogs,
		},
		{
			name:     "improve listings",
			metrics:  Metrics{ListingsCreated: publishedForImprove, ListingsPublished: publishedForImprove},
			expected: ActionImproveListings,
		},
		{
			name:     "open top category",
			metrics:  Metrics{TopCategoryCode: "electronics", TopCategory: "Электроника"},
			expected: ActionOpenTopCategory,
		},
		{name: "neutral explore fallback", metrics: Metrics{}, expected: ActionExploreRecommendations},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := BuildNextAction(test.metrics)
			if actual.Code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, actual.Code)
			}
			if actual.Title == "" || actual.Description == "" || actual.ButtonText == "" || actual.Reason == "" {
				t.Fatalf("next action contains empty user-facing text: %+v", actual)
			}
		})
	}
}

func TestBuildNextActionUsesBehaviorBeforeGenericFallbacks(t *testing.T) {
	metrics := Metrics{
		TotalViews:         60,
		UniqueListings:     48,
		RepeatedViews:      12,
		FavoritesAdded:     12,
		ChatsStarted:       12,
		ChatsWithPurchase:  3,
		PurchasesCompleted: 3,
		TopCategoryCode:    "furniture",
		TopCategory:        "Мебель и интерьер",
	}

	if actual := BuildNextAction(metrics); actual.Code != ActionViewSimilarListings {
		t.Fatalf("decisive buyer action must win over favorites, chats and category, got %s", actual.Code)
	}
}

func TestBuildNextActionNormalizesTopCategory(t *testing.T) {
	action := BuildNextAction(Metrics{
		TopCategoryCode: "  electronics ",
		TopCategory:     "  Электроника\n",
	})
	if action.Code != ActionOpenTopCategory {
		t.Fatalf("expected top-category action, got %s", action.Code)
	}
	if action.Description != "Вернись в категорию «Электроника» и проверь новые варианты." {
		t.Fatalf("category was not normalized in action text: %q", action.Description)
	}
}
