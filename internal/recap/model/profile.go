package model

import (
	"github.com/google/uuid"
	"time"
)

type ActivityType string

const (
	ActivitySearch            ActivityType = "search"
	ActivityListingView       ActivityType = "listing_view"
	ActivityFavoriteAdded     ActivityType = "favorite_added"
	ActivityChatStarted       ActivityType = "chat_started"
	ActivityListingCreated    ActivityType = "listing_created"
	ActivityListingPublished  ActivityType = "listing_published"
	ActivityPurchaseCompleted ActivityType = "purchase_completed"
	ActivitySaleCompleted     ActivityType = "sale_completed"
)

type Profile struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	DisplayName string    `json:"displayName"`
	Description string    `json:"description"`
	AvatarURL   string    `json:"avatarUrl,omitempty"`
}

// RecapPeriod is a half-open UTC interval [StartAt, EndAt).
// Annual recaps are generated only for completed calendar years and are final.
type RecapPeriod struct {
	Year    uint32    `json:"year"`
	StartAt time.Time `json:"startAt"`
	EndAt   time.Time `json:"endAt"`
	Final   bool      `json:"final"`
}
