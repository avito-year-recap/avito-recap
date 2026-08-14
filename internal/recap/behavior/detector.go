package behavior

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

type behaviorCandidate struct {
	behavior model.Behavior
	eligible bool
	priority uint32
	checks   []model.RuleCheck
}

type behaviorRule struct {
	code     model.BehaviorCode
	priority uint32
	evaluate func(model.Metrics, ruleset.BehaviorThresholds) behaviorCandidate
}

// Detect evaluates transparent eligibility rules against canonical, enriched
// metrics. Ruleset selection belongs to the engine/composition root.
func Detect(r ruleset.Ruleset, metrics model.Metrics) model.Behavior {
	candidates := make([]behaviorCandidate, 0, len(behaviorRules()))
	for _, rule := range behaviorRules() {
		candidate := rule.evaluate(metrics, r.Thresholds)
		candidate.priority = rule.priority
		if candidate.eligible {
			candidates = append(candidates, candidate)
		}
	}

	return selectBehaviorCandidate(candidates)
}

func selectBehaviorCandidate(candidates []behaviorCandidate) model.Behavior {
	if len(candidates) == 0 {
		return model.Behavior{
			Code:        model.BehaviorUniversal,
			Title:       "Разные сценарии",
			Description: "В течение года использовались разные возможности площадки без одного доминирующего сценария.",
			Reason:      "Ни один специализированный сценарий не выполнил полный набор пороговых условий.",
		}
	}

	ordered := append([]behaviorCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].priority != ordered[j].priority {
			return ordered[i].priority > ordered[j].priority
		}
		return ordered[i].behavior.Code < ordered[j].behavior.Code
	})
	return ordered[0].behavior
}

func behaviorRules() []behaviorRule {
	return []behaviorRule{
		{code: model.BehaviorResearcher, priority: 10, evaluate: evaluateResearcher},
		{code: model.BehaviorFindHunter, priority: 20, evaluate: evaluateFindHunter},
		{code: model.BehaviorStartingSeller, priority: 30, evaluate: evaluateStartingSeller},
		{code: model.BehaviorDecisiveBuyer, priority: 40, evaluate: evaluateDecisiveBuyer},
		{code: model.BehaviorActiveSeller, priority: 50, evaluate: evaluateActiveSeller},
	}
}
