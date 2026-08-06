package recap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
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
	Code     AchievementCode     `json:"code"`
	Category AchievementCategory `json:"category"`
	Priority int                 `json:"priority"`
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
	AllowedAchievementCodes  []AchievementCode
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

func DefaultRuleset() Ruleset {
	return Ruleset{
		Version:   CurrentRulesVersion,
		Algorithm: currentRulesAlgorithm,
		Thresholds: BehaviorThresholds{
			ActiveSellerMinListings:      5,
			ActiveSellerMinDeals:         3,
			StartingSellerMinCreated:     3,
			StartingSellerMaxPublished:   2,
			DecisiveBuyerMinPurchases:    3,
			DecisiveBuyerMinChats:        5,
			DecisiveBuyerMinLinkedChats:  3,
			DecisiveBuyerMinPurchaseRate: 0.20,
			FindHunterMinViews:           20,
			FindHunterMinFavorites:       3,
			FindHunterMinRepeatRate:      0.20,
			ResearcherMinViews:           100,
			ResearcherMinCategories:      5,
			ResearcherMaxChats:           4,
		},
		AchievementThresholds: AchievementThresholds{
			BalancedMinPurchases:      5,
			BalancedMinSales:          5,
			BalancedMaxDifferenceRate: 0.50,
			ThematicMinViews:          30,
			ThematicMinFavorites:      8,
			ThematicMinPurchases:      3,
			ThematicMinDominanceRate:  0.20,
		},
		AchievementPolicy: AchievementPolicy{
			MaxAwarded: maxAchievements,
			Rules: []AchievementRuleConfig{
				{Code: AchievementAllRounder, Category: AchievementCategoryVersatility, Priority: 130},
				{Code: AchievementSuccessfulSeller, Category: AchievementCategorySelling, Priority: 120},
				{Code: AchievementQuickDecision, Category: AchievementCategoryBuying, Priority: 115},
				{Code: AchievementDealCloser, Category: AchievementCategoryBuying, Priority: 110},
				{Code: AchievementConsistentPublisher, Category: AchievementCategorySelling, Priority: 100},
				{Code: AchievementBroadInterests, Category: AchievementCategoryDiscovery, Priority: 98},
				{Code: AchievementFirstSellingSteps, Category: AchievementCategorySelling, Priority: 96},
				{Code: AchievementAttentiveResearcher, Category: AchievementCategoryDiscovery, Priority: 90},
				{Code: AchievementMasterOfFavorites, Category: AchievementCategoryCollection, Priority: 80},
				{Code: AchievementStyleIcon, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementFashionableMan, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementTraveler, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementForTheSoul, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementBookworm, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementBeautyConnoisseur, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementInTheRhythmOfMusic, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementWorldOfPlay, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementMasterCraft, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementCaringOwner, Category: AchievementCategoryInterest, Priority: 70},
				{Code: AchievementLittleDiscoveries, Category: AchievementCategoryInterest, Priority: 70},
			},
		},
		RecommendationPriorities: RecommendationPriorities{
			FinishDraft: 1000, ContinueDialog: 900, ImproveListing: 800,
			SimilarToPurchase: 750, SaveSearch: 700, OpenFavorites: 650,
			CreateForStarter: 520, CreateForSeller: 500, OpenTopCategory: 400,
			NeutralFallback: 0,
		},
		SharePolicy: SharePolicy{
			Version:                  "privacy-v2",
			AllowTopCategory:         true,
			RequireCategoryShareFlag: true,
			MaxPublicTextRunes:       80,
			AllowedAchievementCodes: []AchievementCode{
				AchievementSuccessfulSeller, AchievementConsistentPublisher,
				AchievementAttentiveResearcher, AchievementMasterOfFavorites,
				AchievementBroadInterests, AchievementAllRounder,
				AchievementFirstSellingSteps, AchievementDealCloser,
				AchievementQuickDecision,
			},
		},
	}
}

func (r Ruleset) Validate() error {
	r.Version = strings.TrimSpace(r.Version)
	r.Algorithm = strings.TrimSpace(r.Algorithm)
	if !semanticVersionPattern.MatchString(r.Version) {
		return fmt.Errorf("%w: version %q must be semantic", ErrInvalidRuleset, r.Version)
	}
	if r.Algorithm != currentRulesAlgorithm {
		return fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidRuleset, r.Algorithm)
	}
	t := r.Thresholds
	if t.ActiveSellerMinListings == 0 || t.ActiveSellerMinDeals == 0 ||
		t.StartingSellerMinCreated == 0 || t.DecisiveBuyerMinPurchases == 0 ||
		t.DecisiveBuyerMinChats == 0 || t.DecisiveBuyerMinLinkedChats == 0 ||
		t.FindHunterMinViews == 0 || t.FindHunterMinFavorites == 0 ||
		t.ResearcherMinViews == 0 || t.ResearcherMinCategories == 0 {
		return fmt.Errorf("%w: count thresholds must be positive", ErrInvalidRuleset)
	}
	if t.StartingSellerMaxPublished >= t.StartingSellerMinCreated {
		return fmt.Errorf("%w: starting-seller publication maximum must be below creation minimum", ErrInvalidRuleset)
	}
	if t.DecisiveBuyerMinLinkedChats > t.DecisiveBuyerMinChats {
		return fmt.Errorf("%w: linked-chat minimum exceeds total-chat minimum", ErrInvalidRuleset)
	}
	if t.DecisiveBuyerMinPurchaseRate <= 0 || t.DecisiveBuyerMinPurchaseRate > 1 ||
		t.FindHunterMinRepeatRate <= 0 || t.FindHunterMinRepeatRate > 1 {
		return fmt.Errorf("%w: rate thresholds must be in (0,1]", ErrInvalidRuleset)
	}
	a := r.AchievementThresholds
	if a.BalancedMinPurchases == 0 || a.BalancedMinSales == 0 ||
		a.ThematicMinViews == 0 || a.ThematicMinFavorites == 0 || a.ThematicMinPurchases == 0 {
		return fmt.Errorf("%w: achievement count thresholds must be positive", ErrInvalidRuleset)
	}
	if a.BalancedMaxDifferenceRate < 0 || a.BalancedMaxDifferenceRate > 1 ||
		a.ThematicMinDominanceRate <= 0 || a.ThematicMinDominanceRate > 1 {
		return fmt.Errorf("%w: achievement rate thresholds are outside valid ranges", ErrInvalidRuleset)
	}
	policy := r.AchievementPolicy
	if policy.MaxAwarded < 1 || policy.MaxAwarded > maxAchievements {
		return fmt.Errorf("%w: achievement award limit must be in [1,%d]", ErrInvalidRuleset, maxAchievements)
	}
	definitions := achievementDefinitions()
	if len(policy.Rules) != len(definitions) {
		return fmt.Errorf("%w: achievement policy must configure every catalogue item exactly once", ErrInvalidRuleset)
	}
	seenAchievementCodes := make(map[AchievementCode]struct{}, len(policy.Rules))
	for index, rule := range policy.Rules {
		if _, ok := definitions[rule.Code]; !ok {
			return fmt.Errorf("%w: achievement rule %d has unknown code %q", ErrInvalidRuleset, index, rule.Code)
		}
		if !isKnownAchievementCategory(rule.Category) {
			return fmt.Errorf("%w: achievement %q has unknown category %q", ErrInvalidRuleset, rule.Code, rule.Category)
		}
		if rule.Priority <= 0 {
			return fmt.Errorf("%w: achievement %q priority must be positive", ErrInvalidRuleset, rule.Code)
		}
		if _, exists := seenAchievementCodes[rule.Code]; exists {
			return fmt.Errorf("%w: duplicate achievement configuration %q", ErrInvalidRuleset, rule.Code)
		}
		seenAchievementCodes[rule.Code] = struct{}{}
	}

	priorities := []int{
		r.RecommendationPriorities.FinishDraft, r.RecommendationPriorities.ContinueDialog,
		r.RecommendationPriorities.ImproveListing, r.RecommendationPriorities.SimilarToPurchase,
		r.RecommendationPriorities.SaveSearch, r.RecommendationPriorities.OpenFavorites,
		r.RecommendationPriorities.CreateForStarter, r.RecommendationPriorities.CreateForSeller,
		r.RecommendationPriorities.OpenTopCategory,
	}
	for _, priority := range priorities {
		if priority <= r.RecommendationPriorities.NeutralFallback {
			return fmt.Errorf("%w: actionable recommendation priorities must exceed fallback", ErrInvalidRuleset)
		}
	}
	if !(r.RecommendationPriorities.FinishDraft > r.RecommendationPriorities.ContinueDialog &&
		r.RecommendationPriorities.ContinueDialog > r.RecommendationPriorities.ImproveListing &&
		r.RecommendationPriorities.ImproveListing > r.RecommendationPriorities.SimilarToPurchase &&
		r.RecommendationPriorities.SimilarToPurchase > r.RecommendationPriorities.SaveSearch &&
		r.RecommendationPriorities.SaveSearch > r.RecommendationPriorities.OpenFavorites &&
		r.RecommendationPriorities.OpenFavorites > r.RecommendationPriorities.CreateForStarter &&
		r.RecommendationPriorities.CreateForStarter > r.RecommendationPriorities.CreateForSeller &&
		r.RecommendationPriorities.CreateForSeller > r.RecommendationPriorities.OpenTopCategory &&
		r.RecommendationPriorities.OpenTopCategory > r.RecommendationPriorities.NeutralFallback) {
		return fmt.Errorf("%w: recommendation priorities violate the executable-work-first policy", ErrInvalidRuleset)
	}
	p := r.SharePolicy
	if strings.TrimSpace(p.Version) == "" || p.MaxPublicTextRunes < 16 || len(p.AllowedAchievementCodes) == 0 {
		return fmt.Errorf("%w: invalid share policy", ErrInvalidRuleset)
	}
	seen := make(map[AchievementCode]struct{}, len(p.AllowedAchievementCodes))
	for _, code := range p.AllowedAchievementCodes {
		if !isKnownAchievementCode(code) {
			return fmt.Errorf("%w: share policy contains unknown achievement %q", ErrInvalidRuleset, code)
		}
		if _, ok := seen[code]; ok {
			return fmt.Errorf("%w: duplicate shareable achievement %q", ErrInvalidRuleset, code)
		}
		seen[code] = struct{}{}
	}
	return nil
}

// Digest binds the stored result to the complete configurable rules contract,
// not merely to a human-readable version label.
func (r Ruleset) Digest() string {
	copyRules := r
	copyRules.Version = strings.TrimSpace(copyRules.Version)
	copyRules.Algorithm = strings.TrimSpace(copyRules.Algorithm)
	copyRules.SharePolicy.Version = strings.TrimSpace(copyRules.SharePolicy.Version)
	copyRules.AchievementPolicy.Rules = append([]AchievementRuleConfig(nil), copyRules.AchievementPolicy.Rules...)
	sort.Slice(copyRules.AchievementPolicy.Rules, func(i, j int) bool {
		return copyRules.AchievementPolicy.Rules[i].Code < copyRules.AchievementPolicy.Rules[j].Code
	})
	copyRules.SharePolicy.AllowedAchievementCodes = append([]AchievementCode(nil), copyRules.SharePolicy.AllowedAchievementCodes...)
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

func (p SharePolicy) achievementAllowed(code AchievementCode) bool {
	for _, allowed := range p.AllowedAchievementCodes {
		if code == allowed {
			return true
		}
	}
	return false
}
