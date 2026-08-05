package recap

import (
	"fmt"
	"sort"
)

const maxAchievements = 3

type achievementRule struct {
	priority int
	match    func(Metrics) bool
	build    func(Metrics) Achievement
}

func BuildAchievements(metrics Metrics) []Achievement {
	metrics = EnrichMetrics(metrics)
	rules := achievementRules()
	result := make([]Achievement, 0, len(rules))

	for _, rule := range rules {
		if !rule.match(metrics) {
			continue
		}

		achievement := rule.build(metrics)
		achievement.Priority = rule.priority
		result = append(result, achievement)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})

	if len(result) > maxAchievements {
		result = result[:maxAchievements]
	}

	return result
}

func achievementRules() []achievementRule {
	return []achievementRule{
		{
			priority: 110,
			match:    func(m Metrics) bool { return m.SalesCompleted >= 5 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementSuccessfulSeller,
					Title:       "Успешные продажи",
					Description: "Несколько сделок были уверенно доведены до результата.",
					Reason:      fmt.Sprintf("Продаж завершено: %d.", m.SalesCompleted),
					Shareable:   true,
				}
			},
		},
		{
			priority: 100,
			match: func(m Metrics) bool {
				return m.ListingsPublished >= 5 && m.SalesCompleted >= 1
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementConsistentPublisher,
					Title:       "Стабильные публикации",
					Description: "Объявления появлялись регулярно и поддерживали активный сценарий продаж.",
					Reason:      fmt.Sprintf("Объявлений опубликовано: %d.", m.ListingsPublished),
					Shareable:   true,
				}
			},
		},
		{
			priority: 105,
			match: func(m Metrics) bool {
				return m.PurchasesCompleted >= 3
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementDealCloser,
					Title:       "Сделка состоялась",
					Description: "Выбранные варианты несколько раз превращались в завершённые покупки.",
					Reason:      fmt.Sprintf("Покупок завершено: %d.", m.PurchasesCompleted),
					Shareable:   true,
				}
			},
		},
		{
			priority: 95,
			match: func(m Metrics) bool {
				return m.PurchasesCompleted >= 3 && m.PurchaseRate >= decisiveBuyerMinPurchaseRate
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementQuickDecision,
					Title:       "Быстрое решение",
					Description: "Заметная доля начатых диалогов завершилась покупкой.",
					Reason: fmt.Sprintf(
						"Покупкой завершилось %.0f%% начатых диалогов.",
						m.PurchaseRate*100,
					),
					Shareable: true,
				}
			},
		},
		{
			priority: 98,
			match: func(m Metrics) bool {
				return m.CategoriesCount >= 6
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementBroadInterests,
					Title:       "Широкий круг интересов",
					Description: "Активность охватила много разных направлений.",
					Reason:      fmt.Sprintf("Категорий с активностью: %d.", m.CategoriesCount),
					Shareable:   true,
				}
			},
		},
		{
			priority: 90,
			match:    func(m Metrics) bool { return m.TotalViews >= 150 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementAttentiveResearcher,
					Title:       "Внимательное сравнение",
					Description: "Перед следующим шагом было изучено много вариантов.",
					Reason:      fmt.Sprintf("Просмотров объявлений: %d.", m.TotalViews),
					Shareable:   true,
				}
			},
		},
		{
			priority: 80,
			match:    func(m Metrics) bool { return m.FavoritesAdded >= 20 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementMasterOfFavorites,
					Title:       "Коллекция находок",
					Description: "В избранном собрана заметная подборка интересных вариантов.",
					Reason:      fmt.Sprintf("В избранное добавлено: %d.", m.FavoritesAdded),
					Shareable:   true,
				}
			},
		},
		{
			priority: 97,
			match: func(m Metrics) bool {
				return m.PurchasesCompleted >= 1 &&
					m.SalesCompleted >= 1 &&
					m.ListingsPublished >= 1 &&
					m.ChatsStarted >= 3
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementAllRounder,
					Title:       "Всё в одном году",
					Description: "Поиск, общение, покупки и продажи сложились в один разносторонний сценарий.",
					Reason: fmt.Sprintf(
						"Покупок: %d. Продаж: %d. Опубликованных объявлений: %d.",
						m.PurchasesCompleted,
						m.SalesCompleted,
						m.ListingsPublished,
					),
					Shareable: true,
				}
			},
		},
		{
			priority: 96,
			match: func(m Metrics) bool {
				return m.ListingsCreated >= startingSellerMinCreated &&
					m.ListingsCreated > m.ListingsPublished &&
					m.SalesCompleted == 0
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementFirstSellingSteps,
					Title:       "Первые шаги в продажах",
					Description: "Сценарий продажи уже начался с создания собственных объявлений.",
					Reason:      fmt.Sprintf("Объявлений создано: %d.", m.ListingsCreated),
					Shareable:   true,
				}
			},
		},
	}
}
