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
			priority: 100,
			match:    func(m Metrics) bool { return m.SalesCompleted >= 5 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementSuccessfulSeller,
					Title:       "Успешный продавец",
					Description: "Ты уверенно доводил сделки до результата.",
					Reason:      fmt.Sprintf("За год завершено %d сделок.", m.SalesCompleted),
					Shareable:   true,
				}
			},
		},
		{
			priority: 90,
			match:    func(m Metrics) bool { return m.ListingsPublished >= 5 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementActivePublisher,
					Title:       "Активный автор объявлений",
					Description: "Ты регулярно давал вещам шанс найти нового владельца.",
					Reason:      fmt.Sprintf("Опубликовано %d объявлений.", m.ListingsPublished),
					Shareable:   true,
				}
			},
		},
		{
			priority: 80,
			match:    func(m Metrics) bool { return m.TotalViews >= 150 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementAttentiveResearcher,
					Title:       "Внимательный исследователь",
					Description: "Ты изучил много вариантов перед следующим шагом.",
					Reason:      fmt.Sprintf("Просмотрено %d объявлений.", m.TotalViews),
					Shareable:   true,
				}
			},
		},
		{
			priority: 70,
			match:    func(m Metrics) bool { return m.FavoritesAdded >= 20 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementFavoritesCurator,
					Title:       "Куратор находок",
					Description: "Ты собрал внушительную подборку интересных вариантов.",
					Reason:      fmt.Sprintf("В избранное добавлено %d объявлений.", m.FavoritesAdded),
					Shareable:   true,
				}
			},
		},
		{
			priority: 60,
			match:    func(m Metrics) bool { return m.CategoriesCount >= 6 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementCategoryExplorer,
					Title:       "Исследователь категорий",
					Description: "Твои интересы не ограничивались одним направлением.",
					Reason:      fmt.Sprintf("Активность отмечена в %d категориях.", m.CategoriesCount),
					Shareable:   true,
				}
			},
		},
		{
			priority: 50,
			match:    func(m Metrics) bool { return m.ActiveDays >= 30 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code:        AchievementConsistentUser,
					Title:       "На связи круглый год",
					Description: "Ты регулярно возвращался к своим задачам на площадке.",
					Reason:      fmt.Sprintf("Активность зафиксирована в %d разных дней.", m.ActiveDays),
					Shareable:   false,
				}
			},
		},
	}
}
