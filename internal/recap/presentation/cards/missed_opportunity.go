package cards

import (
	"fmt"

	"github.com/year-recap/internal/recap/model"
)

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
