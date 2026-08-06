package cards

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/presentation/share"
	"github.com/year-recap/internal/recap/ruleset"
)

func Build(
	profile model.Profile,
	year uint32,
	shareID uuid.UUID,
	metrics model.Metrics,
	behavior model.Behavior,
	achievements []model.Achievement,
	nextAction model.NextAction,
) []model.Card {
	return BuildWithRuleset(ruleset.DefaultRuleset(), profile, year, shareID, metrics, behavior, achievements, nextAction)
}

func BuildWithRuleset(
	configured ruleset.Ruleset,
	profile model.Profile,
	year uint32,
	shareID uuid.UUID,
	metrics model.Metrics,
	behavior model.Behavior,
	achievements []model.Achievement,
	nextAction model.NextAction,
) []model.Card {
	profile = model.NormalizeProfile(profile)
	metrics = model.NormalizeMetrics(metrics)
	shareCard := share.BuildFromParts(configured.SharePolicy, shareID, year, metrics, behavior, achievements)
	cards := make([]model.Card, 0, 9)

	appendCard := func(card model.Card) {
		card.Position = uint32(len(cards) + 1)
		cards = append(cards, card)
	}

	appendCard(model.Card{
		ID:          "intro",
		Type:        model.CardIntro,
		Title:       fmt.Sprintf("%s, вот твои итоги за %d год", profile.DisplayName, year),
		Description: "Это финальный recap за завершённый календарный год.",
	})

	appendCard(model.Card{
		ID:          "year-activity",
		Type:        model.CardYearActivity,
		Title:       "Год в цифрах",
		Description: fmt.Sprintf("За завершённый год система учла %d действий.", metrics.TotalEvents),
		Explanation: "Учитывались поиски, просмотры, избранное, диалоги, создание и публикация объявлений, покупки и продажи.",
		Payload: model.YearActivityPayload{
			TotalEvents:        metrics.TotalEvents,
			Searches:           metrics.Searches,
			TotalViews:         metrics.TotalViews,
			FavoritesAdded:     metrics.FavoritesAdded,
			ChatsStarted:       metrics.ChatsStarted,
			ListingsPublished:  metrics.ListingsPublished,
			PurchasesCompleted: metrics.PurchasesCompleted,
			SalesCompleted:     metrics.SalesCompleted,
		},
	})

	if metrics.TopCategory != "" && metrics.TopCategoryViews > 0 {
		appendCard(model.Card{
			ID:          "top-category",
			Type:        model.CardTopCategory,
			Title:       "Главный интерес года",
			Description: fmt.Sprintf("Чаще всего внимание привлекала категория «%s».", metrics.TopCategory),
			Explanation: fmt.Sprintf("Просмотров в этой категории: %d.", metrics.TopCategoryViews),
			Payload: model.TopCategoryPayload{
				CategoryCode:  metrics.TopCategoryCode,
				Category:      metrics.TopCategory,
				CategoryViews: metrics.TopCategoryViews,
			},
		})
	}

	if metrics.MostActiveMonth >= 1 && metrics.MostActiveMonth <= 12 {
		appendCard(model.Card{
			ID:          "active-month",
			Type:        model.CardActiveMonth,
			Title:       "Самый активный месяц",
			Description: fmt.Sprintf("Больше всего действий пришлось на %s.", MonthName(metrics.MostActiveMonth)),
			Explanation: "Месяц выбран по максимальному числу событий профиля.",
			Payload:     model.ActiveMonthPayload{Month: metrics.MostActiveMonth},
		})
	}

	appendCard(model.Card{
		ID:          "behavior",
		Type:        model.CardBehavior,
		Title:       behavior.Title,
		Description: behavior.Description,
		Explanation: behavior.Reason,
		Payload: model.BehaviorPayload{
			Code: behavior.Code, Score: behavior.Score, Evidence: behavior.Evidence,
		},
	})

	if achievementCard := buildAchievementCard(achievements); achievementCard != nil {
		appendCard(*achievementCard)
	}

	if missed := buildMissedOpportunityCard(metrics, nextAction); missed != nil {
		appendCard(*missed)
	}

	appendCard(model.Card{
		ID:          "next-action",
		Type:        model.CardNextAction,
		Title:       nextAction.Title,
		Description: nextAction.Description,
		Explanation: nextAction.Reason,
		Payload:     model.ActionPayload{Code: nextAction.Code, Target: nextAction.Target},
	})

	appendCard(model.Card{
		ID:          "share",
		Type:        model.CardShare,
		Title:       "Итоги готовы — можно делиться",
		Description: fmt.Sprintf("В %d году ты — «%s». Карточка содержит только данные, разрешённые для публичного показа.", year, behavior.Title),
		Explanation: "Годовой recap сохранён как неизменяемый снимок состояния на момент генерации.",
		Shareable:   true,
		Payload:     shareCard,
	})

	return cards
}
