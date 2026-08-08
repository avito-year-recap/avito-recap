package ruleset

import (
	"fmt"
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func (r Ruleset) Validate() error {
	r.Version = strings.TrimSpace(r.Version)
	r.Algorithm = strings.TrimSpace(r.Algorithm)
	if !semanticVersionPattern.MatchString(r.Version) {
		return fmt.Errorf("%w: version %q must be semantic", ErrInvalidRuleset, r.Version)
	}
	if r.Algorithm != currentRulesAlgorithm {
		return fmt.Errorf("%w: unsupported algorithm %q", ErrInvalidRuleset, r.Algorithm)
	}
	if r.Eligibility.MinEvents == 0 {
		return fmt.Errorf("%w: minimum event count must be positive", ErrInvalidRuleset)
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
	if policy.MaxAwarded < 1 || policy.MaxAwarded > MaxAchievements {
		return fmt.Errorf("%w: achievement award limit must be in [1,%d]", ErrInvalidRuleset, MaxAchievements)
	}
	knownAchievementCodes := model.AllAchievementCodes()
	if len(policy.Rules) != len(knownAchievementCodes) {
		return fmt.Errorf("%w: achievement policy must configure every catalogue item exactly once", ErrInvalidRuleset)
	}
	seenAchievementCodes := make(map[model.AchievementCode]struct{}, len(policy.Rules))
	for index, rule := range policy.Rules {
		if !model.IsKnownAchievementCode(rule.Code) {
			return fmt.Errorf("%w: achievement rule %d has unknown code %q", ErrInvalidRuleset, index, rule.Code)
		}
		if !model.IsKnownAchievementCategory(rule.Category) {
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

	if r.RecommendationThresholds.ImproveListingsMinActive == 0 {
		return fmt.Errorf("%w: recommendation count thresholds must be positive", ErrInvalidRuleset)
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
	if r.RecommendationPriorities.FinishDraft <= r.RecommendationPriorities.ContinueDialog ||
		r.RecommendationPriorities.ContinueDialog <= r.RecommendationPriorities.ImproveListing ||
		r.RecommendationPriorities.ImproveListing <= r.RecommendationPriorities.SimilarToPurchase ||
		r.RecommendationPriorities.SimilarToPurchase <= r.RecommendationPriorities.SaveSearch ||
		r.RecommendationPriorities.SaveSearch <= r.RecommendationPriorities.OpenFavorites ||
		r.RecommendationPriorities.OpenFavorites <= r.RecommendationPriorities.CreateForStarter ||
		r.RecommendationPriorities.CreateForStarter <= r.RecommendationPriorities.CreateForSeller ||
		r.RecommendationPriorities.CreateForSeller <= r.RecommendationPriorities.OpenTopCategory ||
		r.RecommendationPriorities.OpenTopCategory <= r.RecommendationPriorities.NeutralFallback {
		return fmt.Errorf("%w: recommendation priorities violate the executable-work-first policy", ErrInvalidRuleset)
	}
	p := r.SharePolicy
	if strings.TrimSpace(p.Version) == "" || p.MaxPublicTextRunes < 16 || len(p.AllowedAchievementCodes) == 0 {
		return fmt.Errorf("%w: invalid share policy", ErrInvalidRuleset)
	}
	seen := make(map[model.AchievementCode]struct{}, len(p.AllowedAchievementCodes))
	for _, code := range p.AllowedAchievementCodes {
		if !model.IsKnownAchievementCode(code) {
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
