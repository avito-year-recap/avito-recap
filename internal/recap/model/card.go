package model

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
