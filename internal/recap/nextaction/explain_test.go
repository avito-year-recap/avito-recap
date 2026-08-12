package nextaction

import (
	"testing"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestExplainMarksHighestPriorityMatchedAction(t *testing.T) {
	configured := ruleset.DefaultRuleset()
	metrics := analytics.EnrichMetrics(model.Metrics{TopCategoryCode: "auto", TopCategory: "Авто", TopCategoryViews: 1})
	state := model.ActionableState{FavoritesCount: 4}
	detected := behavior.Detect(configured, metrics)

	explanation := Explain(configured, metrics, state, detected)
	if len(explanation) == 0 {
		t.Fatal("expected recommendation candidates")
	}
	selected := 0
	for _, item := range explanation {
		if !item.Selected {
			continue
		}
		selected++
		if item.Code != model.ActionOpenFavorites {
			t.Fatalf("selected = %s, want %s", item.Code, model.ActionOpenFavorites)
		}
		if !item.Matched {
			t.Fatalf("selected action must match: %+v", item)
		}
	}
	if selected != 1 {
		t.Fatalf("selected count = %d, want 1", selected)
	}
}
