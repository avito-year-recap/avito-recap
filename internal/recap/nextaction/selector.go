package nextaction

import (
	"sort"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

type recommendationContext struct {
	metrics  model.Metrics
	state    model.ActionableState
	behavior model.Behavior
}

type recommendationRule struct {
	name     string
	code     model.ActionCode
	priority int
	match    func(recommendationContext) bool
	build    func(recommendationContext) model.NextAction
}

// Build evaluates the executable CTA table against canonical metrics/state.
// Ruleset selection and input normalization belong to the engine.
func Build(r ruleset.Ruleset, metrics model.Metrics, state model.ActionableState, detected model.Behavior) model.NextAction {
	ctx := recommendationContext{
		metrics: metrics, state: state, behavior: detected,
	}
	rules := recommendationRules(r)
	matched := make([]recommendationRule, 0, len(rules))
	for _, rule := range rules {
		if rule.match(ctx) {
			matched = append(matched, rule)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].priority != matched[j].priority {
			return matched[i].priority > matched[j].priority
		}
		if matched[i].code != matched[j].code {
			return matched[i].code < matched[j].code
		}
		return matched[i].name < matched[j].name
	})
	return matched[0].build(ctx)
}
