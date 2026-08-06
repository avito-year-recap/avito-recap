package nextaction

import (
	"sort"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
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

// BuildNextAction is a compatibility wrapper. New code should pass an explicit
// model.ActionableState to Ruleset.BuildNextAction.
func Build(metrics model.Metrics, states ...model.ActionableState) model.NextAction {
	state := model.ActionableState{}
	if len(states) > 0 {
		state = states[0]
	}
	configured := ruleset.DefaultRuleset()
	detected := behavior.DetectWithRuleset(configured, metrics)
	return BuildWithRuleset(configured, metrics, state, detected)
}

// BuildNextAction evaluates an explicit priority table. The user-facing output
// is deliberately restricted to three product-approved variants.
func BuildWithRuleset(r ruleset.Ruleset, metrics model.Metrics, state model.ActionableState, detected model.Behavior) model.NextAction {
	ctx := recommendationContext{
		metrics: analytics.EnrichMetrics(metrics), state: model.NormalizeActionableState(state), behavior: detected,
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
