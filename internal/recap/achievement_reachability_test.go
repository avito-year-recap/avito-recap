package recap

import (
	"context"
	"testing"
	"time"
)

func TestAuditEveryConfiguredAchievementIsReachableThroughService(t *testing.T) {
	thematic := func(code, title string) Metrics {
		return Metrics{
			TotalViews:         30,
			CategoriesCount:    1,
			CategoryActivities: []CategoryActivity{{CategoryCode: code, Category: title, Views: 30}},
		}
	}

	witnesses := map[AchievementCode]Metrics{
		AchievementFirstSellingSteps:   {ListingsCreated: 3, ListingsPublished: 2},
		AchievementConsistentPublisher: {ListingsPublished: 5, SalesCompleted: 1},
		AchievementSuccessfulSeller:    {ListingsPublished: 5, SalesCompleted: 5},
		AchievementDealCloser:          {PurchasesCompleted: 3},
		AchievementQuickDecision:       {PurchasesCompleted: 3, ChatsStarted: 5, ChatsWithPurchase: 3},
		AchievementBroadInterests:      {CategoriesCount: 6},
		AchievementAttentiveResearcher: {TotalViews: 150},
		AchievementMasterOfFavorites:   {FavoritesAdded: 20},
		AchievementAllRounder:          {ListingsPublished: 5, PurchasesCompleted: 5, SalesCompleted: 5},
		AchievementStyleIcon:           thematic(CategoryWomensFashion, "Женская одежда и аксессуары"),
		AchievementFashionableMan:      thematic(CategoryMensFashion, "Мужская одежда и аксессуары"),
		AchievementTraveler:            thematic(CategoryOutdoorTravel, "Туризм и путешествия"),
		AchievementForTheSoul:          thematic(CategoryGarden, "Дача и сад"),
		AchievementBookworm:            thematic(CategoryBooks, "Книги"),
		AchievementBeautyConnoisseur:   thematic(CategoryJewelry, "Украшения и ювелирные изделия"),
		AchievementInTheRhythmOfMusic:  thematic(CategoryMusic, "Музыкальные инструменты и аудио"),
		AchievementWorldOfPlay:         thematic(CategoryToysDolls, "Игрушки, куклы и коллекционирование"),
		AchievementMasterCraft:         thematic(CategoryTools, "Инструменты"),
		AchievementCaringOwner:         thematic(CategoryPets, "Товары для животных"),
		AchievementLittleDiscoveries:   thematic(CategoryKids, "Детские товары"),
	}

	configured := DefaultRuleset().AchievementPolicy.Rules
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
			metrics = finalizeAuditMetrics(metrics)
			if err := validateMetricsForPeriod(metrics, validPeriod()); err != nil {
				t.Fatalf("witness metrics are invalid: %v\n%+v", err, metrics)
			}

			profiles := &profileStorageStub{profile: validProfile()}
			analytics := &analyticsStorageStub{metrics: metrics}
			recaps := &recapStorageStub{}
			service := mustService(t, profiles, analytics, recaps,
				WithClock(func() time.Time { return fixedClock() }),
				WithIDGenerator(sequenceIDGenerator(testRecapID, testShareID)),
			)
			value, err := service.Generate(context.Background(), testProfileID, 2025)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if _, found := findAchievement(value.Achievements, rule.Code); !found {
				t.Fatalf("achievement %s was not selected; got %+v", rule.Code, value.Achievements)
			}
		})
	}
}

func finalizeAuditMetrics(m Metrics) Metrics {
	var categoryViews, categoryFavorites, categoryPurchases uint64
	for _, activity := range m.CategoryActivities {
		categoryViews += activity.Views
		categoryFavorites += activity.FavoritesAdded
		categoryPurchases += activity.PurchasesCompleted
	}
	if m.TotalViews < categoryViews {
		m.TotalViews = categoryViews
	}
	if m.FavoritesAdded < categoryFavorites {
		m.FavoritesAdded = categoryFavorites
	}
	if m.PurchasesCompleted < categoryPurchases {
		m.PurchasesCompleted = categoryPurchases
	}
	if m.CategoriesCount < uint64(len(m.CategoryActivities)) {
		m.CategoriesCount = uint64(len(m.CategoryActivities))
	}

	m.UniqueListings = m.TotalViews
	m.RepeatedViews = 0
	known := m.Searches + m.TotalViews + m.FavoritesAdded + m.ChatsStarted +
		m.ListingsCreated + m.ListingsPublished + m.PurchasesCompleted + m.SalesCompleted
	required := uint64(minEventsForRecap)
	if m.CategoriesCount > required {
		required = m.CategoriesCount
	}
	if known < required {
		m.Searches += required - known
		known = required
	}
	m.TotalEvents = known
	m.ActiveDays = 1
	m.MostActiveMonth = 1
	return m
}

func TestAuditPortfolioBoundaryRules(t *testing.T) {
	t.Run("seller dominant always includes seller achievement", func(t *testing.T) {
		for sales := uint64(1); sales <= 20; sales++ {
			for purchases := uint64(0); purchases < sales; purchases++ {
				result := BuildAchievements(Metrics{SalesCompleted: sales, PurchasesCompleted: purchases})
				foundSeller := false
				for _, a := range result {
					if a.Category == AchievementCategorySelling {
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
			result := BuildAchievements(Metrics{SalesCompleted: sales, ListingsPublished: 20})
			if len(result) != 1 || result[0].Category != AchievementCategorySelling {
				t.Fatalf("sales=%d: got %+v", sales, result)
			}
		}
	})

	t.Run("buyer only can get three purchase-backed themes", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			PurchasesCompleted: 9,
			CategoriesCount:    3,
			CategoryActivities: []CategoryActivity{
				{CategoryCode: CategoryBooks, Category: "Книги", PurchasesCompleted: 3},
				{CategoryCode: CategoryMusic, Category: "Музыка", PurchasesCompleted: 3},
				{CategoryCode: CategoryPets, Category: "Товары для животных", PurchasesCompleted: 3},
			},
		})
		if len(result) != 3 {
			t.Fatalf("got %d achievements: %+v", len(result), result)
		}
		for _, code := range []AchievementCode{AchievementBookworm, AchievementInTheRhythmOfMusic, AchievementCaringOwner} {
			if _, ok := findAchievement(result, code); !ok {
				t.Fatalf("missing %s in %+v", code, result)
			}
		}
	})

	t.Run("four buyer themes are capped at three", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			TotalViews: 120, PurchasesCompleted: 1, CategoriesCount: 4,
			CategoryActivities: []CategoryActivity{
				{CategoryCode: CategoryBooks, Category: "Книги", Views: 30},
				{CategoryCode: CategoryMusic, Category: "Музыка", Views: 30},
				{CategoryCode: CategoryPets, Category: "Животные", Views: 30},
				{CategoryCode: CategoryTools, Category: "Инструменты", Views: 30},
			},
		})
		if len(result) != 3 {
			t.Fatalf("got %d achievements: %+v", len(result), result)
		}
	})

	t.Run("balance boundary is inclusive only at fifty percent of larger side", func(t *testing.T) {
		th := DefaultRuleset().AchievementThresholds
		if !isBalancedSellerBuyer(Metrics{PurchasesCompleted: 5, SalesCompleted: 10}, th) {
			t.Fatal("5 vs 10 should qualify")
		}
		if isBalancedSellerBuyer(Metrics{PurchasesCompleted: 5, SalesCompleted: 11}, th) {
			t.Fatal("5 vs 11 should not qualify")
		}
	})
}
