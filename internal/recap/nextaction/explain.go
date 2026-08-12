package nextaction

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

// Explain evaluates the full recommendation table and marks the rule that wins
// after deterministic priority sorting. Targets are intentionally omitted from
// the trace so the endpoint does not expose listing/dialog identifiers.
func Explain(r ruleset.Ruleset, metrics model.Metrics, state model.ActionableState, detected model.Behavior) []model.ActionRuleEvaluation {
	ctx := recommendationContext{metrics: metrics, state: state, behavior: detected}
	rules := recommendationRules(r)
	result := make([]model.ActionRuleEvaluation, 0, len(rules))
	for _, rule := range rules {
		result = append(result, model.ActionRuleEvaluation{
			Name: rule.name, Code: rule.code, Priority: rule.priority, Matched: rule.match(ctx),
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Priority != result[j].Priority {
			return result[i].Priority > result[j].Priority
		}
		if result[i].Code != result[j].Code {
			return result[i].Code < result[j].Code
		}
		return result[i].Name < result[j].Name
	})
	for index := range result {
		if result[index].Matched {
			result[index].Selected = true
			break
		}
	}
	return result
}
