package recap

import "strings"

func normalizeProfile(profile Profile) Profile {
	profile.Code = strings.TrimSpace(profile.Code)
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.Description = strings.TrimSpace(profile.Description)
	profile.AvatarURL = strings.TrimSpace(profile.AvatarURL)
	return profile
}

func normalizeMetrics(metrics Metrics) Metrics {
	metrics.TopCategoryCode = strings.TrimSpace(metrics.TopCategoryCode)
	metrics.TopCategory = strings.TrimSpace(metrics.TopCategory)
	return metrics
}

func normalizeRecap(value Recap) Recap {
	value.Profile = normalizeProfile(value.Profile)
	value.RulesVersion = strings.TrimSpace(value.RulesVersion)
	value.Metrics = normalizeMetrics(value.Metrics)
	value.Behavior.Title = strings.TrimSpace(value.Behavior.Title)
	value.Behavior.Description = strings.TrimSpace(value.Behavior.Description)
	value.Behavior.Reason = strings.TrimSpace(value.Behavior.Reason)
	value.NextAction.Title = strings.TrimSpace(value.NextAction.Title)
	value.NextAction.Description = strings.TrimSpace(value.NextAction.Description)
	value.NextAction.ButtonText = strings.TrimSpace(value.NextAction.ButtonText)
	value.NextAction.Reason = strings.TrimSpace(value.NextAction.Reason)

	for index := range value.Achievements {
		achievement := &value.Achievements[index]
		achievement.Title = strings.TrimSpace(achievement.Title)
		achievement.Description = strings.TrimSpace(achievement.Description)
		achievement.Reason = strings.TrimSpace(achievement.Reason)
	}

	for index := range value.Cards {
		card := &value.Cards[index]
		card.ID = strings.TrimSpace(card.ID)
		card.Title = strings.TrimSpace(card.Title)
		card.Description = strings.TrimSpace(card.Description)
		card.Explanation = strings.TrimSpace(card.Explanation)
		card.Payload.CategoryCode = strings.TrimSpace(card.Payload.CategoryCode)
		card.Payload.Category = strings.TrimSpace(card.Payload.Category)
	}

	return value
}
