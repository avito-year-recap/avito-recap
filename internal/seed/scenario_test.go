package seed_test

import (
	"errors"
	"testing"

	"github.com/year-recap/internal/seed"
)

func TestMetricsFromScenarioBuildsDeterministicAggregates(t *testing.T) {
	scenario := seed.Scenario{
		ProfileCode: "profile",
		Year:        2025,
		Activity: seed.Activity{
			Searches:           10,
			ListingViews:       100,
			UniqueListings:     70,
			FavoritesAdded:     20,
			ChatsStarted:       10,
			ChatsWithPurchase:  1,
			ListingsCreated:    2,
			ListingsPublished:  2,
			PurchasesCompleted: 1,
			SalesCompleted:     1,
			ActiveDays:         30,
		},
		Categories: []seed.WeightedCategory{
			{Code: "furniture", Title: "Мебель", Weight: 40, Shareable: true},
			{Code: "electronics", Title: "Электроника", Weight: 60, Shareable: true},
		},
		Months: []seed.WeightedMonth{
			{Month: 11, Weight: 70},
			{Month: 4, Weight: 30},
		},
	}

	actual, err := seed.MetricsFromScenario(scenario)
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if actual.TotalEvents != 146 || actual.RepeatedViews != 30 {
		t.Fatalf("unexpected aggregate metrics: %+v", actual)
	}
	if actual.TopCategoryCode != "electronics" ||
		actual.TopCategory != "Электроника" ||
		actual.TopCategoryViews != 60 {
		t.Fatalf("unexpected top category: %+v", actual)
	}
	if actual.MostActiveMonth != 11 || actual.CategoriesCount != 2 {
		t.Fatalf("unexpected period highlights: %+v", actual)
	}
}

func TestMetricsFromScenarioRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		scenario seed.Scenario
	}{
		{
			name: "missing profile code",
			scenario: seed.Scenario{
				Year: 2025,
				Activity: seed.Activity{
					ListingViews: 1, UniqueListings: 1, ActiveDays: 1,
				},
				Months: []seed.WeightedMonth{{Month: 1, Weight: 100}},
			},
		},
		{
			name: "unique listings exceed views",
			scenario: seed.Scenario{
				ProfileCode: "profile",
				Year:        2025,
				Activity: seed.Activity{
					ListingViews: 1, UniqueListings: 2, ActiveDays: 1,
				},
				Months: []seed.WeightedMonth{{Month: 1, Weight: 100}},
			},
		},
		{
			name: "category weights do not sum to 100",
			scenario: seed.Scenario{
				ProfileCode: "profile",
				Year:        2025,
				Activity: seed.Activity{
					ListingViews: 10, UniqueListings: 10, ActiveDays: 1,
				},
				Categories: []seed.WeightedCategory{
					{Code: "auto", Title: "Авто", Weight: 40, Shareable: true},
				},
				Months: []seed.WeightedMonth{{Month: 1, Weight: 100}},
			},
		},
		{
			name: "invalid month",
			scenario: seed.Scenario{
				ProfileCode: "profile",
				Year:        2025,
				Activity: seed.Activity{
					ListingViews: 1, UniqueListings: 1, ActiveDays: 1,
				},
				Months: []seed.WeightedMonth{{Month: 13, Weight: 100}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := seed.MetricsFromScenario(test.scenario)
			if !errors.Is(err, seed.ErrInvalidScenario) {
				t.Fatalf("error = %v, want invalid scenario", err)
			}
		})
	}
}
