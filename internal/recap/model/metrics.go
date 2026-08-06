package model

import (
	"github.com/google/uuid"
	"time"
)

type Metrics struct {
	TotalEvents        uint64 `json:"totalEvents"`
	Searches           uint64 `json:"searches"`
	TotalViews         uint64 `json:"totalViews"`
	UniqueListings     uint64 `json:"uniqueListings"`
	RepeatedViews      uint64 `json:"repeatedViews"`
	FavoritesAdded     uint64 `json:"favoritesAdded"`
	ChatsStarted       uint64 `json:"chatsStarted"`
	ListingsCreated    uint64 `json:"listingsCreated"`
	ListingsPublished  uint64 `json:"listingsPublished"`
	PurchasesCompleted uint64 `json:"purchasesCompleted"`
	SalesCompleted     uint64 `json:"salesCompleted"`
	ActiveDays         uint64 `json:"activeDays"`
	CategoriesCount    uint64 `json:"categoriesCount"`
	ChatsWithPurchase  uint64 `json:"chatsWithPurchase"`

	TopCategoryCode      string `json:"topCategoryCode,omitempty"`
	TopCategory          string `json:"topCategory,omitempty"`
	TopCategoryViews     uint64 `json:"topCategoryViews"`
	TopCategoryShareable bool   `json:"topCategoryShareable"`
	MostActiveMonth      uint32 `json:"mostActiveMonth"`

	RepeatRate   float64 `json:"repeatRate"`
	PurchaseRate float64 `json:"purchaseRate"`

	// CategoryActivities contains the per-category evidence used by thematic
	// achievements. The slice is normalized into category-code order so the
	// stored recap and its integrity digest remain deterministic.
	CategoryActivities []CategoryActivity `json:"categoryActivities,omitempty"`
}

type CategoryActivity struct {
	CategoryCode       string `json:"categoryCode"`
	Category           string `json:"category"`
	Shareable          bool   `json:"shareable"`
	Views              uint64 `json:"views"`
	FavoritesAdded     uint64 `json:"favoritesAdded"`
	PurchasesCompleted uint64 `json:"purchasesCompleted"`
}

// ActionableState is a point-in-time snapshot used only to select executable CTAs.
// Historical annual counters must not be used as a substitute for this state.
type ActionableState struct {
	CapturedAt time.Time `json:"capturedAt"`

	CurrentDrafts  uint64    `json:"currentDrafts"`
	DraftListingID uuid.UUID `json:"draftListingId,omitempty"`

	OpenDialogs  uint64    `json:"openDialogs"`
	OpenDialogID uuid.UUID `json:"openDialogId,omitempty"`

	ActiveListings  uint64    `json:"activeListings"`
	ActiveListingID uuid.UUID `json:"activeListingId,omitempty"`

	FavoritesCount uint64 `json:"favoritesCount"`

	HasSavedSearchForTopCategory bool      `json:"hasSavedSearchForTopCategory"`
	LastPurchasedListingID       uuid.UUID `json:"lastPurchasedListingId,omitempty"`
	HasEverPublishedListing      bool      `json:"hasEverPublishedListing"`
}
