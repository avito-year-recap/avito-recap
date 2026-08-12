package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/recap/validation/structural"
)

func TestAuditEveryConfiguredAchievementIsReachableThroughService(t *testing.T) {
	thematic := func(code, title string) model.Metrics {
		return model.Metrics{TotalViews: 30, CategoriesCount: 1, CategoryActivities: []model.CategoryActivity{{CategoryCode: code, Category: title, Views: 30}}}
	}
	witnesses := map[model.AchievementCode]model.Metrics{
		model.AchievementFirstSellingSteps:   {ListingsCreated: 3, ListingsPublished: 2},
		model.AchievementConsistentPublisher: {ListingsPublished: 10, SalesCompleted: 5},
		model.AchievementSuccessfulSeller:    {ListingsPublished: 10, SalesCompleted: 7},
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
		model.AchievementDecisiveStep:        thematic(analytics.CategoryRealEstate, "Недвижимость"),
	}
	for _, rule := range ruleset.DefaultRuleset().AchievementPolicy.Rules {
		metrics := finalizeAuditMetrics(witnesses[rule.Code])
		t.Run(string(rule.Code), func(t *testing.T) {
			if err := structural.ValidateMetricsForPeriod(metrics, testkit.Period()); err != nil {
				t.Fatalf("witness metrics are invalid: %v\n%+v", err, metrics)
			}
			recaps := &testkit.RecapStorage{}
			service, err := application.NewService(
				&testkit.ProfileStorage{Profile: testkit.Profile()},
				&testkit.AnalyticsStorage{Metrics: metrics},
				&testkit.ActionStateStorage{State: testkit.ActionableState()},
				recaps,
				application.WithClock(testkit.Clock),
				application.WithIDGenerator(sequenceIDGenerator(testkit.RecapID, testkit.ShareID)),
			)
			if err != nil {
				t.Fatal(err)
			}
			value, err := service.Generate(context.Background(), testkit.ProfileID, 2025)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if !hasAchievement(value.Achievements, rule.Code) {
				t.Fatalf("achievement %s was not selected; got %+v", rule.Code, value.Achievements)
			}
		})
	}
}

func finalizeAuditMetrics(metrics model.Metrics) model.Metrics {
	var categoryViews, categoryFavorites, categoryPurchases uint64
	for _, activity := range metrics.CategoryActivities {
		categoryViews += activity.Views
		categoryFavorites += activity.FavoritesAdded
		categoryPurchases += activity.PurchasesCompleted
	}
	if metrics.TotalViews < categoryViews {
		metrics.TotalViews = categoryViews
	}
	if metrics.FavoritesAdded < categoryFavorites {
		metrics.FavoritesAdded = categoryFavorites
	}
	if metrics.PurchasesCompleted < categoryPurchases {
		metrics.PurchasesCompleted = categoryPurchases
	}
	if metrics.CategoriesCount < uint64(len(metrics.CategoryActivities)) {
		metrics.CategoriesCount = uint64(len(metrics.CategoryActivities))
	}
	metrics.UniqueListings = metrics.TotalViews
	metrics.RepeatedViews = 0
	known := metrics.Searches + metrics.TotalViews + metrics.FavoritesAdded + metrics.ChatsStarted + metrics.ListingsCreated + metrics.ListingsPublished + metrics.PurchasesCompleted + metrics.SalesCompleted
	required := ruleset.DefaultRuleset().Eligibility.MinEvents
	if metrics.CategoriesCount > required {
		required = metrics.CategoriesCount
	}
	if known < required {
		metrics.Searches += required - known
		known = required
	}
	metrics.TotalEvents = known
	metrics.ActiveDays = 1
	metrics.MostActiveMonth = 1
	return analytics.EnrichMetrics(metrics)
}

func hasAchievement(values []model.Achievement, code model.AchievementCode) bool {
	for _, value := range values {
		if value.Code == code {
			return true
		}
	}
	return false
}

func sequenceIDGenerator(values ...uuid.UUID) application.IDGenerator {
	index := 0
	return func() (uuid.UUID, error) {
		value := values[index]
		index++
		return value, nil
	}
}
