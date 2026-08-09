package model

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
