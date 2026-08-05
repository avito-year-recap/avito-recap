package recap

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

// BuildShareCard is a compatibility helper using the current default policy.
func BuildShareCard(value Recap) ShareCard {
	return BuildShareCardWithRuleset(DefaultRuleset(), value)
}

func BuildShareCardWithRuleset(ruleset Ruleset, value Recap) ShareCard {
	return buildShareCard(
		ruleset.SharePolicy,
		value.ShareID,
		value.Year,
		value.Metrics,
		value.Behavior,
		value.Achievements,
	)
}

// buildShareCard is the only public projection. It uses an explicit allow-list
// policy and rejects unsafe externally supplied text instead of trusting a flag.
func buildShareCard(
	policy SharePolicy,
	shareID uuid.UUID,
	year uint32,
	metrics Metrics,
	behavior Behavior,
	achievements []Achievement,
) ShareCard {
	result := ShareCard{
		ShareID:        shareID,
		Year:           year,
		PrivacyVersion: strings.TrimSpace(policy.Version),
	}
	if isSafePublicText(behavior.Title, policy.MaxPublicTextRunes) {
		result.BehaviorTitle = behavior.Title
	}
	for _, achievement := range achievements {
		if achievement.Shareable && policy.achievementAllowed(achievement.Code) &&
			isSafePublicText(achievement.Title, policy.MaxPublicTextRunes) {
			result.AchievementTitle = achievement.Title
			break
		}
	}
	categoryAllowed := policy.AllowTopCategory && isSafeCategoryCode(metrics.TopCategoryCode)
	if policy.RequireCategoryShareFlag {
		categoryAllowed = categoryAllowed && metrics.TopCategoryShareable
	}
	if categoryAllowed && isSafePublicText(metrics.TopCategory, policy.MaxPublicTextRunes) {
		result.TopCategory = metrics.TopCategory
	}
	return result
}

func isSafePublicText(value string, maxRunes int) bool {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxRunes {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.Is(unicode.Bidi_Control, r) {
			return false
		}
	}
	return true
}
