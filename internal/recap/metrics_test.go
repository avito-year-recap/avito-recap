package recap

import (
	"math"
	"testing"
)

func TestEnrichMetrics(t *testing.T) {
	input := Metrics{
		TotalViews:         200,
		RepeatedViews:      50,
		FavoritesAdded:     30,
		ChatsStarted:       10,
		ChatsWithPurchase:  2,
		ListingsCreated:    8,
		ListingsPublished:  6,
		PurchasesCompleted: 2,
		SalesCompleted:     3,
	}

	actual := EnrichMetrics(input)

	assertFloat(t, "FavoriteRate", actual.FavoriteRate, 0.15)
	assertFloat(t, "ChatRate", actual.ChatRate, 0.05)
	assertFloat(t, "RepeatRate", actual.RepeatRate, 0.25)
	assertFloat(t, "PublicationRate", actual.PublicationRate, 0.75)
	assertFloat(t, "SaleRate", actual.SaleRate, 0.5)
	assertFloat(t, "PurchaseRate", actual.PurchaseRate, 0.2)

	if actual.TotalViews != input.TotalViews {
		t.Fatalf("base metric changed: got %d, want %d", actual.TotalViews, input.TotalViews)
	}
}

func TestEnrichMetricsReplacesStaleRates(t *testing.T) {
	actual := EnrichMetrics(Metrics{
		TotalViews:     10,
		FavoritesAdded: 1,
		FavoriteRate:   0.99,
	})
	assertFloat(t, "FavoriteRate", actual.FavoriteRate, 0.1)
}

func TestEnrichMetricsZeroDenominators(t *testing.T) {
	actual := EnrichMetrics(Metrics{
		FavoritesAdded:     3,
		ChatsStarted:       0,
		RepeatedViews:      1,
		PurchasesCompleted: 1,
		SalesCompleted:     1,
	})

	if actual.FavoriteRate != 0 || actual.ChatRate != 0 || actual.RepeatRate != 0 ||
		actual.PublicationRate != 0 || actual.SaleRate != 0 || actual.PurchaseRate != 0 {
		t.Fatalf("expected all rates to be zero, got %+v", actual)
	}
}

func assertFloat(t *testing.T, name string, actual, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("%s = %f, want %f", name, actual, expected)
	}
}
