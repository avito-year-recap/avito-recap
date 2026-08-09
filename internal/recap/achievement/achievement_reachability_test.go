package achievement

import (
	"testing"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestAuditEveryConfiguredAchievementIsReachable(t *testing.T) {
	thematic := func(code, title string) model.Metrics {
		return model.Metrics{
			TotalViews:         30,
			CategoriesCount:    1,
			CategoryActivities: []model.CategoryActivity{{CategoryCode: code, Category: title, Views: 30}},
		}
	}
	witnesses := map[model.AchievementCode]model.Metrics{
		model.AchievementFirstSellingSteps:   {ListingsCreated: 3, ListingsPublished: 2},
		model.AchievementConsistentPublisher: {ListingsPublished: 5, SalesCompleted: 1},
		model.AchievementSuccessfulSeller:    {ListingsPublished: 5, SalesCompleted: 5},
		model.AchievementDealCloser:          {PurchasesCompleted: 3},
		model.AchievementQuickDecision:       {PurchasesCompleted: 3, ChatsStarted: 5, ChatsWithPurchase: 3},
		model.AchievementBroadInterests:      {CategoriesCount: 6},
		model.AchievementAttentiveResearcher: {TotalViews: 150},
		model.AchievementMasterOfFavorites:   {FavoritesAdded: 20},
		model.AchievementAllRounder:          {ListingsPublished: 5, PurchasesCompleted: 5, SalesCompleted: 5},
		model.AchievementStyleIcon:           thematic(analytics.CategoryWomensFashion, "Женская одежда и аксессуары"),
		model.AchievementFashionableMan:      thematic(analytics.CategoryMensFashion, "Мужская одежда и аксессуары"),
		model.AchievementTraveler:            thematic(analytics.CategoryOutdoorTravel, "Туризм и путешествия"),
		model.AchievementForTheSoul:          thematic(analytics.CategoryGarden, "Дача и сад"),
		model.AchievementBookworm:            thematic(analytics.CategoryBooks, "Книги"),
		model.AchievementBeautyConnoisseur:   thematic(analytics.CategoryJewelry, "Украшения и ювелирные изделия"),
		model.AchievementInTheRhythmOfMusic:  thematic(analytics.CategoryMusic, "Музыкальные инструменты и аудио"),
		model.AchievementWorldOfPlay:         thematic(analytics.CategoryToysDolls, "Игрушки, куклы и коллекционирование"),
		model.AchievementMasterCraft:         thematic(analytics.CategoryTools, "Инструменты"),
		model.AchievementCaringOwner:         thematic(analytics.CategoryPets, "Товары для животных"),
		model.AchievementLittleDiscoveries:   thematic(analytics.CategoryKids, "Детские товары"),
	}
	configured := ruleset.DefaultRuleset().AchievementPolicy.Rules
	if len(witnesses) != len(configured) {
		t.Fatalf("witness count = %d, configured achievements = %d", len(witnesses), len(configured))
	}
	for _, rule := range configured {
		metrics, ok := witnesses[rule.Code]
		if !ok {
			t.Errorf("no reachability witness for configured achievement %s", rule.Code)
			continue
		}
		t.Run(string(rule.Code), func(t *testing.T) {
			result := Build(metrics)
			if _, found := findAchievement(result, rule.Code); !found {
				t.Fatalf("achievement %s was not selected; got %+v", rule.Code, result)
			}
		})
	}
}

func TestAuditPortfolioBoundaryRules(t *testing.T) {
	t.Run("seller dominant always includes seller achievement", func(t *testing.T) {
		for sales := uint64(1); sales <= 20; sales++ {
			for purchases := uint64(0); purchases < sales; purchases++ {
				result := Build(model.Metrics{SalesCompleted: sales, PurchasesCompleted: purchases})
				foundSeller := false
				for _, item := range result {
					if item.Category == model.AchievementCategorySelling {
						foundSeller = true
						break
					}
				}
				if !foundSeller {
					t.Fatalf("sales=%d purchases=%d produced no selling achievement: %+v", sales, purchases, result)
				}
			}
		}
	})
	t.Run("seller only gets exactly one", func(t *testing.T) {
		for sales := uint64(1); sales <= 20; sales++ {
			result := Build(model.Metrics{SalesCompleted: sales, ListingsPublished: 20})
			if len(result) != 1 || result[0].Category != model.AchievementCategorySelling {
				t.Fatalf("sales=%d: got %+v", sales, result)
			}
		}
	})
	t.Run("buyer only can get three purchase-backed themes", func(t *testing.T) {
		result := Build(model.Metrics{
			PurchasesCompleted: 9, CategoriesCount: 3,
			CategoryActivities: []model.CategoryActivity{
				{CategoryCode: analytics.CategoryBooks, Category: "Книги", PurchasesCompleted: 3},
				{CategoryCode: analytics.CategoryMusic, Category: "Музыка", PurchasesCompleted: 3},
				{CategoryCode: analytics.CategoryPets, Category: "Товары для животных", PurchasesCompleted: 3},
			},
		})
		if len(result) != 3 {
			t.Fatalf("got %d achievements: %+v", len(result), result)
		}
		for _, code := range []model.AchievementCode{model.AchievementBookworm, model.AchievementInTheRhythmOfMusic, model.AchievementCaringOwner} {
			if _, ok := findAchievement(result, code); !ok {
				t.Fatalf("missing %s in %+v", code, result)
			}
		}
	})
	t.Run("four buyer themes are capped at three", func(t *testing.T) {
		result := Build(model.Metrics{
			TotalViews: 120, PurchasesCompleted: 1, CategoriesCount: 4,
			CategoryActivities: []model.CategoryActivity{
				{CategoryCode: analytics.CategoryBooks, Category: "Книги", Views: 30},
				{CategoryCode: analytics.CategoryMusic, Category: "Музыка", Views: 30},
				{CategoryCode: analytics.CategoryPets, Category: "Животные", Views: 30},
				{CategoryCode: analytics.CategoryTools, Category: "Инструменты", Views: 30},
			},
		})
		if len(result) != 3 {
			t.Fatalf("got %d achievements: %+v", len(result), result)
		}
	})
	t.Run("balance boundary is inclusive only at fifty percent of larger side", func(t *testing.T) {
		thresholds := ruleset.DefaultRuleset().AchievementThresholds
		if !isBalancedSellerBuyer(model.Metrics{PurchasesCompleted: 5, SalesCompleted: 10}, thresholds) {
			t.Fatal("5 vs 10 should qualify")
		}
		if isBalancedSellerBuyer(model.Metrics{PurchasesCompleted: 5, SalesCompleted: 11}, thresholds) {
			t.Fatal("5 vs 11 should not qualify")
		}
	})
}

func findAchievement(values []model.Achievement, code model.AchievementCode) (model.Achievement, bool) {
	for _, value := range values {
		if value.Code == code {
			return value, true
		}
	}
	return model.Achievement{}, false
}
