package model

import (
	"time"

	"github.com/google/uuid"
)

// Event is a single raw, immutable fact about a user action on the platform.
// AnalyticsStorage.CalculateMetrics is derived from a profile's events for a
// given year, not stored as a pre-computed blob — Event is the source of
// truth, model.Metrics is a cache of an aggregation over it.
type Event struct {
	ID         uuid.UUID    `json:"id"`
	ProfileID  uuid.UUID    `json:"profileId"`
	Type       ActivityType `json:"type"`
	OccurredAt time.Time    `json:"occurredAt"`

	// Category is empty for events that are not category-scoped (e.g. a
	// search that did not resolve to a single category).
	Category string `json:"category,omitempty"`

	// AdID identifies the listing an event refers to. Repeated ListingView
	// events sharing an AdID are what makes UniqueListings/RepeatedViews
	// derivable from raw events instead of being separately tracked.
	AdID *uint64 `json:"adId,omitempty"`

	// DialogID identifies the conversation an event refers to. A
	// PurchaseCompleted event carrying the same DialogID as a ChatStarted
	// event is what makes ChatsWithPurchase derivable from raw events.
	DialogID *uint64 `json:"dialogId,omitempty"`
}
