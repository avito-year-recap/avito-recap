package nextaction

import (
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

func TestEveryActionCodeIsReachable(t *testing.T) {
	listingID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	dialogID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	purchaseID := uuid.MustParse("55555555-5555-4555-8555-555555555555")

	tests := []struct {
		name    string
		metrics model.Metrics
		state   model.ActionableState
		want    model.ActionCode
	}{
		{
			name:  "finish draft",
			state: model.ActionableState{CurrentDrafts: 1, DraftListingID: listingID},
			want:  model.ActionFinishDraft,
		},
		{
			name:  "continue dialog",
			state: model.ActionableState{OpenDialogs: 1, OpenDialogID: dialogID},
			want:  model.ActionContinueDialogs,
		},
		{
			name:  "improve listing",
			state: model.ActionableState{ActiveListings: 3, ActiveListingID: listingID, HasEverPublishedListing: true},
			want:  model.ActionImproveListings,
		},
		{
			name: "similar to purchase",
			metrics: model.Metrics{
				PurchasesCompleted: 3, ChatsStarted: 5, ChatsWithPurchase: 3,
			},
			state: model.ActionableState{LastPurchasedListingID: purchaseID},
			want:  model.ActionViewSimilarListings,
		},
		{
			name: "save search",
			metrics: model.Metrics{
				TotalViews: 100, UniqueListings: 100, CategoriesCount: 5,
				TopCategoryCode: "electronics", TopCategory: "Электроника", TopCategoryViews: 40,
			},
			state: model.ActionableState{HasSavedSearchForTopCategory: false},
			want:  model.ActionSaveSearch,
		},
		{
			name:  "open favorites",
			state: model.ActionableState{FavoritesCount: 3},
			want:  model.ActionOpenFavorites,
		},
		{
			name: "create first listing",
			metrics: model.Metrics{
				ListingsCreated: 3, ListingsPublished: 0, SalesCompleted: 0,
			},
			state: model.ActionableState{HasEverPublishedListing: false},
			want:  model.ActionCreateFirstListing,
		},
		{
			name: "create listing",
			metrics: model.Metrics{
				ListingsPublished: 5, SalesCompleted: 3,
			},
			state: model.ActionableState{HasEverPublishedListing: true},
			want:  model.ActionCreateListing,
		},
		{
			name: "open top category",
			metrics: model.Metrics{
				TopCategoryCode: "auto", TopCategory: "Авто", TopCategoryViews: 1,
			},
			want: model.ActionOpenTopCategory,
		},
		{
			name: "recommendations fallback",
			want: model.ActionExploreRecommendations,
		},
	}

	seen := make(map[model.ActionCode]bool, len(tests))
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Build(tc.metrics, tc.state)
			if got.Code != tc.want {
				t.Fatalf("action = %s, want %s; reason=%s", got.Code, tc.want, got.Reason)
			}
			if got.Title == "" || got.Description == "" || got.ButtonText == "" || got.Reason == "" {
				t.Fatalf("incomplete action: %+v", got)
			}
			seen[got.Code] = true
		})
	}

	for _, code := range []model.ActionCode{
		model.ActionFinishDraft,
		model.ActionContinueDialogs,
		model.ActionImproveListings,
		model.ActionViewSimilarListings,
		model.ActionSaveSearch,
		model.ActionOpenFavorites,
		model.ActionCreateFirstListing,
		model.ActionCreateListing,
		model.ActionOpenTopCategory,
		model.ActionExploreRecommendations,
	} {
		if !seen[code] {
			t.Errorf("action %s has no reachability witness", code)
		}
	}
}

func TestHigherPriorityExecutableWorkWins(t *testing.T) {
	listingID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	dialogID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	got := Build(
		model.Metrics{TopCategoryCode: "auto", TopCategory: "Авто", TopCategoryViews: 1},
		model.ActionableState{
			CurrentDrafts: 1, DraftListingID: listingID,
			OpenDialogs: 1, OpenDialogID: dialogID,
			FavoritesCount: 10,
		},
	)
	if got.Code != model.ActionFinishDraft {
		t.Fatalf("action = %s, want %s", got.Code, model.ActionFinishDraft)
	}
}
