package recap

import "github.com/google/uuid"

// BuildShareCard creates a strict public DTO. It intentionally exposes only the
// public share identifier; the internal recap identifier never crosses this boundary.
func BuildShareCard(value Recap) ShareCard {
	return buildShareCard(
		value.ShareID,
		value.Year,
		value.Metrics,
		value.Behavior,
		value.Achievements,
	)
}

// buildShareCard is the single source of truth for the final SHARE story card
// and the public GetShareCard response.
func buildShareCard(
	shareID uuid.UUID,
	year uint32,
	metrics Metrics,
	behavior Behavior,
	achievements []Achievement,
) ShareCard {
	result := ShareCard{
		ShareID:       shareID,
		Year:          year,
		BehaviorTitle: behavior.Title,
	}
	for _, achievement := range achievements {
		if achievement.Shareable {
			result.AchievementTitle = achievement.Title
			break
		}
	}
	if metrics.TopCategoryShareable {
		result.TopCategory = metrics.TopCategory
	}
	return result
}
