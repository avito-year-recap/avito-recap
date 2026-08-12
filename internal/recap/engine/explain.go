package engine

import (
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
)

// Explain produces a deterministic, read-only trace for an already generated
// recap. The trace is derived from the same ruleset and canonical metrics used
// by Build, so it can be used in demos, support tooling and rule debugging.
func (e *Engine) Explain(value model.Recap) model.RecapExplanation {
	value = model.NormalizeRecap(value)
	achievements := make([]model.AchievementExplanation, 0, len(value.Achievements))
	for _, achievement := range value.Achievements {
		achievements = append(achievements, model.AchievementExplanation{
			Code: achievement.Code, Title: achievement.Title, Reason: achievement.Reason,
			Shareable: achievement.Shareable,
		})
	}
	return model.RecapExplanation{
		ProfileCode:        value.Profile.Code,
		Year:               value.Year,
		RulesVersion:       value.RulesVersion,
		RulesDigest:        value.RulesDigest,
		Behavior:           value.Behavior,
		BehaviorCandidates: behavior.Explain(e.rules, value.Metrics),
		Achievements:       achievements,
		NextAction: model.NextActionExplanation{
			Code: value.NextAction.Code, Title: value.NextAction.Title, Description: value.NextAction.Description,
			ButtonText: value.NextAction.ButtonText, Reason: value.NextAction.Reason,
		},
		ActionCandidates: nextaction.Explain(e.rules, value.Metrics, value.ActionableState, value.Behavior),
	}
}
