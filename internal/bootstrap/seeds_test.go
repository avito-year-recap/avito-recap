package bootstrap

import (
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/validation/structural"
)

// TestEventsFromScenarioAggregateToDeclaredActivity checks the event-sourced
// path end to end without a database: eventsFromScenario expands a seed
// scenario into raw events, and re-aggregating those events must reproduce
// the scenario's own declared activity numbers exactly. This is what
// AnalyticsStorage.CalculateMetrics does for real on a cache miss, just
// against events read from ClickHouse instead of generated in memory.
func TestEventsFromScenarioAggregateToDeclaredActivity(t *testing.T) {
	scenarios, err := readJSON[[]scenario]("../../seeds/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range scenarios {
		t.Run(item.ProfileCode, func(t *testing.T) {
			events, err := eventsFromScenario(uuid.New(), item)
			if err != nil {
				t.Fatal(err)
			}
			if len(events) == 0 {
				t.Fatal("no events generated")
			}

			got := analytics.AggregateEvents(events)
			if err := structural.ValidateMetrics(got); err != nil {
				t.Fatalf("aggregated metrics are invalid: %v", err)
			}

			wantTotal, err := analytics.SumUint64(
				item.Activity.Searches,
				item.Activity.ListingViews,
				item.Activity.FavoritesAdded,
				item.Activity.ChatsStarted,
				item.Activity.ListingsCreated,
				item.Activity.ListingsPublished,
				item.Activity.PurchasesCompleted,
				item.Activity.SalesCompleted,
			)
			if err != nil {
				t.Fatal(err)
			}

			checks := []struct {
				name string
				got  uint64
				want uint64
			}{
				{"TotalEvents", got.TotalEvents, wantTotal},
				{"Searches", got.Searches, item.Activity.Searches},
				{"TotalViews", got.TotalViews, item.Activity.ListingViews},
				{"UniqueListings", got.UniqueListings, item.Activity.UniqueListings},
				{"RepeatedViews", got.RepeatedViews, item.Activity.ListingViews - item.Activity.UniqueListings},
				{"FavoritesAdded", got.FavoritesAdded, item.Activity.FavoritesAdded},
				{"ChatsStarted", got.ChatsStarted, item.Activity.ChatsStarted},
				{"ChatsWithPurchase", got.ChatsWithPurchase, item.Activity.ChatsWithPurchase},
				{"ListingsCreated", got.ListingsCreated, item.Activity.ListingsCreated},
				{"ListingsPublished", got.ListingsPublished, item.Activity.ListingsPublished},
				{"PurchasesCompleted", got.PurchasesCompleted, item.Activity.PurchasesCompleted},
				{"SalesCompleted", got.SalesCompleted, item.Activity.SalesCompleted},
				{"ActiveDays", got.ActiveDays, item.Activity.ActiveDays},
			}
			for _, check := range checks {
				if check.got != check.want {
					t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
				}
			}

			if len(item.Categories) > 0 && len(got.CategoryActivities) == 0 {
				t.Error("expected at least one category activity, got none")
			}
			if got.TopCategoryCode == "" && item.Activity.ListingViews > 0 {
				t.Error("expected a top category to be selected")
			}
			if got.MostActiveMonth < 1 || got.MostActiveMonth > 12 {
				t.Errorf("MostActiveMonth = %d, want 1..12", got.MostActiveMonth)
			}
		})
	}
}
