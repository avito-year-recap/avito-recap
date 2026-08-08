package cards

import (
	"fmt"
	"strings"

	"github.com/year-recap/internal/recap/model"
)

func MonthName(month uint32) string {
	names := [...]string{"", "январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"}
	if month >= uint32(len(names)) {
		return ""
	}
	return names[month]
}

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

func buildMissedOpportunityCard(metrics model.Metrics, nextAction model.NextAction) *model.Card {
	switch nextAction.Code {
	case model.ActionSaveSearch:
		return &model.Card{
			ID:          "missed-opportunity",
			Type:        model.CardMissedOpportunity,
			Title:       "Возможность, которая сэкономит время",
			Description: "Сохранённый поиск может сам показывать новые объявления по главному интересу.",
			Explanation: fmt.Sprintf("Просмотров было %d, а сохранённого поиска по главной категории сейчас нет.", metrics.TotalViews),
			Payload:     model.ActionPayload{Code: nextAction.Code, Target: nextAction.Target},
		}
	case model.ActionFinishDraft:
		return &model.Card{
			ID:          "missed-opportunity",
			Type:        model.CardMissedOpportunity,
			Title:       "Шанс довести актуальный черновик до публикации",
			Description: "Текущий черновик можно открыть напрямую.",
			Explanation: "Карточка опирается на актуальное состояние объявлений, а не на разницу годовых счётчиков.",
			Payload:     model.ActionPayload{Code: nextAction.Code, Target: nextAction.Target},
		}
	case model.ActionOpenFavorites,
		model.ActionImproveListings,
		model.ActionContinueDialogs,
		model.ActionOpenTopCategory,
		model.ActionCreateFirstListing,
		model.ActionCreateListing,
		model.ActionViewSimilarListings,
		model.ActionExploreRecommendations:
		return nil
	}
	return nil
}
