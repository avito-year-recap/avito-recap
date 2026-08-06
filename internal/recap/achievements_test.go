package recap

import (
	"slices"
	"testing"
)

func TestBuildAchievementsForMVPProfiles(t *testing.T) {
	tests := []struct {
		name       string
		metrics    Metrics
		expected   []AchievementCode
		categories []AchievementCategory
	}{
		{
			name: "active buyer", metrics: Metrics{TotalViews: 180, FavoritesAdded: 28},
			expected:   []AchievementCode{AchievementAttentiveResearcher, AchievementMasterOfFavorites},
			categories: []AchievementCategory{AchievementCategoryDiscovery, AchievementCategoryCollection},
		},
		{
			name: "active seller", metrics: Metrics{ListingsPublished: 9, SalesCompleted: 6},
			expected:   []AchievementCode{AchievementSuccessfulSeller},
			categories: []AchievementCategory{AchievementCategorySelling},
		},
		{
			name: "researcher", metrics: Metrics{TotalViews: 260, CategoriesCount: 7},
			expected:   []AchievementCode{AchievementBroadInterests},
			categories: []AchievementCategory{AchievementCategoryDiscovery},
		},
		{
			name: "seller-dominant universal", metrics: Metrics{PurchasesCompleted: 1, SalesCompleted: 2, ListingsPublished: 4, ChatsStarted: 9},
			expected:   []AchievementCode{AchievementFirstSellingSteps},
			categories: []AchievementCategory{AchievementCategorySelling},
		},
		{
			name: "draft seller", metrics: Metrics{ListingsCreated: 7, ListingsPublished: 2},
			expected:   []AchievementCode{AchievementFirstSellingSteps},
			categories: []AchievementCategory{AchievementCategorySelling},
		},
		{
			name: "decisive buyer", metrics: Metrics{ChatsStarted: 15, ChatsWithPurchase: 4, PurchasesCompleted: 4},
			expected:   []AchievementCode{AchievementQuickDecision},
			categories: []AchievementCategory{AchievementCategoryBuying},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := BuildAchievements(test.metrics)
			assertAchievementCodes(t, actual, test.expected)
			for index, achievement := range actual {
				if achievement.Category != test.categories[index] {
					t.Fatalf("achievement %d category = %s, want %s", index, achievement.Category, test.categories[index])
				}
				if achievement.Priority == 0 || achievement.Title == "" || achievement.Reason == "" {
					t.Fatalf("incomplete achievement: %+v", achievement)
				}
			}
		})
	}
}

func TestAchievementTitlesUseNewCopy(t *testing.T) {
	cases := []struct {
		metrics Metrics
		code    AchievementCode
		title   string
	}{
		{Metrics{ListingsCreated: 3, ListingsPublished: 1}, AchievementFirstSellingSteps, "Начинающий бизнесмен"},
		{Metrics{ListingsPublished: 5, SalesCompleted: 1}, AchievementConsistentPublisher, "Маяк стабильности"},
		{Metrics{SalesCompleted: 5}, AchievementSuccessfulSeller, "Мастер переговоров"},
		{Metrics{PurchasesCompleted: 3, ChatsStarted: 5, ChatsWithPurchase: 3}, AchievementQuickDecision, "Молния"},
		{Metrics{FavoritesAdded: 20}, AchievementMasterOfFavorites, "Собиратель жемчужин"},
		{Metrics{CategoriesCount: 6}, AchievementBroadInterests, "Человек-оркестр"},
		{Metrics{PurchasesCompleted: 5, SalesCompleted: 5}, AchievementAllRounder, "Человек-швейцарский нож"},
		{Metrics{TotalViews: 150}, AchievementAttentiveResearcher, "Стратег"},
	}
	for _, test := range cases {
		result := BuildAchievements(test.metrics)
		achievement, ok := findAchievement(result, test.code)
		if !ok {
			t.Fatalf("achievement %s was not awarded: %+v", test.code, result)
		}
		if achievement.Title != test.title {
			t.Fatalf("achievement %s title = %q, want %q", test.code, achievement.Title, test.title)
		}
	}
}

func TestBuildAchievementsBuyingGradesRemainReachable(t *testing.T) {
	t.Run("quick decision outranks broad deal closer", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			PurchasesCompleted: 4,
			ChatsStarted:       15,
			ChatsWithPurchase:  4,
		})
		assertAchievementCodes(t, result, []AchievementCode{AchievementQuickDecision})
	})

	t.Run("deal closer remains available without conversion evidence", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			PurchasesCompleted: 4,
			ChatsStarted:       20,
			ChatsWithPurchase:  2,
		})
		assertAchievementCodes(t, result, []AchievementCode{AchievementDealCloser})
	})
}

func TestBalancedSellerBuyerGetsVersatilitySellingAndBuying(t *testing.T) {
	result := BuildAchievements(allAchievementMetrics())
	assertAchievementCodes(t, result, []AchievementCode{
		AchievementAllRounder,
		AchievementSuccessfulSeller,
		AchievementQuickDecision,
	})
	if len(result) != maxAchievements {
		t.Fatalf("achievement count = %d, want %d", len(result), maxAchievements)
	}
}

func TestBalancedSellerBuyerRequiresFiveOnBothSidesAndAtMostFiftyPercentDifference(t *testing.T) {
	thresholds := DefaultRuleset().AchievementThresholds
	if !isBalancedSellerBuyer(Metrics{PurchasesCompleted: 5, SalesCompleted: 10}, thresholds) {
		t.Fatal("exactly 50% difference should be balanced")
	}
	if isBalancedSellerBuyer(Metrics{PurchasesCompleted: 5, SalesCompleted: 11}, thresholds) {
		t.Fatal("difference above 50% should not be balanced")
	}
	if isBalancedSellerBuyer(Metrics{PurchasesCompleted: 4, SalesCompleted: 8}, thresholds) {
		t.Fatal("both sides must reach five")
	}
}

func TestSellerPortfoliosFollowProductRules(t *testing.T) {
	t.Run("seller only gets one strongest seller achievement", func(t *testing.T) {
		result := BuildAchievements(Metrics{ListingsPublished: 10, SalesCompleted: 7})
		assertAchievementCodes(t, result, []AchievementCode{AchievementSuccessfulSeller})
	})

	t.Run("seller dominant mixed profile prefers selling achievements", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			ListingsPublished: 10, SalesCompleted: 7, PurchasesCompleted: 2,
			CategoriesCount: 8,
		})
		assertAchievementCodes(t, result, []AchievementCode{
			AchievementSuccessfulSeller,
			AchievementConsistentPublisher,
			AchievementBroadInterests,
		})
	})
}

func TestBuyerOnlyCanReceiveThreeCategoryAchievements(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews: 105, PurchasesCompleted: 4, CategoriesCount: 3,
		CategoryActivities: []CategoryActivity{
			{CategoryCode: CategoryBooks, Category: "Книги", Views: 40, PurchasesCompleted: 2},
			{CategoryCode: CategoryMusic, Category: "Музыка", Views: 35, PurchasesCompleted: 1},
			{CategoryCode: CategoryPets, Category: "Товары для животных", Views: 30, PurchasesCompleted: 1},
		},
	})
	assertAchievementCodes(t, result, []AchievementCode{
		AchievementBookworm,
		AchievementInTheRhythmOfMusic,
		AchievementCaringOwner,
	})
}

func TestEveryThematicAchievementIsReachable(t *testing.T) {
	cases := []struct {
		code         AchievementCode
		categoryCode string
		category     string
		title        string
	}{
		{AchievementStyleIcon, CategoryWomensFashion, "Женская одежда", "Икона стиля"},
		{AchievementFashionableMan, CategoryMensFashion, "Мужская одежда", "Модник"},
		{AchievementTraveler, CategoryOutdoorTravel, "Туризм", "Путешественник"},
		{AchievementForTheSoul, CategoryGarden, "Дача и сад", "Для души"},
		{AchievementBookworm, CategoryBooks, "Книги", "Книжный червь"},
		{AchievementBeautyConnoisseur, CategoryJewelry, "Украшения", "Ценитель прекрасного"},
		{AchievementInTheRhythmOfMusic, CategoryMusic, "Музыка", "В ритме музыки"},
		{AchievementWorldOfPlay, CategoryToysDolls, "Игрушки", "Мир игры"},
		{AchievementMasterCraft, CategoryTools, "Инструменты", "Дело мастера"},
		{AchievementCaringOwner, CategoryPets, "Товары для животных", "Заботливый хозяин"},
		{AchievementLittleDiscoveries, CategoryKids, "Детские товары", "Для маленьких открытий"},
	}

	for _, test := range cases {
		t.Run(string(test.code), func(t *testing.T) {
			result := BuildAchievements(Metrics{
				TotalViews: 30, CategoriesCount: 1,
				CategoryActivities: []CategoryActivity{{
					CategoryCode: test.categoryCode, Category: test.category, Views: 30,
				}},
			})
			assertAchievementCodes(t, result, []AchievementCode{test.code})
			if result[0].Title != test.title || result[0].Shareable {
				t.Fatalf("unexpected thematic achievement: %+v", result[0])
			}
		})
	}
}

func TestThematicAchievementRequiresMeaningfulVolumeAndShare(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews: 200, CategoriesCount: 2,
		CategoryActivities: []CategoryActivity{
			{CategoryCode: CategoryBooks, Category: "Книги", Views: 30},
			{CategoryCode: CategoryElectronics, Category: "Электроника", Views: 170},
		},
	})
	if _, ok := findAchievement(result, AchievementBookworm); ok {
		t.Fatalf("low-share thematic achievement should not be awarded: %+v", result)
	}
}

func TestBuildAchievementsUsesCodeAsExplicitTieBreak(t *testing.T) {
	ruleset := DefaultRuleset()
	for index := range ruleset.AchievementPolicy.Rules {
		rule := &ruleset.AchievementPolicy.Rules[index]
		switch rule.Code {
		case AchievementBroadInterests, AchievementAttentiveResearcher:
			rule.Priority = 98
		}
	}
	if err := ruleset.Validate(); err != nil {
		t.Fatal(err)
	}

	result := ruleset.BuildAchievements(Metrics{TotalViews: 260, CategoriesCount: 7})
	assertAchievementCodes(t, result, []AchievementCode{AchievementAttentiveResearcher})
}

func TestBuildAchievementsDoesNotDependOnPolicySliceOrder(t *testing.T) {
	ruleset := DefaultRuleset()
	forward := ruleset.BuildAchievements(allAchievementMetrics())
	forwardDigest := ruleset.Digest()
	slices.Reverse(ruleset.AchievementPolicy.Rules)
	backward := ruleset.BuildAchievements(allAchievementMetrics())
	if !equalAchievements(forward, backward) {
		t.Fatalf("policy order changed awards:\nforward: %+v\nreverse: %+v", forward, backward)
	}
	if ruleset.Digest() != forwardDigest {
		t.Fatal("semantically irrelevant policy order changed rules digest")
	}
}

func TestFullRecapProjectionContainsOnlySelectedThreeAchievements(t *testing.T) {
	ruleset := DefaultRuleset()
	metrics := EnrichMetrics(allAchievementMetrics())
	state := validActionableState()
	behavior := ruleset.DetectBehavior(metrics)
	achievements := ruleset.BuildAchievements(metrics)
	action := ruleset.BuildNextAction(metrics, state, behavior)
	cards := BuildCardsWithRuleset(ruleset, validProfile(), 2025, testShareID, metrics, behavior, achievements, action)

	card := findCard(t, cards, CardAchievement)
	payload := card.Payload.(AchievementPayload)
	if len(payload.Codes) != maxAchievements {
		t.Fatalf("achievement card contains %d codes, want %d", len(payload.Codes), maxAchievements)
	}
	for index, code := range payload.Codes {
		if code != achievements[index].Code {
			t.Fatalf("card code %d = %s, want %s", index, code, achievements[index].Code)
		}
	}

	value := Recap{
		ID: testRecapID, ShareID: testShareID, Profile: validProfile(), Year: 2025,
		Period: validPeriod(), RulesVersion: ruleset.Version, RulesDigest: ruleset.Digest(),
		Metrics: metrics, ActionableState: state, Behavior: behavior,
		Achievements: achievements, Cards: cards, NextAction: action, GeneratedAt: fixedClock(),
	}
	if err := validateRecapAgainstRuleset(value, ruleset, fixedClock()); err != nil {
		t.Fatalf("full recap with selected awards is invalid: %v", err)
	}
}

func TestBuildAchievementsBelowThresholds(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews: 149, FavoritesAdded: 19, ListingsCreated: 2,
		ListingsPublished: 2, CategoriesCount: 5,
	})
	if len(result) != 0 {
		t.Fatalf("expected no achievements, got %+v", result)
	}
}

func allAchievementMetrics() Metrics {
	return Metrics{
		TotalEvents: 392, TotalViews: 300, UniqueListings: 250, RepeatedViews: 50,
		FavoritesAdded: 40, ChatsStarted: 20, ChatsWithPurchase: 5,
		PurchasesCompleted: 5, ListingsCreated: 10, ListingsPublished: 10,
		SalesCompleted: 7, ActiveDays: 100, CategoriesCount: 8, MostActiveMonth: 1,
	}
}

func findAchievement(values []Achievement, code AchievementCode) (Achievement, bool) {
	for _, value := range values {
		if value.Code == code {
			return value, true
		}
	}
	return Achievement{}, false
}

func assertAchievementCodes(t *testing.T, actual []Achievement, expected []AchievementCode) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("achievement count = %d, want %d: %+v", len(actual), len(expected), actual)
	}
	for index, code := range expected {
		if actual[index].Code != code {
			t.Fatalf("achievement %d = %s, want %s", index, actual[index].Code, code)
		}
	}
}
