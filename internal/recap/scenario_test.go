package recap

import (
	"errors"
	"testing"
)

func TestMetricsFromScenario(t *testing.T) {
	scenario := SeedScenario{
		ProfileCode: "profile",
		Year:        2025,
		RandomSeed:  42,
		Activity: SeedActivity{
			Searches:           10,
			ListingViews:       100,
			UniqueListings:     70,
			FavoritesAdded:     20,
			ChatsStarted:       10,
			ListingsCreated:    2,
			ListingsPublished:  2,
			PurchasesCompleted: 1,
			SalesCompleted:     1,
			ActiveDays:         30,
		},
		Categories: []WeightedCategory{
			{Code: "furniture", Title: "Мебель", Weight: 40, Shareable: true},
			{Code: "electronics", Title: "Электроника", Weight: 60, Shareable: true},
		},
		Months: []WeightedMonth{
			{Month: 11, Weight: 70},
			{Month: 4, Weight: 30},
		},
	}

	actual, err := MetricsFromScenario(scenario)
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if actual.TotalEvents != 146 {
		t.Fatalf("total events = %d, want 146", actual.TotalEvents)
	}
	if actual.RepeatedViews != 30 {
		t.Fatalf("repeated views = %d, want 30", actual.RepeatedViews)
	}
	if actual.TopCategoryCode != "electronics" || actual.TopCategory != "Электроника" || actual.TopCategoryViews != 60 {
		t.Fatalf("unexpected top category: %+v", actual)
	}
	if actual.MostActiveMonth != 11 || actual.CategoriesCount != 2 {
		t.Fatalf("unexpected grouped metrics: %+v", actual)
	}
	assertFloat(t, "FavoriteRate", actual.FavoriteRate, 0.2)
	assertFloat(t, "PurchaseRate", actual.PurchaseRate, 0.1)
}

func TestMetricsFromScenarioResolvesTiesDeterministically(t *testing.T) {
	scenario := SeedScenario{
		ProfileCode: "profile",
		Year:        2025,
		Activity: SeedActivity{
			ListingViews:   10,
			UniqueListings: 10,
			ActiveDays:     1,
		},
		Categories: []WeightedCategory{
			{Code: "z", Title: "Z", Weight: 50},
			{Code: "a", Title: "A", Weight: 50},
		},
		Months: []WeightedMonth{
			{Month: 9, Weight: 50},
			{Month: 2, Weight: 50},
		},
	}

	actual, err := MetricsFromScenario(scenario)
	if err != nil {
		t.Fatalf("MetricsFromScenario() error = %v", err)
	}
	if actual.TopCategoryCode != "a" || actual.MostActiveMonth != 2 {
		t.Fatalf("tie resolution is not deterministic: %+v", actual)
	}
}

func TestMetricsFromScenarioRejectsInvalidSeeds(t *testing.T) {
	valid := SeedScenario{
		ProfileCode: "profile",
		Year:        2025,
		Activity: SeedActivity{
			ListingViews:   10,
			UniqueListings: 10,
			ActiveDays:     1,
		},
		Categories: []WeightedCategory{{Code: "a", Title: "A", Weight: 100}},
		Months:     []WeightedMonth{{Month: 1, Weight: 100}},
	}

	tests := []struct {
		name   string
		mutate func(*SeedScenario)
	}{
		{name: "missing profile", mutate: func(s *SeedScenario) { s.ProfileCode = "" }},
		{name: "missing year", mutate: func(s *SeedScenario) { s.Year = 0 }},
		{name: "unique exceed views", mutate: func(s *SeedScenario) { s.Activity.UniqueListings = 11 }},
		{name: "bad category weights", mutate: func(s *SeedScenario) { s.Categories[0].Weight = 99 }},
		{name: "duplicate category", mutate: func(s *SeedScenario) {
			s.Categories = append(s.Categories, s.Categories[0])
			s.Categories[0].Weight = 50
			s.Categories[1].Weight = 50
		}},
		{name: "bad month", mutate: func(s *SeedScenario) { s.Months[0].Month = 13 }},
		{name: "bad month weights", mutate: func(s *SeedScenario) { s.Months[0].Weight = 90 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scenario := valid
			scenario.Categories = append([]WeightedCategory(nil), valid.Categories...)
			scenario.Months = append([]WeightedMonth(nil), valid.Months...)
			test.mutate(&scenario)
			_, err := MetricsFromScenario(scenario)
			if !errors.Is(err, ErrInvalidScenario) {
				t.Fatalf("expected ErrInvalidScenario, got %v", err)
			}
		})
	}
}
