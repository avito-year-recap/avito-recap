package recap

import (
	"math"
	"testing"
)

func TestEnrichMetrics(t *testing.T) {
	input := Metrics{
		TotalViews:        200,
		RepeatedViews:     50,
		ChatsStarted:      10,
		ChatsWithPurchase: 2,
	}

	actual := EnrichMetrics(input)

	assertFloat(t, "RepeatRate", actual.RepeatRate, 0.25)
	assertFloat(t, "PurchaseRate", actual.PurchaseRate, 0.2)

	if actual.TotalViews != input.TotalViews {
		t.Fatalf("base metric changed: got %d, want %d", actual.TotalViews, input.TotalViews)
	}
}

func TestEnrichMetricsReplacesStaleRates(t *testing.T) {
	actual := EnrichMetrics(Metrics{
		TotalViews:        10,
		RepeatedViews:     1,
		ChatsStarted:      4,
		ChatsWithPurchase: 1,
		RepeatRate:        0.99,
		PurchaseRate:      0.99,
	})
	assertFloat(t, "RepeatRate", actual.RepeatRate, 0.1)
	assertFloat(t, "PurchaseRate", actual.PurchaseRate, 0.25)
}

func TestEnrichMetricsZeroDenominators(t *testing.T) {
	actual := EnrichMetrics(Metrics{})

	if actual.RepeatRate != 0 || actual.PurchaseRate != 0 {
		t.Fatalf("expected all rates to be zero, got %+v", actual)
	}
}

func TestEnrichMetricsNormalizesCategory(t *testing.T) {
	actual := EnrichMetrics(Metrics{
		TopCategoryCode: "  electronics\t",
		TopCategory:     "  Электроника\n",
	})
	if actual.TopCategoryCode != "electronics" || actual.TopCategory != "Электроника" {
		t.Fatalf("category was not normalized: %+v", actual)
	}
}

func assertFloat(t *testing.T, name string, actual, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("%s = %f, want %f", name, actual, expected)
	}
}
