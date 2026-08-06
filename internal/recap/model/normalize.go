package model

import (
	"sort"
	"strings"
)

func NormalizeString(value string) string { return strings.TrimSpace(value) }

func NormalizeProfile(profile Profile) Profile {
	profile.Code = NormalizeString(profile.Code)
	profile.DisplayName = NormalizeString(profile.DisplayName)
	profile.Description = NormalizeString(profile.Description)
	profile.AvatarURL = NormalizeString(profile.AvatarURL)
	return profile
}

func NormalizeMetrics(metrics Metrics) Metrics {
	metrics.TopCategoryCode = NormalizeString(metrics.TopCategoryCode)
	metrics.TopCategory = NormalizeString(metrics.TopCategory)
	metrics.CategoryActivities = append([]CategoryActivity(nil), metrics.CategoryActivities...)
	for index := range metrics.CategoryActivities {
		activity := &metrics.CategoryActivities[index]
		activity.CategoryCode = NormalizeString(activity.CategoryCode)
		activity.Category = NormalizeString(activity.Category)
	}
	sort.Slice(metrics.CategoryActivities, func(i, j int) bool {
		return metrics.CategoryActivities[i].CategoryCode < metrics.CategoryActivities[j].CategoryCode
	})
	return metrics
}

func NormalizeActionableState(state ActionableState) ActionableState { return state }

func NormalizeActionTarget(target ActionTarget) ActionTarget {
	if target.Route != nil {
		value := *target.Route
		value.Route = NormalizeString(value.Route)
		target.Route = &value
	}
	if target.Category != nil {
		value := *target.Category
		value.CategoryCode = NormalizeString(value.CategoryCode)
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
		value.CategoryCode = NormalizeString(value.CategoryCode)
		target.Search = &value
	}
	return target
}

func NormalizeRecap(value Recap) Recap {
	value.Behavior.Evidence = append([]BehaviorEvidence(nil), value.Behavior.Evidence...)
	value.Achievements = append([]Achievement(nil), value.Achievements...)
	value.Cards = append([]Card(nil), value.Cards...)
	value.Profile = NormalizeProfile(value.Profile)
	value.RulesVersion = NormalizeString(value.RulesVersion)
	value.RulesDigest = NormalizeString(value.RulesDigest)
	value.Metrics = NormalizeMetrics(value.Metrics)
	value.ActionableState = NormalizeActionableState(value.ActionableState)
	value.Behavior.Title = NormalizeString(value.Behavior.Title)
	value.Behavior.Description = NormalizeString(value.Behavior.Description)
	value.Behavior.Reason = NormalizeString(value.Behavior.Reason)
	for index := range value.Behavior.Evidence {
		value.Behavior.Evidence[index].Metric = NormalizeString(value.Behavior.Evidence[index].Metric)
		value.Behavior.Evidence[index].Detail = NormalizeString(value.Behavior.Evidence[index].Detail)
	}
	value.NextAction.Title = NormalizeString(value.NextAction.Title)
	value.NextAction.Description = NormalizeString(value.NextAction.Description)
	value.NextAction.ButtonText = NormalizeString(value.NextAction.ButtonText)
	value.NextAction.Reason = NormalizeString(value.NextAction.Reason)
	value.NextAction.Target = NormalizeActionTarget(value.NextAction.Target)

	for index := range value.Achievements {
		achievement := &value.Achievements[index]
		achievement.Title = NormalizeString(achievement.Title)
		achievement.Description = NormalizeString(achievement.Description)
		achievement.Reason = NormalizeString(achievement.Reason)
	}

	for index := range value.Cards {
		card := &value.Cards[index]
		card.ID = NormalizeString(card.ID)
		card.Title = NormalizeString(card.Title)
		card.Description = NormalizeString(card.Description)
		card.Explanation = NormalizeString(card.Explanation)
		card.Payload = NormalizeCardPayload(card.Payload)
	}
	return value
}

func NormalizeCardPayload(payload CardPayload) CardPayload {
	switch value := payload.(type) {
	case TopCategoryPayload:
		value.CategoryCode = NormalizeString(value.CategoryCode)
		value.Category = NormalizeString(value.Category)
		return value
	case BehaviorPayload:
		value.Evidence = append([]BehaviorEvidence(nil), value.Evidence...)
		for index := range value.Evidence {
			value.Evidence[index].Metric = NormalizeString(value.Evidence[index].Metric)
			value.Evidence[index].Detail = NormalizeString(value.Evidence[index].Detail)
		}
		return value
	case AchievementPayload:
		value.Codes = append([]AchievementCode(nil), value.Codes...)
		return value
	case ActionPayload:
		value.Target = NormalizeActionTarget(value.Target)
		return value
	case ShareCard:
		value.BehaviorTitle = NormalizeString(value.BehaviorTitle)
		value.AchievementTitle = NormalizeString(value.AchievementTitle)
		value.TopCategory = NormalizeString(value.TopCategory)
		return value
	default:
		return payload
	}
}
