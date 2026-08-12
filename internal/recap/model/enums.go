package model

func IsKnownBehaviorCode(code BehaviorCode) bool {
	switch code {
	case BehaviorActiveSeller, BehaviorStartingSeller, BehaviorDecisiveBuyer, BehaviorFindHunter, BehaviorResearcher, BehaviorUniversal:
		return true
	default:
		return false
	}
}

func IsKnownAchievementCategory(category AchievementCategory) bool {
	switch category {
	case AchievementCategorySelling, AchievementCategoryBuying, AchievementCategoryDiscovery, AchievementCategoryCollection, AchievementCategoryVersatility, AchievementCategoryInterest:
		return true
	default:
		return false
	}
}

func AllAchievementCodes() []AchievementCode {
	return []AchievementCode{
		AchievementSuccessfulSeller, AchievementConsistentPublisher, AchievementAttentiveResearcher,
		AchievementMasterOfFavorites, AchievementBroadInterests, AchievementAllRounder,
		AchievementFirstSellingSteps, AchievementDealCloser, AchievementQuickDecision,
		AchievementStyleIcon, AchievementFashionableMan, AchievementTraveler, AchievementForTheSoul,
		AchievementBookworm, AchievementBeautyConnoisseur, AchievementInTheRhythmOfMusic,
		AchievementWorldOfPlay, AchievementMasterCraft, AchievementCaringOwner, AchievementLittleDiscoveries,
		AchievementDecisiveStep,
	}
}

func IsKnownAchievementCode(code AchievementCode) bool {
	for _, known := range AllAchievementCodes() {
		if code == known {
			return true
		}
	}
	return false
}

func IsKnownActionCode(code ActionCode) bool {
	switch code {
	case ActionFinishDraft, ActionOpenFavorites, ActionImproveListings, ActionContinueDialogs,
		ActionOpenTopCategory, ActionCreateFirstListing, ActionCreateListing, ActionSaveSearch,
		ActionViewSimilarListings, ActionExploreRecommendations:
		return true
	default:
		return false
	}
}

func IsKnownCardType(cardType CardType) bool {
	switch cardType {
	case CardIntro, CardYearActivity, CardTopCategory, CardActiveMonth, CardBehavior,
		CardAchievement, CardMissedOpportunity, CardNextAction, CardShare:
		return true
	default:
		return false
	}
}
