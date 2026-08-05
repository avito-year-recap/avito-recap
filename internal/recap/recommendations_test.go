package recap

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBuildNextActionAllBranches(t *testing.T) {
	listingID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	dialogID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	purchaseID := uuid.MustParse("55555555-5555-4555-8555-555555555555")

	tests := []struct {
		name     string
		metrics  Metrics
		state    ActionableState
		expected ActionCode
	}{
		{name: "finish current draft", metrics: Metrics{}, state: ActionableState{CurrentDrafts: 2, DraftListingID: listingID}, expected: ActionFinishDraft},
		{name: "continue current dialog", metrics: Metrics{}, state: ActionableState{OpenDialogs: 1, OpenDialogID: dialogID}, expected: ActionContinueDialogs},
		{name: "create listing for active seller", metrics: Metrics{ListingsPublished: 5, SalesCompleted: 3}, expected: ActionCreateListing},
		{name: "view similar for decisive buyer", metrics: Metrics{ChatsStarted: 12, ChatsWithPurchase: 3, PurchasesCompleted: 3}, state: ActionableState{LastPurchasedListingID: purchaseID}, expected: ActionViewSimilarListings},
		{name: "save search for researcher", metrics: Metrics{TotalViews: 100, UniqueListings: 100, ChatsStarted: 4, CategoriesCount: 5, TopCategoryCode: "cars", TopCategory: "Авто"}, expected: ActionSaveSearch},
		{name: "open favorites for find hunter", metrics: Metrics{TotalViews: 20, UniqueListings: 16, RepeatedViews: 4, FavoritesAdded: 3}, state: ActionableState{FavoritesCount: 2}, expected: ActionOpenFavorites},
		{name: "improve active listing", metrics: Metrics{ListingsPublished: 3}, state: ActionableState{ActiveListings: 3, ActiveListingID: listingID, HasEverPublishedListing: true}, expected: ActionImproveListings},
		{name: "open top category", metrics: Metrics{TopCategoryCode: "electronics", TopCategory: "Электроника"}, state: ActionableState{HasEverPublishedListing: true}, expected: ActionOpenTopCategory},
		{name: "create listing only when current state confirms no publications", metrics: Metrics{}, state: ActionableState{}, expected: ActionCreateListing},
		{name: "neutral explore fallback", metrics: Metrics{}, state: ActionableState{HasEverPublishedListing: true}, expected: ActionExploreRecommendations},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := BuildNextAction(test.metrics, test.state)
			if actual.Code != test.expected {
				t.Fatalf("expected %s, got %s", test.expected, actual.Code)
			}
			if actual.Title == "" || actual.Description == "" || actual.ButtonText == "" || actual.Reason == "" {
				t.Fatalf("next action contains empty text: %+v", actual)
			}
			if err := validateNextAction(actual); err != nil {
				t.Fatalf("invalid action: %v", err)
			}
		})
	}
}

func TestBuildNextActionUsesCurrentStateBeforeHistoricalBehavior(t *testing.T) {
	draftID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	metrics := Metrics{ListingsPublished: 10, SalesCompleted: 8}
	action := BuildNextAction(metrics, ActionableState{CurrentDrafts: 1, DraftListingID: draftID})
	if action.Code != ActionFinishDraft {
		t.Fatalf("current actionable work must win, got %s", action.Code)
	}
	if action.Target.Listing == nil || action.Target.Listing.ListingID != draftID {
		t.Fatalf("unexpected target: %+v", action.Target)
	}
}

func TestBuildNextActionNormalizesTopCategory(t *testing.T) {
	action := BuildNextAction(Metrics{TopCategoryCode: "  electronics ", TopCategory: "  Электроника\n"}, ActionableState{HasEverPublishedListing: true})
	if action.Code != ActionOpenTopCategory {
		t.Fatalf("expected top-category action, got %s", action.Code)
	}
	if action.Description != "Вернись в категорию «Электроника» и проверь новые варианты." {
		t.Fatalf("category was not normalized: %q", action.Description)
	}
	if action.Target.Category == nil || action.Target.Category.CategoryCode != "electronics" {
		t.Fatalf("category target was not normalized: %+v", action.Target)
	}
}

func TestEveryActionHasExactlyOneStructuredTarget(t *testing.T) {
	actions := []NextAction{
		BuildNextAction(Metrics{}, ActionableState{}),
		BuildNextAction(Metrics{}, ActionableState{HasEverPublishedListing: true}),
		BuildNextAction(Metrics{TopCategoryCode: "cars", TopCategory: "Авто"}, ActionableState{HasEverPublishedListing: true}),
	}
	for _, action := range actions {
		if err := validateActionTarget(action.Target); err != nil {
			t.Fatalf("action %s target: %v", action.Code, err)
		}
	}
}

func TestEveryActionBranchCarriesItsRequiredTarget(t *testing.T) {
	listingID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	dialogID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
	purchaseID := uuid.MustParse("55555555-5555-4555-8555-555555555555")

	tests := []struct {
		name       string
		metrics    Metrics
		state      ActionableState
		wantCode   ActionCode
		wantTarget string
		wantValue  string
	}{
		{name: "finish draft", state: ActionableState{CurrentDrafts: 1, DraftListingID: listingID}, wantCode: ActionFinishDraft, wantTarget: "listing", wantValue: listingID.String()},
		{name: "continue dialog", state: ActionableState{OpenDialogs: 1, OpenDialogID: dialogID}, wantCode: ActionContinueDialogs, wantTarget: "dialog", wantValue: dialogID.String()},
		{name: "active seller", metrics: Metrics{ListingsPublished: 5, SalesCompleted: 3}, wantCode: ActionCreateListing, wantTarget: "route", wantValue: "/listings/new"},
		{name: "similar listing", metrics: Metrics{ChatsStarted: 12, ChatsWithPurchase: 3, PurchasesCompleted: 3}, state: ActionableState{LastPurchasedListingID: purchaseID}, wantCode: ActionViewSimilarListings, wantTarget: "listing", wantValue: purchaseID.String()},
		{name: "save search", metrics: Metrics{TotalViews: 100, UniqueListings: 100, ChatsStarted: 4, CategoriesCount: 5, TopCategoryCode: "cars", TopCategory: "Авто"}, wantCode: ActionSaveSearch, wantTarget: "search", wantValue: "cars"},
		{name: "open favorites", metrics: Metrics{TotalViews: 20, UniqueListings: 16, RepeatedViews: 4, FavoritesAdded: 3}, state: ActionableState{FavoritesCount: 2}, wantCode: ActionOpenFavorites, wantTarget: "route", wantValue: "/favorites"},
		{name: "improve listing", metrics: Metrics{ListingsPublished: 3}, state: ActionableState{ActiveListings: 3, ActiveListingID: listingID, HasEverPublishedListing: true}, wantCode: ActionImproveListings, wantTarget: "listing", wantValue: listingID.String()},
		{name: "open category", metrics: Metrics{TopCategoryCode: "electronics", TopCategory: "Электроника"}, state: ActionableState{HasEverPublishedListing: true}, wantCode: ActionOpenTopCategory, wantTarget: "category", wantValue: "electronics"},
		{name: "create listing", state: ActionableState{}, wantCode: ActionCreateListing, wantTarget: "route", wantValue: "/listings/new"},
		{name: "explore fallback", state: ActionableState{HasEverPublishedListing: true}, wantCode: ActionExploreRecommendations, wantTarget: "route", wantValue: "/recommendations"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			action := BuildNextAction(test.metrics, test.state)
			if action.Code != test.wantCode {
				t.Fatalf("code = %s, want %s", action.Code, test.wantCode)
			}
			kind, value := actionTargetIdentity(action.Target)
			if kind != test.wantTarget || value != test.wantValue {
				t.Fatalf("target = %s:%s, want %s:%s; action=%+v", kind, value, test.wantTarget, test.wantValue, action)
			}
			if err := validateNextAction(action); err != nil {
				t.Fatalf("action is invalid: %v", err)
			}
		})
	}
}

func TestFallbackRecommendationIsNeutralAndMeaningful(t *testing.T) {
	action := BuildNextAction(Metrics{}, ActionableState{HasEverPublishedListing: true})
	if action.Code != ActionExploreRecommendations {
		t.Fatalf("fallback code = %s, want %s", action.Code, ActionExploreRecommendations)
	}
	text := strings.ToLower(strings.Join([]string{action.Title, action.Description, action.ButtonText, action.Reason}, " "))
	for _, misleading := range []string{"первое объявление", "первое", "черновик", "открытый диалог"} {
		if strings.Contains(text, misleading) {
			t.Fatalf("fallback makes an unsupported claim %q: %+v", misleading, action)
		}
	}
	if action.Target.Route == nil || action.Target.Route.Route != "/recommendations" {
		t.Fatalf("fallback must lead to a neutral recommendations route: %+v", action.Target)
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
