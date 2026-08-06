package ruleset

import (
	"errors"
	"regexp"

	"github.com/year-recap/internal/recap/model"
)

var ErrInvalidRuleset = errors.New("invalid ruleset")

const currentRulesAlgorithm = "recap-v3.4-three-next-actions-v1"

var semanticVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type BehaviorThresholds struct {
	ActiveSellerMinListings      uint64
	ActiveSellerMinDeals         uint64
	StartingSellerMinCreated     uint64
	StartingSellerMaxPublished   uint64
	DecisiveBuyerMinPurchases    uint64
	DecisiveBuyerMinChats        uint64
	DecisiveBuyerMinLinkedChats  uint64
	DecisiveBuyerMinPurchaseRate float64
	FindHunterMinViews           uint64
	FindHunterMinFavorites       uint64
	FindHunterMinRepeatRate      float64
	ResearcherMinViews           uint64
	ResearcherMinCategories      uint64
	ResearcherMaxChats           uint64
}

type AchievementThresholds struct {
	BalancedMinPurchases      uint64
	BalancedMinSales          uint64
	BalancedMaxDifferenceRate float64
	ThematicMinViews          uint64
	ThematicMinFavorites      uint64
	ThematicMinPurchases      uint64
	ThematicMinDominanceRate  float64
}

// RecommendationPriorities makes product ordering explicit and fingerprinted.
// A larger number wins; equal priorities are resolved by action code.
type RecommendationPriorities struct {
	FinishDraft       int
	ContinueDialog    int
	ImproveListing    int
	SimilarToPurchase int
	SaveSearch        int
	OpenFavorites     int
	CreateForStarter  int
	CreateForSeller   int
	OpenTopCategory   int
	NeutralFallback   int
}

// AchievementRuleConfig binds each catalogue item to a product category and
// priority. The executable match/text logic is keyed by Code; this configuration
// is digest-bound so category, grade ordering, and the global limit cannot drift
// under an unchanged rules identity.
type AchievementRuleConfig struct {
	Code     model.AchievementCode     `json:"code"`
	Category model.AchievementCategory `json:"category"`
	Priority int                       `json:"priority"`
}

type AchievementPolicy struct {
	MaxAwarded int                     `json:"maxAwarded"`
	Rules      []AchievementRuleConfig `json:"rules"`
}

// SharePolicy is an allow-list policy for the public DTO. Upstream data flags
// are necessary but never sufficient: generated text and category identifiers
// must also satisfy this policy.
type SharePolicy struct {
	Version                  string
	AllowTopCategory         bool
	RequireCategoryShareFlag bool
	MaxPublicTextRunes       int
	AllowedAchievementCodes  []model.AchievementCode
}

type Ruleset struct {
	Version                  string
	Algorithm                string
	Thresholds               BehaviorThresholds
	AchievementThresholds    AchievementThresholds
	AchievementPolicy        AchievementPolicy
	RecommendationPriorities RecommendationPriorities
	SharePolicy              SharePolicy
}
