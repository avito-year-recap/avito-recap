package recap

import (
	"time"

	"github.com/google/uuid"
)

const CurrentRulesVersion = "2.1.0"

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

	FavoriteRate    float64 `json:"favoriteRate"`
	ChatRate        float64 `json:"chatRate"`
	RepeatRate      float64 `json:"repeatRate"`
	PublicationRate float64 `json:"publicationRate"`
	SaleRate        float64 `json:"saleRate"`
	PurchaseRate    float64 `json:"purchaseRate"`
}

type BehaviorCode string

const (
	BehaviorActiveSeller   BehaviorCode = "ACTIVE_SELLER"
	BehaviorStartingSeller BehaviorCode = "STARTING_SELLER"
	BehaviorDecisiveBuyer  BehaviorCode = "DECISIVE_BUYER"
	BehaviorFindHunter     BehaviorCode = "FIND_HUNTER"
	BehaviorResearcher     BehaviorCode = "RESEARCHER"
	BehaviorUniversal      BehaviorCode = "UNIVERSAL_USER"
)

type Behavior struct {
	Code        BehaviorCode `json:"code"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Reason      string       `json:"reason"`
}

type AchievementCode string

const (
	AchievementSuccessfulSeller    AchievementCode = "SUCCESSFUL_SELLER"
	AchievementConsistentPublisher AchievementCode = "CONSISTENT_PUBLISHER"
	AchievementAttentiveResearcher AchievementCode = "ATTENTIVE_RESEARCHER"
	AchievementMasterOfFavorites   AchievementCode = "MASTER_OF_FAVORITES"
	AchievementBroadInterests      AchievementCode = "BROAD_INTERESTS"
	AchievementAllRounder          AchievementCode = "ALL_ROUNDER"
	AchievementFirstSellingSteps   AchievementCode = "FIRST_SELLING_STEPS"
	AchievementDealCloser          AchievementCode = "DEAL_CLOSER"
	AchievementQuickDecision       AchievementCode = "QUICK_DECISION"
)

// Backward-compatible aliases for code that used the first MVP names.
const (
	AchievementActivePublisher  = AchievementConsistentPublisher
	AchievementFavoritesCurator = AchievementMasterOfFavorites
	AchievementCategoryExplorer = AchievementBroadInterests
)

type Achievement struct {
	Code        AchievementCode `json:"code"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Reason      string          `json:"reason"`
	Priority    int             `json:"-"`
	Shareable   bool            `json:"shareable"`
}

type ActionCode string

const (
	ActionFinishDraft         ActionCode = "FINISH_DRAFT"
	ActionOpenFavorites       ActionCode = "OPEN_FAVORITES"
	ActionImproveListings     ActionCode = "IMPROVE_LISTINGS"
	ActionContinueDialogs     ActionCode = "CONTINUE_DIALOGS"
	ActionOpenTopCategory     ActionCode = "OPEN_TOP_CATEGORY"
	ActionCreateFirstListing  ActionCode = "CREATE_FIRST_LISTING"
	ActionCreateListing       ActionCode = "CREATE_LISTING"
	ActionSaveSearch          ActionCode = "SAVE_SEARCH"
	ActionViewSimilarListings ActionCode = "VIEW_SIMILAR_LISTINGS"
)

const ActionContinueChats = ActionContinueDialogs

type NextAction struct {
	Code        ActionCode `json:"code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ButtonText  string     `json:"buttonText"`
	Reason      string     `json:"reason"`
}

type CardType string

const (
	CardIntro             CardType = "INTRO"
	CardYearActivity      CardType = "YEAR_ACTIVITY"
	CardTopCategory       CardType = "TOP_CATEGORY"
	CardActiveMonth       CardType = "ACTIVE_MONTH"
	CardBehavior          CardType = "BEHAVIOR"
	CardAchievement       CardType = "ACHIEVEMENT"
	CardMissedOpportunity CardType = "MISSED_OPPORTUNITY"
	CardNextAction        CardType = "NEXT_ACTION"
	CardSummary           CardType = "SUMMARY"
)

type CardPayload struct {
	TotalEvents        uint64 `json:"totalEvents,omitempty"`
	Searches           uint64 `json:"searches,omitempty"`
	TotalViews         uint64 `json:"totalViews,omitempty"`
	FavoritesAdded     uint64 `json:"favoritesAdded,omitempty"`
	ChatsStarted       uint64 `json:"chatsStarted,omitempty"`
	ListingsPublished  uint64 `json:"listingsPublished,omitempty"`
	PurchasesCompleted uint64 `json:"purchasesCompleted,omitempty"`
	SalesCompleted     uint64 `json:"salesCompleted,omitempty"`

	CategoryCode  string `json:"categoryCode,omitempty"`
	Category      string `json:"category,omitempty"`
	CategoryViews uint64 `json:"categoryViews,omitempty"`
	Month         uint32 `json:"month,omitempty"`

	BehaviorCode     BehaviorCode      `json:"behaviorCode,omitempty"`
	AchievementCode  AchievementCode   `json:"achievementCode,omitempty"`
	AchievementCodes []AchievementCode `json:"achievementCodes,omitempty"`
	ActionCode       ActionCode        `json:"actionCode,omitempty"`
}

type Card struct {
	ID          string      `json:"id"`
	Type        CardType    `json:"type"`
	Position    uint32      `json:"position"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Explanation string      `json:"explanation,omitempty"`
	Shareable   bool        `json:"shareable"`
	Payload     CardPayload `json:"payload"`
}

type Recap struct {
	ID           uuid.UUID     `json:"id"`
	Profile      Profile       `json:"profile"`
	Year         uint32        `json:"year"`
	RulesVersion string        `json:"rulesVersion"`
	Metrics      Metrics       `json:"metrics"`
	Behavior     Behavior      `json:"behavior"`
	Achievements []Achievement `json:"achievements"`
	Cards        []Card        `json:"cards"`
	NextAction   NextAction    `json:"nextAction"`
	GeneratedAt  time.Time     `json:"generatedAt"`
}

type ShareCard struct {
	RecapID          uuid.UUID `json:"recapId"`
	Year             uint32    `json:"year"`
	BehaviorTitle    string    `json:"behaviorTitle"`
	AchievementTitle string    `json:"achievementTitle,omitempty"`
	TopCategory      string    `json:"topCategory,omitempty"`
}
