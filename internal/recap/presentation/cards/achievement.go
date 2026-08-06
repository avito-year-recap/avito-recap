package cards

import (
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func buildAchievementCard(achievements []model.Achievement) *model.Card {
	if len(achievements) == 0 {
		return nil
	}
	titles := make([]string, 0, len(achievements))
	reasons := make([]string, 0, len(achievements))
	codes := make([]model.AchievementCode, 0, len(achievements))
	for _, achievement := range achievements {
		titles = append(titles, achievement.Title)
		reasons = append(reasons, achievement.Reason)
		codes = append(codes, achievement.Code)
	}
	return &model.Card{
		ID: "achievements", Type: model.CardAchievement, Title: "Ачивки года",
		Description: strings.Join(titles, " • "), Explanation: strings.Join(reasons, " "),
		Payload: model.AchievementPayload{Codes: codes},
	}
}
