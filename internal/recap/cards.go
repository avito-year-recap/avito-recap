package recap

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func BuildCards(
	profile Profile,
	year uint32,
	shareID uuid.UUID,
	metrics Metrics,
	behavior Behavior,
	achievements []Achievement,
	nextAction NextAction,
) []Card {
	return BuildCardsWithRuleset(DefaultRuleset(), profile, year, shareID, metrics, behavior, achievements, nextAction)
}

func BuildCardsWithRuleset(
	ruleset Ruleset,
	profile Profile,
	year uint32,
	shareID uuid.UUID,
	metrics Metrics,
	behavior Behavior,
	achievements []Achievement,
	nextAction NextAction,
) []Card {
	profile = normalizeProfile(profile)
	metrics = normalizeMetrics(metrics)
	shareCard := buildShareCard(ruleset.SharePolicy, shareID, year, metrics, behavior, achievements)
	cards := make([]Card, 0, 9)

	appendCard := func(card Card) {
		card.Position = uint32(len(cards) + 1)
		cards = append(cards, card)
	}

	appendCard(Card{
		ID:          "intro",
		Type:        CardIntro,
		Title:       fmt.Sprintf("%s, вот твои итоги за %d год", profile.DisplayName, year),
		Description: "Это финальный recap за завершённый календарный год.",
	})

	appendCard(Card{
		ID:          "year-activity",
		Type:        CardYearActivity,
		Title:       "Год в цифрах",
		Description: fmt.Sprintf("За завершённый год система учла %d действий.", metrics.TotalEvents),
		Explanation: "Учитывались поиски, просмотры, избранное, диалоги, создание и публикация объявлений, покупки и продажи.",
		Payload: YearActivityPayload{
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
			Payload: TopCategoryPayload{
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
			Payload:     ActiveMonthPayload{Month: metrics.MostActiveMonth},
		})
	}

	appendCard(Card{
		ID:          "behavior",
		Type:        CardBehavior,
		Title:       behavior.Title,
		Description: behavior.Description,
		Explanation: behavior.Reason,
		Payload: BehaviorPayload{
			Code: behavior.Code, Score: behavior.Score, Evidence: behavior.Evidence,
		},
	})

	if achievementCard := buildAchievementCard(achievements); achievementCard != nil {
		appendCard(*achievementCard)
	}

	if missed := buildMissedOpportunityCard(metrics, nextAction); missed != nil {
		appendCard(*missed)
	}

	appendCard(Card{
		ID:          "next-action",
		Type:        CardNextAction,
		Title:       nextAction.Title,
		Description: nextAction.Description,
		Explanation: nextAction.Reason,
		Payload:     ActionPayload{Code: nextAction.Code, Target: nextAction.Target},
	})

	appendCard(Card{
		ID:          "share",
		Type:        CardShare,
		Title:       "Итоги готовы — можно делиться",
		Description: fmt.Sprintf("В %d году ты — «%s». Карточка содержит только данные, разрешённые для публичного показа.", year, behavior.Title),
		Explanation: "Годовой recap сохранён как неизменяемый снимок состояния на момент генерации.",
		Shareable:   true,
		Payload:     shareCard,
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
	for _, achievement := range achievements {
		titles = append(titles, achievement.Title)
		reasons = append(reasons, achievement.Reason)
		codes = append(codes, achievement.Code)
	}
	return &Card{
		ID: "achievements", Type: CardAchievement, Title: "Ачивки года",
		Description: strings.Join(titles, " • "), Explanation: strings.Join(reasons, " "),
		Payload: AchievementPayload{Codes: codes},
	}
}

func buildMissedOpportunityCard(metrics Metrics, nextAction NextAction) *Card {
	switch nextAction.Code {
	case ActionSaveSearch:
		return &Card{
			ID: "missed-opportunity", Type: CardMissedOpportunity,
			Title:       "Возможность, которая сэкономит время",
			Description: "Сохранённый поиск может сам показывать новые объявления по главному интересу.",
			Explanation: fmt.Sprintf("Просмотров было %d, а сохранённого поиска по главной категории сейчас нет.", metrics.TotalViews),
			Payload:     ActionPayload{Code: nextAction.Code, Target: nextAction.Target},
		}
	case ActionFinishDraft:
		return &Card{
			ID: "missed-opportunity", Type: CardMissedOpportunity,
			Title:       "Шанс довести актуальный черновик до публикации",
			Description: "Текущий черновик можно открыть напрямую.",
			Explanation: "Карточка опирается на актуальное состояние объявлений, а не на разницу годовых счётчиков.",
			Payload:     ActionPayload{Code: nextAction.Code, Target: nextAction.Target},
		}
	}
	return nil
}

func monthName(month uint32) string {
	names := [...]string{"", "январь", "февраль", "март", "апрель", "май", "июнь", "июль", "август", "сентябрь", "октябрь", "ноябрь", "декабрь"}
	if month >= uint32(len(names)) {
		return ""
	}
	return names[month]
}
