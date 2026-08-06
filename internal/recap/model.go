package recap

import (
	"time"

	"github.com/google/uuid"
)

const CurrentRulesVersion = "3.4.1"

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

type BehaviorCode string

const (
	BehaviorActiveSeller   BehaviorCode = "ACTIVE_SELLER"
	BehaviorStartingSeller BehaviorCode = "STARTING_SELLER"
	BehaviorDecisiveBuyer  BehaviorCode = "DECISIVE_BUYER"
	BehaviorFindHunter     BehaviorCode = "FIND_HUNTER"
	BehaviorResearcher     BehaviorCode = "RESEARCHER"
	BehaviorUniversal      BehaviorCode = "UNIVERSAL_USER"
)

type BehaviorEvidence struct {
	Metric    string  `json:"metric"`
	Actual    float64 `json:"actual"`
	Threshold float64 `json:"threshold"`
	Points    uint32  `json:"points"`
	Detail    string  `json:"detail"`
}

type Behavior struct {
	Code        BehaviorCode       `json:"code"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Reason      string             `json:"reason"`
	Score       uint32             `json:"score"`
	Evidence    []BehaviorEvidence `json:"evidence,omitempty"`
}

type AchievementCategory string

const (
	AchievementCategorySelling     AchievementCategory = "SELLING"
	AchievementCategoryBuying      AchievementCategory = "BUYING"
	AchievementCategoryDiscovery   AchievementCategory = "DISCOVERY"
	AchievementCategoryCollection  AchievementCategory = "COLLECTION"
	AchievementCategoryVersatility AchievementCategory = "VERSATILITY"
	AchievementCategoryInterest    AchievementCategory = "INTEREST"
)

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
	AchievementStyleIcon           AchievementCode = "STYLE_ICON"
	AchievementFashionableMan      AchievementCode = "FASHIONABLE_MAN"
	AchievementTraveler            AchievementCode = "TRAVELER"
	AchievementForTheSoul          AchievementCode = "FOR_THE_SOUL"
	AchievementBookworm            AchievementCode = "BOOKWORM"
	AchievementBeautyConnoisseur   AchievementCode = "BEAUTY_CONNOISSEUR"
	AchievementInTheRhythmOfMusic  AchievementCode = "IN_THE_RHYTHM_OF_MUSIC"
	AchievementWorldOfPlay         AchievementCode = "WORLD_OF_PLAY"
	AchievementMasterCraft         AchievementCode = "MASTER_CRAFT"
	AchievementCaringOwner         AchievementCode = "CARING_OWNER"
	AchievementLittleDiscoveries   AchievementCode = "LITTLE_DISCOVERIES"
)

const (
	AchievementActivePublisher  = AchievementConsistentPublisher
	AchievementFavoritesCurator = AchievementMasterOfFavorites
	AchievementCategoryExplorer = AchievementBroadInterests
)

type Achievement struct {
	Code        AchievementCode     `json:"code"`
	Category    AchievementCategory `json:"category"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Reason      string              `json:"reason"`
	Priority    int                 `json:"-"`
	Strength    uint64              `json:"-"`
	Shareable   bool                `json:"shareable"`
}

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
	CardShare             CardType = "SHARE"
)

// CardPayload is a sealed union: only the payload types below can be assigned.
type CardPayload interface {
	isCardPayload()
}

type YearActivityPayload struct {
	TotalEvents        uint64 `json:"totalEvents"`
	Searches           uint64 `json:"searches"`
	TotalViews         uint64 `json:"totalViews"`
	FavoritesAdded     uint64 `json:"favoritesAdded"`
	ChatsStarted       uint64 `json:"chatsStarted"`
	ListingsPublished  uint64 `json:"listingsPublished"`
	PurchasesCompleted uint64 `json:"purchasesCompleted"`
	SalesCompleted     uint64 `json:"salesCompleted"`
}

func (YearActivityPayload) isCardPayload() {}

type TopCategoryPayload struct {
	CategoryCode  string `json:"categoryCode"`
	Category      string `json:"category"`
	CategoryViews uint64 `json:"categoryViews"`
}

func (TopCategoryPayload) isCardPayload() {}

type ActiveMonthPayload struct {
	Month uint32 `json:"month"`
}

func (ActiveMonthPayload) isCardPayload() {}

type BehaviorPayload struct {
	Code     BehaviorCode       `json:"code"`
	Score    uint32             `json:"score"`
	Evidence []BehaviorEvidence `json:"evidence,omitempty"`
}

func (BehaviorPayload) isCardPayload() {}

type AchievementPayload struct {
	Codes []AchievementCode `json:"codes"`
}

func (AchievementPayload) isCardPayload() {}

type ActionPayload struct {
	Code   ActionCode   `json:"code"`
	Target ActionTarget `json:"target"`
}

func (ActionPayload) isCardPayload() {}

type Card struct {
	ID          string      `json:"id"`
	Type        CardType    `json:"type"`
	Position    uint32      `json:"position"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Explanation string      `json:"explanation,omitempty"`
	Shareable   bool        `json:"shareable"`
	Payload     CardPayload `json:"payload,omitempty"`
}

type RecapKey struct {
	ProfileID    uuid.UUID `json:"profileId"`
	Year         uint32    `json:"year"`
	RulesVersion string    `json:"rulesVersion"`
	RulesDigest  string    `json:"rulesDigest"`
}

type Recap struct {
	ID              uuid.UUID       `json:"id"`
	ShareID         uuid.UUID       `json:"shareId"`
	Profile         Profile         `json:"profile"`
	Year            uint32          `json:"year"`
	Period          RecapPeriod     `json:"period"`
	RulesVersion    string          `json:"rulesVersion"`
	RulesDigest     string          `json:"rulesDigest"`
	Metrics         Metrics         `json:"metrics"`
	ActionableState ActionableState `json:"actionableState"`
	Behavior        Behavior        `json:"behavior"`
	Achievements    []Achievement   `json:"achievements"`
	Cards           []Card          `json:"cards"`
	NextAction      NextAction      `json:"nextAction"`
	GeneratedAt     time.Time       `json:"generatedAt"`
}

func (r Recap) Key() RecapKey {
	return RecapKey{ProfileID: r.Profile.ID, Year: r.Year, RulesVersion: r.RulesVersion, RulesDigest: r.RulesDigest}
}

type ShareCard struct {
	ShareID          uuid.UUID `json:"shareId"`
	Year             uint32    `json:"year"`
	PrivacyVersion   string    `json:"privacyVersion"`
	BehaviorTitle    string    `json:"behaviorTitle"`
	AchievementTitle string    `json:"achievementTitle,omitempty"`
	TopCategory      string    `json:"topCategory,omitempty"`
}

// ShareCard is both the strict public DTO and the payload of the final story
// card. Keeping one type guarantees that the user sees exactly the same safe
// data that will be available through the public share link.
func (ShareCard) isCardPayload() {}
