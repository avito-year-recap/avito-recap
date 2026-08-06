package structural

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

func ValidateActionableState(state model.ActionableState) error {
	if state.CapturedAt.IsZero() {
		return fmt.Errorf("%w: captured time is required", ErrInvalidActionableState)
	}
	if (state.DraftListingID == uuid.Nil) != (state.CurrentDrafts == 0) {
		return fmt.Errorf("%w: draft count and addressable draft id must be present together", ErrInvalidActionableState)
	}
	if (state.OpenDialogID == uuid.Nil) != (state.OpenDialogs == 0) {
		return fmt.Errorf("%w: open-dialog count and addressable dialog id must be present together", ErrInvalidActionableState)
	}
	if (state.ActiveListingID == uuid.Nil) != (state.ActiveListings == 0) {
		return fmt.Errorf("%w: active-listing count and addressable listing id must be present together", ErrInvalidActionableState)
	}
	return nil
}

func ValidateNextAction(value model.NextAction) error {
	if !model.IsKnownActionCode(value.Code) {
		return fmt.Errorf("unknown code %q", value.Code)
	}
	if value.Title == "" || value.Description == "" || value.ButtonText == "" || value.Reason == "" {
		return errors.New("text is incomplete")
	}
	if err := ValidateActionTarget(value.Target); err != nil {
		return err
	}
	return ValidateTargetForAction(value.Code, value.Target)
}

func ValidateActionTarget(target model.ActionTarget) error {
	count := 0
	if target.Route != nil {
		count++
		if !isSafeApplicationRoute(target.Route.Route) {
			return errors.New("route target must contain a known safe application route")
		}
	}
	if target.Category != nil {
		count++
		if !isSafeCategoryCode(target.Category.CategoryCode) {
			return errors.New("safe category target code is required")
		}
	}
	if target.Listing != nil {
		count++
		if target.Listing.ListingID == uuid.Nil {
			return errors.New("listing target id is required")
		}
	}
	if target.Dialog != nil {
		count++
		if target.Dialog.DialogID == uuid.Nil {
			return errors.New("dialog target id is required")
		}
	}
	if target.Search != nil {
		count++
		if !isSafeCategoryCode(target.Search.CategoryCode) {
			return errors.New("safe search target category is required")
		}
	}
	if count != 1 {
		return fmt.Errorf("action target must contain exactly one destination, got %d", count)
	}
	return nil
}

func ValidateTargetForAction(code model.ActionCode, target model.ActionTarget) error {
	switch code {
	case model.ActionFinishDraft, model.ActionImproveListings, model.ActionViewSimilarListings:
		if target.Listing == nil {
			return fmt.Errorf("action %s requires a listing target", code)
		}
	case model.ActionContinueDialogs:
		if target.Dialog == nil {
			return fmt.Errorf("action %s requires a dialog target", code)
		}
	case model.ActionOpenTopCategory:
		if target.Category == nil {
			return fmt.Errorf("action %s requires a category target", code)
		}
	case model.ActionSaveSearch:
		if target.Search == nil {
			return fmt.Errorf("action %s requires a search target", code)
		}
	case model.ActionOpenFavorites:
		if target.Route == nil || target.Route.Route != "/favorites" {
			return fmt.Errorf("action %s requires /favorites route", code)
		}
	case model.ActionCreateFirstListing, model.ActionCreateListing:
		if target.Route == nil || target.Route.Route != "/listings/new" {
			return fmt.Errorf("action %s requires /listings/new route", code)
		}
	case model.ActionExploreRecommendations:
		if target.Route == nil || target.Route.Route != "/recommendations" {
			return fmt.Errorf("action %s requires /recommendations route", code)
		}
	}
	return nil
}
