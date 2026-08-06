package nextaction

import (
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

func TestBuildUsesOnlyThreeApprovedTitles(t *testing.T) {
	tests := []struct {
		name    string
		metrics model.Metrics
		state   model.ActionableState
		code    model.ActionCode
		title   string
	}{
		{"favorites", model.Metrics{FavoritesAdded: 8}, model.ActionableState{FavoritesCount: 3}, model.ActionOpenFavorites, manyFavoritesTitle},
		{"draft", model.Metrics{ListingsCreated: 4, ListingsPublished: 1}, model.ActionableState{CurrentDrafts: 1, DraftListingID: uuid.MustParse("33333333-3333-4333-8333-333333333333"), FavoritesCount: 5}, model.ActionFinishDraft, createdNotPublishedTitle},
		{"views", model.Metrics{TotalViews: 20, RepeatedViews: 4, TopCategoryCode: "auto"}, model.ActionableState{}, model.ActionOpenTopCategory, viewedWithoutFavoritesTitle},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Build(tc.metrics, tc.state)
			if got.Code != tc.code || got.Title != tc.title {
				t.Fatalf("got %+v", got)
			}
			if got.Target.Route == nil && got.Target.Category == nil && got.Target.Listing == nil {
				t.Fatalf("missing target: %+v", got)
			}
		})
	}
}

func TestCreationGapWinsOverFavorites(t *testing.T) {
	got := Build(model.Metrics{ListingsCreated: 4, ListingsPublished: 1}, model.ActionableState{FavoritesCount: 10})
	if got.Code != model.ActionCreateListing {
		t.Fatalf("got %+v", got)
	}
}
