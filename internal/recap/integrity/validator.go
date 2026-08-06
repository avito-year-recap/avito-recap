package integrity

import (
	"fmt"
	"reflect"
	"time"

	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
	"github.com/year-recap/internal/recap/presentation/cards"
	"github.com/year-recap/internal/recap/presentation/share"
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

func ValidateRecapAt(value model.Recap, now time.Time) error {
	if err := structural.ValidateRecap(value); err != nil {
		return err
	}
	now = now.UTC()
	if value.GeneratedAt.After(now) {
		return fmt.Errorf("%w: generated time is in the future", structural.ErrInvalidRecap)
	}
	if value.Period.EndAt.After(now) {
		return fmt.Errorf("%w: recap period is not complete at read time", structural.ErrInvalidRecap)
	}
	return nil
}

// ValidateRecapAgainstRuleset treats storage as untrusted: every derived field
// is recomputed from the immutable inputs and compared with the stored value.
func ValidateRecapAgainstRuleset(value model.Recap, configured ruleset.Ruleset, now time.Time) error {
	if err := configured.Validate(); err != nil {
		return fmt.Errorf("%w: stored ruleset is invalid: %v", structural.ErrInvalidRecap, err)
	}
	if value.RulesVersion != configured.Version || value.RulesDigest != configured.Digest() {
		return fmt.Errorf("%w: rules identity mismatch", structural.ErrInvalidRecap)
	}
	if err := ValidateRecapAt(value, now); err != nil {
		return err
	}

	expectedBehavior := behavior.DetectWithRuleset(configured, value.Metrics)
	if !reflect.DeepEqual(value.Behavior, expectedBehavior) {
		return fmt.Errorf("%w: stored behavior differs from ruleset result", structural.ErrInvalidRecap)
	}
	if err := ValidateAchievementSelection(value.Achievements, configured.AchievementPolicy); err != nil {
		return fmt.Errorf("%w: stored achievement selection is invalid: %v", structural.ErrInvalidRecap, err)
	}
	expectedAchievements := achievement.BuildWithRuleset(configured, value.Metrics)
	if !equalAchievements(value.Achievements, expectedAchievements) {
		return fmt.Errorf("%w: stored achievements differ from ruleset result", structural.ErrInvalidRecap)
	}
	expectedAction := nextaction.BuildWithRuleset(configured, value.Metrics, value.ActionableState, expectedBehavior)
	if !reflect.DeepEqual(value.NextAction, expectedAction) {
		return fmt.Errorf("%w: stored next action differs from ruleset result", structural.ErrInvalidRecap)
	}
	expectedCards := cards.BuildWithRuleset(
		configured, value.Profile, value.Year, value.ShareID, value.Metrics,
		expectedBehavior, expectedAchievements, expectedAction,
	)
	if !reflect.DeepEqual(value.Cards, expectedCards) {
		return fmt.Errorf("%w: stored cards differ from deterministic card projection", structural.ErrInvalidRecap)
	}
	expectedShare := share.BuildWithRuleset(configured, value)
	if expectedShare != value.Cards[len(value.Cards)-1].Payload.(model.ShareCard) {
		return fmt.Errorf("%w: public projection differs from final card", structural.ErrInvalidRecap)
	}
	return nil
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

// ValidateAchievementSelection checks the stored award list against the
// digest-bound policy without trusting transient Priority fields.
func ValidateAchievementSelection(values []model.Achievement, policy ruleset.AchievementPolicy) error {
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
