package behavior

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func checkGTE(metric string, actual, threshold float64, explanation string) model.RuleCheck {
	return model.RuleCheck{Metric: metric, Actual: actual, Threshold: threshold, Comparison: model.RuleComparisonGTE, Passed: actual >= threshold, Explanation: explanation}
}

func checkLTE(metric string, actual, threshold float64, explanation string) model.RuleCheck {
	return model.RuleCheck{Metric: metric, Actual: actual, Threshold: threshold, Comparison: model.RuleComparisonLTE, Passed: actual <= threshold, Explanation: explanation}
}

func checkGT(metric string, actual, threshold float64, explanation string) model.RuleCheck {
	return model.RuleCheck{Metric: metric, Actual: actual, Threshold: threshold, Comparison: model.RuleComparisonGT, Passed: actual > threshold, Explanation: explanation}
}

func checkEQ(metric string, actual, threshold float64, explanation string) model.RuleCheck {
	return model.RuleCheck{Metric: metric, Actual: actual, Threshold: threshold, Comparison: model.RuleComparisonEQ, Passed: actual == threshold, Explanation: explanation}
}

func allChecksPassed(checks []model.RuleCheck) bool {
	for _, check := range checks {
		if !check.Passed {
			return false
		}
	}
	return true
}

func evidenceFromChecks(checks []model.RuleCheck) []model.BehaviorEvidence {
	evidence := make([]model.BehaviorEvidence, 0, len(checks))
	for _, check := range checks {
		evidence = append(evidence, model.BehaviorEvidence{
			Metric: check.Metric, Actual: check.Actual, Threshold: check.Threshold, Detail: check.Explanation,
		})
	}
	return evidence
}

// Explain evaluates every behavior rule, including the rules that did not win.
// The returned order is deterministic: highest priority first.
func Explain(r ruleset.Ruleset, metrics model.Metrics) []model.BehaviorRuleEvaluation {
	rules := behaviorRules()
	selected := Detect(r, metrics).Code
	result := make([]model.BehaviorRuleEvaluation, 0, len(rules)+1)
	matchedAny := false

	for _, rule := range rules {
		candidate := rule.evaluate(metrics, r.Thresholds)
		if candidate.eligible {
			matchedAny = true
		}
		result = append(result, model.BehaviorRuleEvaluation{
			Code: rule.code, Priority: rule.priority, Matched: candidate.eligible,
			Selected: selected == rule.code, Checks: append([]model.RuleCheck(nil), candidate.checks...),
		})
	}

	if !matchedAny {
		result = append(result, model.BehaviorRuleEvaluation{
			Code: model.BehaviorUniversal, Priority: 0, Matched: true, Selected: true,
			Checks: []model.RuleCheck{},
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Priority != result[j].Priority {
			return result[i].Priority > result[j].Priority
		}
		return result[i].Code < result[j].Code
	})
	return result
}
