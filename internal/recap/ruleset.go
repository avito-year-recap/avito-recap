package recap

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidRuleset = errors.New("invalid ruleset")

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

type Ruleset struct {
	Version    string
	Thresholds BehaviorThresholds
}

func DefaultRuleset() Ruleset {
	return Ruleset{
		Version: CurrentRulesVersion,
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
	}
}

func (r Ruleset) Validate() error {
	if strings.TrimSpace(r.Version) == "" {
		return fmt.Errorf("%w: version is required", ErrInvalidRuleset)
	}
	t := r.Thresholds
	if t.ActiveSellerMinListings == 0 || t.ActiveSellerMinDeals == 0 ||
		t.StartingSellerMinCreated == 0 || t.DecisiveBuyerMinPurchases == 0 ||
		t.DecisiveBuyerMinChats == 0 || t.DecisiveBuyerMinLinkedChats == 0 ||
		t.FindHunterMinViews == 0 || t.FindHunterMinFavorites == 0 ||
		t.ResearcherMinViews == 0 || t.ResearcherMinCategories == 0 {
		return fmt.Errorf("%w: count thresholds must be positive", ErrInvalidRuleset)
	}
	if t.DecisiveBuyerMinPurchaseRate <= 0 || t.DecisiveBuyerMinPurchaseRate > 1 ||
		t.FindHunterMinRepeatRate <= 0 || t.FindHunterMinRepeatRate > 1 {
		return fmt.Errorf("%w: rate thresholds must be in (0,1]", ErrInvalidRuleset)
	}
	return nil
}
