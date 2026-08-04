package recap

import "fmt"

func BuildCards(
	profile Profile,
	year uint32,
	metrics Metrics,
	behavior Behavior,
	achievements []Achievement,
	nextAction NextAction,
) []Card {
	cards := make([]Card, 0, 7+len(achievements))

	appendCard := func(card Card) {
		card.Position = uint32(len(cards) + 1)
		cards = append(cards, card)
	}

	appendCard(Card{
		ID:          "intro",
		Type:        CardIntro,
		Title:       fmt.Sprintf("%s, твой %d год на площадке", profile.DisplayName, year),
		Description: "Мы собрали главные действия и превратили их в короткую историю.",
	})

	appendCard(Card{
		ID:          "year-activity",
		Type:        CardYearActivity,
		Title:       "Год в цифрах",
		Description: fmt.Sprintf("За год система учла %d действий.", metrics.TotalEvents),
		Explanation: "Учитывались просмотры, избранное, диалоги, создание и публикация объявлений, а также завершённые сделки.",
		Payload: CardPayload{
			TotalEvents:        metrics.TotalEvents,
			TotalViews:         metrics.TotalViews,
			FavoritesAdded:     metrics.FavoritesAdded,
			ChatsStarted:       metrics.ChatsStarted,
			ListingsPublished:  metrics.ListingsPublished,
			PurchasesCompleted: metrics.PurchasesCompleted,
			SalesCompleted:     metrics.SalesCompleted,
		},
	})

	if metrics.TopCategory != "" && metrics.TopCategoryViews > 0 {
		appendCard(Card{
			ID:          "top-category",
			Type:        CardTopCategory,
			Title:       "Главный интерес года",
			Description: fmt.Sprintf("Чаще всего тебя интересовала категория «%s».", metrics.TopCategory),
			Explanation: fmt.Sprintf("В этой категории зафиксировано %d просмотров.", metrics.TopCategoryViews),
			Shareable:   metrics.TopCategoryShareable,
			Payload: CardPayload{
				Category:      metrics.TopCategory,
				CategoryViews: metrics.TopCategoryViews,
			},
		})
	}

	if metrics.MostActiveMonth >= 1 && metrics.MostActiveMonth <= 12 {
		appendCard(Card{
			ID:          "active-month",
			Type:        CardActiveMonth,
			Title:       "Самый активный месяц",
			Description: fmt.Sprintf("Больше всего действий пришлось на %s.", monthName(metrics.MostActiveMonth)),
			Explanation: "Месяц выбран по максимальному числу событий профиля.",
			Payload: CardPayload{
				Month: metrics.MostActiveMonth,
			},
		})
	}

	appendCard(Card{
		ID:          "behavior",
		Type:        CardBehavior,
		Title:       behavior.Title,
		Description: behavior.Description,
		Explanation: behavior.Reason,
		Shareable:   true,
		Payload: CardPayload{
			BehaviorCode: behavior.Code,
		},
	})

	for _, achievement := range achievements {
		appendCard(Card{
			ID:          "achievement-" + string(achievement.Code),
			Type:        CardAchievement,
			Title:       achievement.Title,
			Description: achievement.Description,
			Explanation: achievement.Reason,
			Shareable:   achievement.Shareable,
			Payload: CardPayload{
				AchievementCode: achievement.Code,
			},
		})
	}

	appendCard(Card{
		ID:          "next-action",
		Type:        CardNextAction,
		Title:       nextAction.Title,
		Description: nextAction.Description,
		Explanation: nextAction.Reason,
		Payload: CardPayload{
			ActionCode: nextAction.Code,
		},
	})

	appendCard(Card{
		ID:          "summary",
		Type:        CardSummary,
		Title:       "Твои итоги готовы",
		Description: "Главное за год собрано — теперь можно перейти к следующему действию.",
	})

	return cards
}

func monthName(month uint32) string {
	names := [...]string{
		"",
		"январь",
		"февраль",
		"март",
		"апрель",
		"май",
		"июнь",
		"июль",
		"август",
		"сентябрь",
		"октябрь",
		"ноябрь",
		"декабрь",
	}

	if month >= uint32(len(names)) {
		return ""
	}

	return names[month]
}
