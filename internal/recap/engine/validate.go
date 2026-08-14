package engine

import (
	"fmt"
	"reflect"
	"time"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/validation/structural"
)

type comparableAchievement struct {
	Code        model.AchievementCode
	Category    model.AchievementCategory
	Title       string
	Description string
	Reason      string
	Shareable   bool
}

// ValidateStored treats persistence as untrusted. It validates immutable inputs,
// reruns the same deterministic derive pipeline used by Build, and compares all
// derived fields with the stored snapshot.
func (e *Engine) ValidateStored(value model.Recap, now time.Time) (model.Recap, error) {
	value = model.NormalizeRecap(value)
	now = now.UTC()

	if value.RulesVersion != e.rules.Version || value.RulesDigest != e.digest {
		return model.Recap{}, fmt.Errorf("%w: rules identity mismatch", structural.ErrInvalidRecap)
	}
	if err := structural.ValidateRecap(value); err != nil {
		return model.Recap{}, err
	}
	if value.GeneratedAt.After(now) {
		return model.Recap{}, fmt.Errorf("%w: generated time is in the future", structural.ErrInvalidRecap)
	}
	if value.Period.EndAt.After(now) {
		return model.Recap{}, fmt.Errorf("%w: recap period is not complete at read time", structural.ErrInvalidRecap)
	}
	if err := e.ensureEligible(value.Metrics); err != nil {
		return model.Recap{}, fmt.Errorf("%w: %w", structural.ErrInvalidRecap, err)
	}
	if err := e.validateAchievementSelection(value.Achievements); err != nil {
		return model.Recap{}, fmt.Errorf("%w: stored achievement selection is invalid: %w", structural.ErrInvalidRecap, err)
	}

	expected := e.derive(value.Profile, value.Year, value.ShareID, value.Metrics, value.ActionableState)
	if !reflect.DeepEqual(value.Behavior, expected.Behavior) {
		return model.Recap{}, fmt.Errorf("%w: stored behavior differs from engine result", structural.ErrInvalidRecap)
	}
	if !equalAchievements(value.Achievements, expected.Achievements) {
		return model.Recap{}, fmt.Errorf("%w: stored achievements differ from engine result", structural.ErrInvalidRecap)
	}
	if !reflect.DeepEqual(value.NextAction, expected.NextAction) {
		return model.Recap{}, fmt.Errorf("%w: stored next action differs from engine result", structural.ErrInvalidRecap)
	}
	if err := structural.ValidateCardsAgainstProjection(value.Cards, expected.Cards); err != nil {
		return model.Recap{}, fmt.Errorf("%w: stored cards differ from deterministic projection: %w", structural.ErrInvalidRecap, err)
	}
	return value, nil
}

func equalAchievements(left, right []model.Achievement) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		l := comparableAchievement{left[index].Code, left[index].Category, left[index].Title, left[index].Description, left[index].Reason, left[index].Shareable}
		r := comparableAchievement{right[index].Code, right[index].Category, right[index].Title, right[index].Description, right[index].Reason, right[index].Shareable}
		if l != r {
			return false
		}
	}
	return true
}

func (e *Engine) validateAchievementSelection(values []model.Achievement) error {
	policy := e.rules.AchievementPolicy
	if len(values) > policy.MaxAwarded || len(values) > ruleset.MaxAchievements {
		return fmt.Errorf("award count %d exceeds policy limit %d", len(values), policy.MaxAwarded)
	}
	configured := make(map[model.AchievementCode]ruleset.AchievementRuleConfig, len(policy.Rules))
	for _, rule := range policy.Rules {
		configured[rule.Code] = rule
	}
	for index, value := range values {
		rule, ok := configured[value.Code]
		if !ok {
			return fmt.Errorf("achievement %q is absent from policy", value.Code)
		}
		if value.Category != rule.Category {
			return fmt.Errorf("achievement %q category %q differs from policy category %q", value.Code, value.Category, rule.Category)
		}
		if index > 0 {
			previous := configured[values[index-1].Code]
			if previous.Priority < rule.Priority {
				return fmt.Errorf("achievements are not in deterministic priority order")
			}
		}
	}
	return nil
}
