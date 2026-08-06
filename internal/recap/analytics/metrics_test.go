package analytics

import (
	"math"
	"testing"

	"github.com/year-recap/internal/recap/model"
)

func TestEnrichMetricsRecalculatesRatesAndNormalizes(t *testing.T) {
	actual := EnrichMetrics(model.Metrics{TotalViews: 10, RepeatedViews: 2, ChatsStarted: 4, ChatsWithPurchase: 1, RepeatRate: .99, TopCategoryCode: " electronics ", TopCategory: " Электроника "})
	if math.Abs(actual.RepeatRate-.2) > 1e-9 || math.Abs(actual.PurchaseRate-.25) > 1e-9 {
		t.Fatalf("unexpected rates: %+v", actual)
	}
	if actual.TopCategoryCode != "electronics" || actual.TopCategory != "Электроника" {
		t.Fatalf("not normalized: %+v", actual)
	}
}

func TestCategoryCatalogueUsesUniqueCodes(t *testing.T) {
	seen := map[string]bool{}
	for _, item := range CategoryCatalogue() {
		if item.Code == "" || item.Title == "" || seen[item.Code] {
			t.Fatalf("invalid category: %+v", item)
		}
		seen[item.Code] = true
	}
}
