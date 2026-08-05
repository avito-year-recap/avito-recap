package recap

import "strings"

func normalizeString(value string) string { return strings.TrimSpace(value) }

func normalizeProfile(profile Profile) Profile {
	profile.Code = normalizeString(profile.Code)
	profile.DisplayName = normalizeString(profile.DisplayName)
	profile.Description = normalizeString(profile.Description)
	profile.AvatarURL = normalizeString(profile.AvatarURL)
	return profile
}

func normalizeMetrics(metrics Metrics) Metrics {
	metrics.TopCategoryCode = normalizeString(metrics.TopCategoryCode)
	metrics.TopCategory = normalizeString(metrics.TopCategory)
	return metrics
}

func normalizeActionableState(state ActionableState) ActionableState { return state }

func normalizeActionTarget(target ActionTarget) ActionTarget {
	if target.Route != nil {
		value := *target.Route
		value.Route = normalizeString(value.Route)
		target.Route = &value
	}
	if target.Category != nil {
		value := *target.Category
		value.CategoryCode = normalizeString(value.CategoryCode)
		target.Category = &value
	}
	if target.Listing != nil {
		value := *target.Listing
		target.Listing = &value
	}
	if target.Dialog != nil {
		value := *target.Dialog
		target.Dialog = &value
	}
	if target.Search != nil {
		value := *target.Search
		value.CategoryCode = normalizeString(value.CategoryCode)
		target.Search = &value
	}
	return target
}

func normalizeRecap(value Recap) Recap {
	value.Behavior.Evidence = append([]BehaviorEvidence(nil), value.Behavior.Evidence...)
	value.Achievements = append([]Achievement(nil), value.Achievements...)
	value.Cards = append([]Card(nil), value.Cards...)
	value.Profile = normalizeProfile(value.Profile)
	value.RulesVersion = normalizeString(value.RulesVersion)
	value.RulesDigest = normalizeString(value.RulesDigest)
	value.Metrics = normalizeMetrics(value.Metrics)
	value.ActionableState = normalizeActionableState(value.ActionableState)
	value.Behavior.Title = normalizeString(value.Behavior.Title)
	value.Behavior.Description = normalizeString(value.Behavior.Description)
	value.Behavior.Reason = normalizeString(value.Behavior.Reason)
	for index := range value.Behavior.Evidence {
		value.Behavior.Evidence[index].Metric = normalizeString(value.Behavior.Evidence[index].Metric)
		value.Behavior.Evidence[index].Detail = normalizeString(value.Behavior.Evidence[index].Detail)
	}
	value.NextAction.Title = normalizeString(value.NextAction.Title)
	value.NextAction.Description = normalizeString(value.NextAction.Description)
	value.NextAction.ButtonText = normalizeString(value.NextAction.ButtonText)
	value.NextAction.Reason = normalizeString(value.NextAction.Reason)
	value.NextAction.Target = normalizeActionTarget(value.NextAction.Target)

	for index := range value.Achievements {
		achievement := &value.Achievements[index]
		achievement.Title = normalizeString(achievement.Title)
		achievement.Description = normalizeString(achievement.Description)
		achievement.Reason = normalizeString(achievement.Reason)
	}

	for index := range value.Cards {
		card := &value.Cards[index]
		card.ID = normalizeString(card.ID)
		card.Title = normalizeString(card.Title)
		card.Description = normalizeString(card.Description)
		card.Explanation = normalizeString(card.Explanation)
		card.Payload = normalizeCardPayload(card.Payload)
	}
	return value
}

func normalizeCardPayload(payload CardPayload) CardPayload {
	switch value := payload.(type) {
	case TopCategoryPayload:
		value.CategoryCode = normalizeString(value.CategoryCode)
		value.Category = normalizeString(value.Category)
		return value
	case BehaviorPayload:
		value.Evidence = append([]BehaviorEvidence(nil), value.Evidence...)
		for index := range value.Evidence {
			value.Evidence[index].Metric = normalizeString(value.Evidence[index].Metric)
			value.Evidence[index].Detail = normalizeString(value.Evidence[index].Detail)
		}
		return value
	case AchievementPayload:
		value.Codes = append([]AchievementCode(nil), value.Codes...)
		return value
	case ActionPayload:
		value.Target = normalizeActionTarget(value.Target)
		return value
	case ShareCard:
		value.BehaviorTitle = normalizeString(value.BehaviorTitle)
		value.AchievementTitle = normalizeString(value.AchievementTitle)
		value.TopCategory = normalizeString(value.TopCategory)
		return value
	default:
		return payload
	}
}
