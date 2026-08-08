package share

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

// buildShareCard is the only public projection. It uses an explicit allow-list
// policy and rejects unsafe externally supplied text instead of trusting a flag.
func Build(
	policy ruleset.SharePolicy,
	shareID uuid.UUID,
	year uint32,
	metrics model.Metrics,
	behavior model.Behavior,
	achievements []model.Achievement,
) model.ShareCard {
	result := model.ShareCard{
		ShareID:        shareID,
		Year:           year,
		PrivacyVersion: strings.TrimSpace(policy.Version),
	}
	if isSafePublicText(behavior.Title, policy.MaxPublicTextRunes) {
		result.BehaviorTitle = behavior.Title
	}
	for _, achievement := range achievements {
		if achievement.Shareable && policy.AchievementAllowed(achievement.Code) &&
			isSafePublicText(achievement.Title, policy.MaxPublicTextRunes) {
			result.AchievementTitle = achievement.Title
			break
		}
	}
	categoryAllowed := policy.AllowTopCategory &&
		isSafeCategoryCode(metrics.TopCategoryCode) &&
		analytics.IsShareableCategory(metrics.TopCategoryCode)
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

var categoryCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

func isSafeCategoryCode(code string) bool {
	return categoryCodePattern.MatchString(strings.TrimSpace(code))
}
