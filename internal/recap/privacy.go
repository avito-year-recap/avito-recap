package recap

// BuildShareCard creates a strict public DTO instead of exposing the full recap.
// Top category is included only when the analytics layer explicitly marks it safe.
func BuildShareCard(value Recap) ShareCard {
	result := ShareCard{
		RecapID:       value.ID,
		Year:          value.Year,
		BehaviorTitle: value.Behavior.Title,
	}

	if value.Metrics.TopCategoryShareable {
		result.TopCategory = value.Metrics.TopCategory
	}

	for _, achievement := range value.Achievements {
		if achievement.Shareable {
			result.AchievementTitle = achievement.Title
			break
		}
	}

	return result
}
