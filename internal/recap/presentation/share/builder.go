package share

import (
	"strings"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func Build(value model.Recap) model.ShareCard {
	return BuildWithRuleset(ruleset.DefaultRuleset(), value)
}

func BuildWithRuleset(configured ruleset.Ruleset, value model.Recap) model.ShareCard {
	return BuildFromParts(
		configured.SharePolicy,
		value.ShareID,
		value.Year,
		value.Metrics,
		value.Behavior,
		value.Achievements,
	)
}

// buildShareCard is the only public projection. It uses an explicit allow-list
// policy and rejects unsafe externally supplied text instead of trusting a flag.
func BuildFromParts(
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
