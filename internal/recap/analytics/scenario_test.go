package analytics_test

import (
	"errors"
	"math"
	"testing"

	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
)

func TestMetricsFromScenario(t *testing.T) {
	scenario := testkit.SeedScenario{
		ProfileCode: "profile",
		Year:        2025,
		Activity: testkit.SeedActivity{
			Searches: 10, ListingViews: 100, UniqueListings: 70, FavoritesAdded: 20,
			ChatsStarted: 10, ChatsWithPurchase: 1, ListingsCreated: 2,
			ListingsPublished: 2, PurchasesCompleted: 1, SalesCompleted: 1, ActiveDays: 30,
		},
		Categories: []testkit.WeightedCategory{
			{Code: "furniture", Title: "Мебель", Weight: 40, Shareable: true},
			{Code: "electronics", Title: "Электроника", Weight: 60, Shareable: true},
		},
		Months: []testkit.WeightedMonth{{Month: 11, Weight: 70}, {Month: 4, Weight: 30}},
	}

	actual, err := testkit.MetricsFromScenario(scenario)
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if actual.TotalEvents != 146 || actual.RepeatedViews != 30 {
		t.Fatalf("unexpected aggregate metrics: %+v", actual)
	}
	if actual.TopCategoryCode != "electronics" || actual.TopCategory != "Электроника" || actual.TopCategoryViews != 60 {
		t.Fatalf("unexpected top category: %+v", actual)
	}
	if actual.MostActiveMonth != 11 || actual.CategoriesCount != 2 {
		t.Fatalf("unexpected grouped metrics: %+v", actual)
	}
	assertFloat(t, "RepeatRate", actual.RepeatRate, 0.3)
	assertFloat(t, "PurchaseRate", actual.PurchaseRate, 0.1)
}

func TestMetricsFromScenarioResolvesTiesDeterministically(t *testing.T) {
	scenario := testkit.SeedScenario{
		ProfileCode: "profile", Year: 2025,
		Activity:   testkit.SeedActivity{ListingViews: 10, UniqueListings: 10, ActiveDays: 1},
		Categories: []testkit.WeightedCategory{{Code: "z", Title: "Z", Weight: 50}, {Code: "a", Title: "A", Weight: 50}},
		Months:     []testkit.WeightedMonth{{Month: 9, Weight: 50}, {Month: 2, Weight: 50}},
	}
	actual, err := testkit.MetricsFromScenario(scenario)
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if actual.TopCategoryCode != "a" || actual.MostActiveMonth != 2 {
		t.Fatalf("tie resolution is not deterministic: %+v", actual)
	}
}

func TestMetricsFromScenarioAllowsNoCategoryActivity(t *testing.T) {
	actual, err := testkit.MetricsFromScenario(testkit.SeedScenario{
		ProfileCode: "newcomer", Year: 2025,
		Activity: testkit.SeedActivity{Searches: 5, ActiveDays: 3},
		Months:   []testkit.WeightedMonth{{Month: 1, Weight: 100}},
	})
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if actual.CategoriesCount != 0 || actual.TopCategoryCode != "" || actual.TopCategory != "" || actual.TopCategoryViews != 0 {
		t.Fatalf("unexpected category metrics: %+v", actual)
	}
}

func TestMetricsFromScenarioBuildsDetailedCategoryActivity(t *testing.T) {
	actual, err := testkit.MetricsFromScenario(testkit.SeedScenario{
		ProfileCode: "book-buyer", Year: 2025,
		Activity: testkit.SeedActivity{ListingViews: 30, UniqueListings: 25, FavoritesAdded: 8, PurchasesCompleted: 3, ActiveDays: 10},
		Categories: []testkit.WeightedCategory{{
			Code: analytics.CategoryBooks, Title: "Книги", Weight: 100, Shareable: true,
			Views: 30, FavoritesAdded: 8, PurchasesCompleted: 3,
		}},
		Months: []testkit.WeightedMonth{{Month: 6, Weight: 100}},
	})
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if len(actual.CategoryActivities) != 1 {
		t.Fatalf("category activities = %+v", actual.CategoryActivities)
	}
	activity := actual.CategoryActivities[0]
	if activity.CategoryCode != analytics.CategoryBooks || activity.Views != 30 || activity.FavoritesAdded != 8 || activity.PurchasesCompleted != 3 {
		t.Fatalf("unexpected detailed category activity: %+v", activity)
	}
	assertAchievementCodes(t, achievement.Build(actual), []model.AchievementCode{model.AchievementDealCloser, model.AchievementBookworm})
}

func TestMetricsFromScenarioRejectsInvalidSeeds(t *testing.T) {
	valid := testkit.SeedScenario{
		ProfileCode: "profile", Year: 2025,
		Activity:   testkit.SeedActivity{ListingViews: 10, UniqueListings: 10, ActiveDays: 1},
		Categories: []testkit.WeightedCategory{{Code: "a", Title: "A", Weight: 100}},
		Months:     []testkit.WeightedMonth{{Month: 1, Weight: 100}},
	}
	tests := []struct {
		name   string
		mutate func(*testkit.SeedScenario)
	}{
		{"missing profile", func(s *testkit.SeedScenario) { s.ProfileCode = "" }},
		{"missing year", func(s *testkit.SeedScenario) { s.Year = 0 }},
		{"unique exceed views", func(s *testkit.SeedScenario) { s.Activity.UniqueListings = 11 }},
		{"bad category weights", func(s *testkit.SeedScenario) { s.Categories[0].Weight = 99 }},
		{"duplicate category", func(s *testkit.SeedScenario) {
			s.Categories = append(s.Categories, s.Categories[0])
			s.Categories[0].Weight, s.Categories[1].Weight = 50, 50
		}},
		{"category views exceed aggregate", func(s *testkit.SeedScenario) { s.Categories[0].Views = 11 }},
		{"category purchases exceed aggregate", func(s *testkit.SeedScenario) { s.Categories[0].PurchasesCompleted = 1 }},
		{"bad month", func(s *testkit.SeedScenario) { s.Months[0].Month = 13 }},
		{"bad month weights", func(s *testkit.SeedScenario) { s.Months[0].Weight = 90 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario := valid
			scenario.Categories = append([]testkit.WeightedCategory(nil), valid.Categories...)
			scenario.Months = append([]testkit.WeightedMonth(nil), valid.Months...)
			test.mutate(&scenario)
			_, err := testkit.MetricsFromScenario(scenario)
			if !errors.Is(err, testkit.ErrInvalidScenario) {
				t.Fatalf("expected ErrInvalidScenario, got %v", err)
			}
		})
	}
}

func assertFloat(t *testing.T, name string, actual, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-12 {
		t.Fatalf("%s = %v, want %v", name, actual, expected)
	}
}

func assertAchievementCodes(t *testing.T, actual []model.Achievement, expected []model.AchievementCode) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("achievement count = %d, want %d: %+v", len(actual), len(expected), actual)
	}
	for i := range expected {
		if actual[i].Code != expected[i] {
			t.Fatalf("achievement[%d] = %s, want %s: %+v", i, actual[i].Code, expected[i], actual)
		}
	}
}
