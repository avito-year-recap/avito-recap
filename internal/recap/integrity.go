package recap

import (
	"fmt"
	"reflect"
	"time"
)

type comparableAchievement struct {
	Code        AchievementCode
	Category    AchievementCategory
	Title       string
	Description string
	Reason      string
	Shareable   bool
}

func validateRecapAt(value Recap, now time.Time) error {
	if err := validateRecap(value); err != nil {
		return err
	}
	now = now.UTC()
	if value.GeneratedAt.After(now) {
		return fmt.Errorf("%w: generated time is in the future", ErrInvalidRecap)
	}
	if value.Period.EndAt.After(now) {
		return fmt.Errorf("%w: recap period is not complete at read time", ErrInvalidRecap)
	}
	return nil
}

// validateRecapAgainstRuleset treats storage as untrusted: every derived field
// is recomputed from the immutable inputs and compared with the stored value.
func validateRecapAgainstRuleset(value Recap, ruleset Ruleset, now time.Time) error {
	if err := ruleset.Validate(); err != nil {
		return fmt.Errorf("%w: stored ruleset is invalid: %v", ErrInvalidRecap, err)
	}
	if value.RulesVersion != ruleset.Version || value.RulesDigest != ruleset.Digest() {
		return fmt.Errorf("%w: rules identity mismatch", ErrInvalidRecap)
	}
	if err := validateRecapAt(value, now); err != nil {
		return err
	}

	expectedBehavior := ruleset.DetectBehavior(value.Metrics)
	if !reflect.DeepEqual(value.Behavior, expectedBehavior) {
		return fmt.Errorf("%w: stored behavior differs from ruleset result", ErrInvalidRecap)
	}
	if err := validateAchievementSelection(value.Achievements, ruleset.AchievementPolicy); err != nil {
		return fmt.Errorf("%w: stored achievement selection is invalid: %v", ErrInvalidRecap, err)
	}
	expectedAchievements := ruleset.BuildAchievements(value.Metrics)
	if !equalAchievements(value.Achievements, expectedAchievements) {
		return fmt.Errorf("%w: stored achievements differ from ruleset result", ErrInvalidRecap)
	}
	expectedAction := ruleset.BuildNextAction(value.Metrics, value.ActionableState, expectedBehavior)
	if !reflect.DeepEqual(value.NextAction, expectedAction) {
		return fmt.Errorf("%w: stored next action differs from ruleset result", ErrInvalidRecap)
	}
	expectedCards := BuildCardsWithRuleset(
		ruleset, value.Profile, value.Year, value.ShareID, value.Metrics,
		expectedBehavior, expectedAchievements, expectedAction,
	)
	if !reflect.DeepEqual(value.Cards, expectedCards) {
		return fmt.Errorf("%w: stored cards differ from deterministic card projection", ErrInvalidRecap)
	}
	expectedShare := BuildShareCardWithRuleset(ruleset, value)
	if expectedShare != value.Cards[len(value.Cards)-1].Payload.(ShareCard) {
		return fmt.Errorf("%w: public projection differs from final card", ErrInvalidRecap)
	}
	return nil
}

func equalAchievements(left, right []Achievement) bool {
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

// validateAchievementSelection checks the stored award list against the
// digest-bound policy without trusting transient Priority fields.
func validateAchievementSelection(values []Achievement, policy AchievementPolicy) error {
	if len(values) > policy.MaxAwarded || len(values) > maxAchievements {
		return fmt.Errorf("award count %d exceeds policy limit %d", len(values), policy.MaxAwarded)
	}
	configured := make(map[AchievementCode]AchievementRuleConfig, len(policy.Rules))
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
