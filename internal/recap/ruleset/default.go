package ruleset

import "github.com/year-recap/internal/recap/model"

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
			MaxAwarded: MaxAchievements,
			Rules: []AchievementRuleConfig{
				{Code: model.AchievementAllRounder, Category: model.AchievementCategoryVersatility, Priority: 130},
				{Code: model.AchievementSuccessfulSeller, Category: model.AchievementCategorySelling, Priority: 120},
				{Code: model.AchievementQuickDecision, Category: model.AchievementCategoryBuying, Priority: 115},
				{Code: model.AchievementDealCloser, Category: model.AchievementCategoryBuying, Priority: 110},
				{Code: model.AchievementConsistentPublisher, Category: model.AchievementCategorySelling, Priority: 100},
				{Code: model.AchievementBroadInterests, Category: model.AchievementCategoryDiscovery, Priority: 98},
				{Code: model.AchievementFirstSellingSteps, Category: model.AchievementCategorySelling, Priority: 96},
				{Code: model.AchievementAttentiveResearcher, Category: model.AchievementCategoryDiscovery, Priority: 90},
				{Code: model.AchievementMasterOfFavorites, Category: model.AchievementCategoryCollection, Priority: 80},
				{Code: model.AchievementStyleIcon, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementFashionableMan, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementTraveler, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementForTheSoul, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementBookworm, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementBeautyConnoisseur, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementInTheRhythmOfMusic, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementWorldOfPlay, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementMasterCraft, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementCaringOwner, Category: model.AchievementCategoryInterest, Priority: 70},
				{Code: model.AchievementLittleDiscoveries, Category: model.AchievementCategoryInterest, Priority: 70},
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
			AllowedAchievementCodes: []model.AchievementCode{
				model.AchievementSuccessfulSeller, model.AchievementConsistentPublisher,
				model.AchievementAttentiveResearcher, model.AchievementMasterOfFavorites,
				model.AchievementBroadInterests, model.AchievementAllRounder,
				model.AchievementFirstSellingSteps, model.AchievementDealCloser,
				model.AchievementQuickDecision,
			},
		},
	}
}
