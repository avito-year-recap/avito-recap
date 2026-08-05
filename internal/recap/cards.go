package recap

import (
	"fmt"
	"strings"
)

func BuildCards(
	profile Profile,
	year uint32,
	metrics Metrics,
	behavior Behavior,
	achievements []Achievement,
	nextAction NextAction,
) []Card {
	cards := make([]Card, 0, 8)

	appendCard := func(card Card) {
		card.Position = uint32(len(cards) + 1)
		cards = append(cards, card)
	}

	appendCard(Card{
		ID:          "intro",
		Type:        CardIntro,
		Title:       fmt.Sprintf("%s, вот твои итоги за %d год", profile.DisplayName, year),
		Description: "Мы связали главные действия в короткую историю — от интересов до следующего шага.",
	})

	appendCard(Card{
		ID:          "year-activity",
		Type:        CardYearActivity,
		Title:       "Год в цифрах",
		Description: fmt.Sprintf("За год система учла %d действий.", metrics.TotalEvents),
		Explanation: "Учитывались поиски, просмотры, избранное, диалоги, создание и публикация объявлений, покупки и продажи.",
		Payload: CardPayload{
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
		appendCard(Card{
			ID:          "top-category",
			Type:        CardTopCategory,
			Title:       "Главный интерес года",
			Description: fmt.Sprintf("Чаще всего внимание привлекала категория «%s».", metrics.TopCategory),
			Explanation: fmt.Sprintf("Просмотров в этой категории: %d.", metrics.TopCategoryViews),
			Shareable:   metrics.TopCategoryShareable,
			Payload: CardPayload{
				CategoryCode:  metrics.TopCategoryCode,
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

	if missed := buildMissedOpportunityCard(metrics, nextAction); missed != nil {
		appendCard(*missed)
	} else if achievementCard := buildAchievementCard(achievements); achievementCard != nil {
		appendCard(*achievementCard)
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
		Title:       "Итоги готовы",
		Description: "Главное за год собрано — следующий шаг уже связан с тем, что было важно именно в этом сценарии.",
	})

	return cards
}

func buildAchievementCard(achievements []Achievement) *Card {
	if len(achievements) == 0 {
		return nil
	}

	titles := make([]string, 0, len(achievements))
	reasons := make([]string, 0, len(achievements))
	codes := make([]AchievementCode, 0, len(achievements))
	shareable := true

	for _, achievement := range achievements {
		titles = append(titles, achievement.Title)
		reasons = append(reasons, achievement.Reason)
		codes = append(codes, achievement.Code)
		shareable = shareable && achievement.Shareable
	}

	return &Card{
		ID:          "achievements",
		Type:        CardAchievement,
		Title:       "Ачивки года",
		Description: strings.Join(titles, " • "),
		Explanation: strings.Join(reasons, " "),
		Shareable:   shareable,
		Payload: CardPayload{
			AchievementCode:  codes[0],
			AchievementCodes: codes,
		},
	}
}

func buildMissedOpportunityCard(metrics Metrics, nextAction NextAction) *Card {
	switch nextAction.Code {
	case ActionSaveSearch:
		return &Card{
			ID:          "missed-opportunity",
			Type:        CardMissedOpportunity,
			Title:       "Возможность, которая сэкономит время",
			Description: "Сохранённый поиск может сам показывать новые объявления по интересующим параметрам.",
			Explanation: fmt.Sprintf(
				"Просмотров было %d, а диалогов — %d: автоматическое обновление поиска поможет не повторять сравнение вручную.",
				metrics.TotalViews,
				metrics.ChatsStarted,
			),
			Payload: CardPayload{ActionCode: nextAction.Code},
		}

	case ActionFinishDraft:
		drafts := metrics.ListingsCreated - metrics.ListingsPublished
		return &Card{
			ID:          "missed-opportunity",
			Type:        CardMissedOpportunity,
			Title:       "Шанс довести начатое до публикации",
			Description: "Незавершённые объявления ещё могут привести к просмотрам и сообщениям.",
			Explanation: fmt.Sprintf("Черновиков осталось: %d.", drafts),
			Payload:     CardPayload{ActionCode: nextAction.Code},
		}
	}

	return nil
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
