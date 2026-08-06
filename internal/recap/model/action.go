package model

import "github.com/google/uuid"

type ActionCode string

const (
	ActionFinishDraft            ActionCode = "FINISH_DRAFT"
	ActionOpenFavorites          ActionCode = "OPEN_FAVORITES"
	ActionImproveListings        ActionCode = "IMPROVE_LISTINGS"
	ActionContinueDialogs        ActionCode = "CONTINUE_DIALOGS"
	ActionOpenTopCategory        ActionCode = "OPEN_TOP_CATEGORY"
	ActionCreateFirstListing     ActionCode = "CREATE_FIRST_LISTING" // legacy stored value
	ActionCreateListing          ActionCode = "CREATE_LISTING"
	ActionSaveSearch             ActionCode = "SAVE_SEARCH"
	ActionViewSimilarListings    ActionCode = "VIEW_SIMILAR_LISTINGS"
	ActionExploreRecommendations ActionCode = "EXPLORE_RECOMMENDATIONS"
)

const ActionContinueChats = ActionContinueDialogs

type RouteTarget struct {
	Route string `json:"route"`
}

type CategoryTarget struct {
	CategoryCode string `json:"categoryCode"`
}

type ListingTarget struct {
	ListingID uuid.UUID `json:"listingId"`
}

type DialogTarget struct {
	DialogID uuid.UUID `json:"dialogId"`
}

type SearchTarget struct {
	CategoryCode string `json:"categoryCode"`
}

// ActionTarget mirrors protobuf oneof semantics. Exactly one field must be set.
type ActionTarget struct {
	Route    *RouteTarget    `json:"route,omitempty"`
	Category *CategoryTarget `json:"category,omitempty"`
	Listing  *ListingTarget  `json:"listing,omitempty"`
	Dialog   *DialogTarget   `json:"dialog,omitempty"`
	Search   *SearchTarget   `json:"search,omitempty"`
}

type NextAction struct {
	Code        ActionCode   `json:"code"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	ButtonText  string       `json:"buttonText"`
	Reason      string       `json:"reason"`
	Target      ActionTarget `json:"target"`
}
