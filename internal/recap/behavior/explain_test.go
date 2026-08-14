package behavior

import (
	"testing"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestExplainShowsFailedAndSelectedBehaviorRules(t *testing.T) {
	configured := ruleset.DefaultRuleset()
	metrics := analytics.EnrichMetrics(model.Metrics{
		TotalViews:      configured.Thresholds.ResearcherMinViews,
		CategoriesCount: configured.Thresholds.ResearcherMinCategories,
		ChatsStarted:    configured.Thresholds.ResearcherMaxChats,
	})

	explanation := Explain(configured, metrics)
	if len(explanation) != 5 {
		t.Fatalf("candidate count = %d, want 5", len(explanation))
	}

	var selected *model.BehaviorRuleEvaluation
	var activeSeller *model.BehaviorRuleEvaluation
	for index := range explanation {
		item := &explanation[index]
		if item.Selected {
			selected = item
		}
		if item.Code == model.BehaviorActiveSeller {
			activeSeller = item
		}
	}
	if selected == nil || selected.Code != model.BehaviorResearcher || !selected.Matched {
		t.Fatalf("unexpected selected behavior: %+v", selected)
	}
	if activeSeller == nil || activeSeller.Matched || len(activeSeller.Checks) != 2 {
		t.Fatalf("unexpected active-seller trace: %+v", activeSeller)
	}
	if activeSeller.Checks[0].Passed {
		t.Fatalf("expected failed listing threshold: %+v", activeSeller.Checks[0])
	}
}
