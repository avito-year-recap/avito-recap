package recap

import (
	"testing"

	"github.com/google/uuid"
)

func TestBuildNextActionUsesOnlyApprovedVariants(t *testing.T) {
	draftID := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	tests := []struct {
		name            string
		metrics         Metrics
		state           ActionableState
		wantCode        ActionCode
		wantTitle       string
		wantDescription string
	}{
		{
			name:            "many favorites",
			metrics:         Metrics{TotalViews: 20, UniqueListings: 16, RepeatedViews: 4, FavoritesAdded: 8},
			state:           ActionableState{FavoritesCount: 3},
			wantCode:        ActionOpenFavorites,
			wantTitle:       manyFavoritesTitle,
			wantDescription: manyFavoritesDescription,
		},
		{
			name:            "created but not published with current draft",
			metrics:         Metrics{ListingsCreated: 7, ListingsPublished: 2},
			state:           ActionableState{CurrentDrafts: 1, DraftListingID: draftID, FavoritesCount: 5},
			wantCode:        ActionFinishDraft,
			wantTitle:       createdNotPublishedTitle,
			wantDescription: createdNotPublishedDescription,
		},
		{
			name:            "created but not published without addressable draft",
			metrics:         Metrics{ListingsCreated: 4, ListingsPublished: 1},
			wantCode:        ActionCreateListing,
			wantTitle:       createdNotPublishedTitle,
			wantDescription: createdNotPublishedDescription,
		},
		{
			name:            "viewed without favorites in category",
			metrics:         Metrics{TotalViews: 20, UniqueListings: 16, RepeatedViews: 4, TopCategoryCode: "cars", TopCategory: "Авто"},
			wantCode:        ActionOpenTopCategory,
			wantTitle:       viewedWithoutFavoritesTitle,
			wantDescription: viewedWithoutFavoritesDescription,
		},
		{
			name:            "viewed without favorites fallback",
			metrics:         Metrics{},
			wantCode:        ActionExploreRecommendations,
			wantTitle:       viewedWithoutFavoritesTitle,
			wantDescription: viewedWithoutFavoritesDescription,
		},
	}

	approved := map[string]string{
		manyFavoritesTitle:          manyFavoritesDescription,
		createdNotPublishedTitle:    createdNotPublishedDescription,
		viewedWithoutFavoritesTitle: viewedWithoutFavoritesDescription,
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := BuildNextAction(test.metrics, test.state)
			if action.Code != test.wantCode {
				t.Fatalf("code = %s, want %s", action.Code, test.wantCode)
			}
			if action.Title != test.wantTitle || action.Description != test.wantDescription {
				t.Fatalf("unexpected copy: %+v", action)
			}
			if approved[action.Title] != action.Description {
				t.Fatalf("unapproved next-action variant: %+v", action)
			}
			if action.ButtonText == "" || action.Reason == "" {
				t.Fatalf("next action contains empty text: %+v", action)
			}
			if err := validateNextAction(action); err != nil {
				t.Fatalf("invalid action: %v", err)
			}
		})
	}
}

func TestCreationPublicationGapWinsOverFavorites(t *testing.T) {
	draftID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	metrics := Metrics{ListingsCreated: 7, ListingsPublished: 2, FavoritesAdded: 30}
	state := ActionableState{CurrentDrafts: 1, DraftListingID: draftID, FavoritesCount: 12}
	action := BuildNextAction(metrics, state)
	if action.Code != ActionFinishDraft || action.Title != createdNotPublishedTitle {
		t.Fatalf("creation-publication gap must win, got %+v", action)
	}
	if action.Target.Listing == nil || action.Target.Listing.ListingID != draftID {
		t.Fatalf("unexpected target: %+v", action.Target)
	}
}

func TestBuildNextActionNormalizesTopCategoryTarget(t *testing.T) {
	action := BuildNextAction(Metrics{
		TotalViews: 10, UniqueListings: 8, RepeatedViews: 2,
		TopCategoryCode: "  electronics ", TopCategory: "  Электроника\n",
	}, ActionableState{})
	if action.Code != ActionOpenTopCategory {
		t.Fatalf("expected top-category action, got %s", action.Code)
	}
	if action.Title != viewedWithoutFavoritesTitle || action.Description != viewedWithoutFavoritesDescription {
		t.Fatalf("unexpected copy: %+v", action)
	}
	if action.Target.Category == nil || action.Target.Category.CategoryCode != "electronics" {
		t.Fatalf("category target was not normalized: %+v", action.Target)
	}
}

func TestEveryApprovedActionHasExactlyOneStructuredTarget(t *testing.T) {
	draftID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	actions := []NextAction{
		BuildNextAction(Metrics{}, ActionableState{}),
		BuildNextAction(Metrics{TopCategoryCode: "cars", TopCategory: "Авто"}, ActionableState{}),
		BuildNextAction(Metrics{FavoritesAdded: 5}, ActionableState{FavoritesCount: 2}),
		BuildNextAction(Metrics{ListingsCreated: 4, ListingsPublished: 1}, ActionableState{}),
		BuildNextAction(Metrics{ListingsCreated: 4, ListingsPublished: 1}, ActionableState{CurrentDrafts: 1, DraftListingID: draftID}),
	}
	for _, action := range actions {
		if err := validateActionTarget(action.Target); err != nil {
			t.Fatalf("action %s target: %v", action.Code, err)
		}
		if err := validateNextAction(action); err != nil {
			t.Fatalf("action %s is invalid: %v", action.Code, err)
		}
	}
}

func actionTargetIdentity(target ActionTarget) (string, string) {
	switch {
	case target.Route != nil:
		return "route", target.Route.Route
	case target.Category != nil:
		return "category", target.Category.CategoryCode
	case target.Listing != nil:
		return "listing", target.Listing.ListingID.String()
	case target.Dialog != nil:
		return "dialog", target.Dialog.DialogID.String()
	case target.Search != nil:
		return "search", target.Search.CategoryCode
	default:
		return "", ""
	}
}
