package behavior

import (
	"sort"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

type behaviorCandidate struct {
	behavior model.Behavior
	eligible bool
	tieBreak uint32
}

type behaviorRule struct {
	code     model.BehaviorCode
	tieBreak uint32
	evaluate func(model.Metrics, ruleset.BehaviorThresholds) behaviorCandidate
}

// DetectBehavior is a compatibility wrapper around the explicit default ruleset.
func Detect(metrics model.Metrics) model.Behavior {
	return DetectWithRuleset(ruleset.DefaultRuleset(), metrics)
}

// DetectBehavior evaluates every behavior rule, then selects the highest score.
// Slice order is not business logic: ties are resolved by an explicit tie-break rank.
func DetectWithRuleset(r ruleset.Ruleset, metrics model.Metrics) model.Behavior {
	metrics = analytics.EnrichMetrics(metrics)
	candidates := make([]behaviorCandidate, 0, len(behaviorRules()))
	for _, rule := range behaviorRules() {
		candidate := rule.evaluate(metrics, r.Thresholds)
		candidate.tieBreak = rule.tieBreak
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
			Reason:      "Ни один специализированный сценарий не набрал минимальный набор доказательств.",
			Score:       0,
		}
	}

	ordered := append([]behaviorCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].behavior.Score != ordered[j].behavior.Score {
			return ordered[i].behavior.Score > ordered[j].behavior.Score
		}
		if ordered[i].tieBreak != ordered[j].tieBreak {
			return ordered[i].tieBreak > ordered[j].tieBreak
		}
		return ordered[i].behavior.Code < ordered[j].behavior.Code
	})
	return ordered[0].behavior
}

func behaviorRules() []behaviorRule {
	return []behaviorRule{
		{code: model.BehaviorResearcher, tieBreak: 10, evaluate: evaluateResearcher},
		{code: model.BehaviorFindHunter, tieBreak: 20, evaluate: evaluateFindHunter},
		{code: model.BehaviorStartingSeller, tieBreak: 30, evaluate: evaluateStartingSeller},
		{code: model.BehaviorDecisiveBuyer, tieBreak: 40, evaluate: evaluateDecisiveBuyer},
		{code: model.BehaviorActiveSeller, tieBreak: 50, evaluate: evaluateActiveSeller},
	}
}
