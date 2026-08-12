package ruleset

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/year-recap/internal/recap/model"
)

var ErrInvalidRuleset = errors.New("invalid ruleset")

const currentRulesAlgorithm = "recap-v3.6-percentage-achievements-v1"

const (
	CurrentRulesVersion = "3.6.0"
	MaxAchievements     = 3
)

var semanticVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

type EligibilityPolicy struct {
	MinEvents uint64 `json:"minEvents"`
}

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
	SuccessfulSellerMinPublished    uint64
	SuccessfulSellerMinSaleRate     float64
	ConsistentPublisherMinPublished uint64
	ConsistentPublisherMinSaleRate  float64
	BalancedMinPurchases            uint64
	BalancedMinSales                uint64
	BalancedMaxDifferenceRate       float64
	ThematicMinViews                uint64
	ThematicMinFavorites            uint64
	ThematicMinPurchases            uint64
	ThematicMinDominanceRate        float64
}

// RecommendationPriorities makes product ordering explicit and fingerprinted.
// A larger number wins; equal priorities are resolved by action code.

type RecommendationThresholds struct {
	ImproveListingsMinActive uint64
}

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
	Eligibility              EligibilityPolicy
	Thresholds               BehaviorThresholds
	AchievementThresholds    AchievementThresholds
	AchievementPolicy        AchievementPolicy
	RecommendationThresholds RecommendationThresholds
	RecommendationPriorities RecommendationPriorities
	SharePolicy              SharePolicy
}

func (r Ruleset) Digest() string {
	copyRules := r
	copyRules.Version = strings.TrimSpace(copyRules.Version)
	copyRules.Algorithm = strings.TrimSpace(copyRules.Algorithm)
	copyRules.SharePolicy.Version = strings.TrimSpace(copyRules.SharePolicy.Version)
	copyRules.AchievementPolicy.Rules = append([]AchievementRuleConfig(nil), copyRules.AchievementPolicy.Rules...)
	sort.Slice(copyRules.AchievementPolicy.Rules, func(i, j int) bool {
		return copyRules.AchievementPolicy.Rules[i].Code < copyRules.AchievementPolicy.Rules[j].Code
	})
	copyRules.SharePolicy.AllowedAchievementCodes = append([]model.AchievementCode(nil), copyRules.SharePolicy.AllowedAchievementCodes...)
	sort.Slice(copyRules.SharePolicy.AllowedAchievementCodes, func(i, j int) bool {
		return copyRules.SharePolicy.AllowedAchievementCodes[i] < copyRules.SharePolicy.AllowedAchievementCodes[j]
	})
	data, err := json.Marshal(copyRules)
	if err != nil {
		panic(fmt.Sprintf("marshal validated ruleset: %v", err))
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (p SharePolicy) AchievementAllowed(code model.AchievementCode) bool {
	for _, allowed := range p.AllowedAchievementCodes {
		if code == allowed {
			return true
		}
	}
	return false
}
