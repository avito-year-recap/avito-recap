package insight_test

import (
	"strings"
	"testing"
	"time"

	"github.com/year-recap/internal/recap/insight"
	"github.com/year-recap/internal/recap/model"
)

func TestBuildFactsMapsMetricsAndPeriod(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	metrics := model.Metrics{
		TotalEvents: 10, Searches: 2, TotalViews: 5, UniqueListings: 4, RepeatedViews: 1,
		FavoritesAdded: 1, ChatsStarted: 1, ListingsPublished: 0, PurchasesCompleted: 1,
		SalesCompleted: 0, ActiveDays: 3, CategoriesCount: 2, TopCategory: "Электроника",
		TopCategoryViews: 3, RepeatRate: 0.2, PurchaseRate: 1,
	}

	facts := insight.BuildFacts("active-buyer", start, end, metrics)

	if facts.ProfileCode != "active-buyer" {
		t.Fatalf("profile code = %q", facts.ProfileCode)
	}
	if !facts.StartAt.Equal(start) || !facts.EndAt.Equal(end) {
		t.Fatalf("period = %s..%s", facts.StartAt, facts.EndAt)
	}
	if facts.Metrics.TotalEvents != 10 || facts.Metrics.TopCategory != "Электроника" {
		t.Fatalf("unexpected metric facts: %+v", facts.Metrics)
	}
	if facts.Metrics.RepeatRatePercent != 20 || facts.Metrics.PurchaseRatePct != 100 {
		t.Fatalf("rates not converted to percent: %+v", facts.Metrics)
	}
}

func TestValidateCardRequiresTitleAndDescription(t *testing.T) {
	if err := insight.ValidateCard(insight.Card{Description: "text"}); err == nil {
		t.Fatal("expected error for empty title")
	}
	if err := insight.ValidateCard(insight.Card{Title: "title"}); err == nil {
		t.Fatal("expected error for empty description")
	}
}

func TestValidateCardEnforcesLengthLimits(t *testing.T) {
	longTitle := strings.Repeat("a", insight.MaxTitleRunes+1)
	if err := insight.ValidateCard(insight.Card{Title: longTitle, Description: "text"}); err == nil {
		t.Fatal("expected error for oversized title")
	}

	longDescription := strings.Repeat("a", insight.MaxDescriptionRunes+1)
	if err := insight.ValidateCard(insight.Card{Title: "title", Description: longDescription}); err == nil {
		t.Fatal("expected error for oversized description")
	}

	tooManyHighlights := make([]string, insight.MaxHighlights+1)
	for index := range tooManyHighlights {
		tooManyHighlights[index] = "highlight"
	}
	if err := insight.ValidateCard(insight.Card{Title: "title", Description: "text", Highlights: tooManyHighlights}); err == nil {
		t.Fatal("expected error for too many highlights")
	}

	longHighlight := strings.Repeat("a", insight.MaxHighlightRunes+1)
	if err := insight.ValidateCard(insight.Card{Title: "title", Description: "text", Highlights: []string{longHighlight}}); err == nil {
		t.Fatal("expected error for oversized highlight")
	}

	if err := insight.ValidateCard(insight.Card{Title: "title", Description: "text", Highlights: []string{"ok"}}); err != nil {
		t.Fatalf("unexpected error for valid card: %v", err)
	}
}
