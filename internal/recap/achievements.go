package recap

import (
	"fmt"
	"sort"
)

// maxAchievements is a hard product invariant. The ruleset may choose a lower
// limit, but no recap can ever award more than three achievements.
const maxAchievements = 3

type achievementDefinition struct {
	match func(Metrics, BehaviorThresholds) bool
	build func(Metrics) Achievement
}

func BuildAchievements(metrics Metrics) []Achievement {
	return DefaultRuleset().BuildAchievements(metrics)
}

// BuildAchievements evaluates the complete achievement catalogue, keeps only
// the strongest earned grade in each category, then awards at most the ruleset
// limit globally. Both stages use code as an explicit deterministic tie-break.
func (r Ruleset) BuildAchievements(metrics Metrics) []Achievement {
	metrics = EnrichMetrics(metrics)
	bestByCategory := make(map[AchievementCategory]Achievement)

	for _, configured := range r.AchievementPolicy.Rules {
		definition, ok := achievementDefinitionFor(configured.Code)
		if !ok || !definition.match(metrics, r.Thresholds) {
			continue
		}

		candidate := definition.build(metrics)
		candidate.Category = configured.Category
		candidate.Priority = configured.Priority

		current, exists := bestByCategory[candidate.Category]
		if !exists || achievementLess(candidate, current) {
			bestByCategory[candidate.Category] = candidate
		}
	}

	result := make([]Achievement, 0, len(bestByCategory))
	for _, achievement := range bestByCategory {
		result = append(result, achievement)
	}
	sort.Slice(result, func(i, j int) bool { return achievementLess(result[i], result[j]) })

	limit := r.AchievementPolicy.MaxAwarded
	if limit > maxAchievements {
		limit = maxAchievements
	}
	if limit < 0 {
		limit = 0
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

// achievementLess defines the total ordering used both inside a category and
// for the final award list: higher priority wins; equal priority uses the
// lexicographically smaller stable achievement code.
func achievementLess(left, right Achievement) bool {
	if left.Priority != right.Priority {
		return left.Priority > right.Priority
	}
	return left.Code < right.Code
}

func achievementDefinitionFor(code AchievementCode) (achievementDefinition, bool) {
	definition, ok := achievementDefinitions()[code]
	return definition, ok
}

func achievementDefinitions() map[AchievementCode]achievementDefinition {
	return map[AchievementCode]achievementDefinition{
		AchievementSuccessfulSeller: {
			match: func(m Metrics, _ BehaviorThresholds) bool { return m.SalesCompleted >= 5 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementSuccessfulSeller, Title: "Успешные продажи",
					Description: "Несколько сделок были уверенно доведены до результата.",
					Reason:      fmt.Sprintf("Продаж завершено: %d.", m.SalesCompleted), Shareable: true,
				}
			},
		},
		AchievementConsistentPublisher: {
			match: func(m Metrics, _ BehaviorThresholds) bool { return m.ListingsPublished >= 5 && m.SalesCompleted >= 1 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementConsistentPublisher, Title: "Стабильные публикации",
					Description: "Объявления появлялись регулярно и поддерживали активный сценарий продаж.",
					Reason:      fmt.Sprintf("Объявлений опубликовано: %d.", m.ListingsPublished), Shareable: true,
				}
			},
		},
		AchievementDealCloser: {
			match: func(m Metrics, _ BehaviorThresholds) bool { return m.PurchasesCompleted >= 3 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementDealCloser, Title: "Сделка состоялась",
					Description: "Выбранные варианты несколько раз превращались в завершённые покупки.",
					Reason:      fmt.Sprintf("Покупок завершено: %d.", m.PurchasesCompleted), Shareable: true,
				}
			},
		},
		AchievementQuickDecision: {
			match: func(m Metrics, t BehaviorThresholds) bool {
				return m.PurchasesCompleted >= t.DecisiveBuyerMinPurchases &&
					m.ChatsStarted >= t.DecisiveBuyerMinChats &&
					m.ChatsWithPurchase >= t.DecisiveBuyerMinLinkedChats &&
					m.PurchaseRate >= t.DecisiveBuyerMinPurchaseRate
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementQuickDecision, Title: "Быстрое решение",
					Description: "Заметная доля начатых диалогов завершилась покупкой.",
					Reason:      fmt.Sprintf("Покупкой завершилось %.0f%% начатых диалогов.", m.PurchaseRate*100), Shareable: true,
				}
			},
		},
		AchievementBroadInterests: {
			match: func(m Metrics, _ BehaviorThresholds) bool { return m.CategoriesCount >= 6 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementBroadInterests, Title: "Широкий круг интересов",
					Description: "Активность охватила много разных направлений.",
					Reason:      fmt.Sprintf("Категорий с активностью: %d.", m.CategoriesCount), Shareable: true,
				}
			},
		},
		AchievementAttentiveResearcher: {
			match: func(m Metrics, _ BehaviorThresholds) bool { return m.TotalViews >= 150 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementAttentiveResearcher, Title: "Внимательное сравнение",
					Description: "Перед следующим шагом было изучено много вариантов.",
					Reason:      fmt.Sprintf("Просмотров объявлений: %d.", m.TotalViews), Shareable: true,
				}
			},
		},
		AchievementMasterOfFavorites: {
			match: func(m Metrics, _ BehaviorThresholds) bool { return m.FavoritesAdded >= 20 },
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementMasterOfFavorites, Title: "Коллекция находок",
					Description: "В избранном собрана заметная подборка интересных вариантов.",
					Reason:      fmt.Sprintf("В избранное добавлено: %d.", m.FavoritesAdded), Shareable: true,
				}
			},
		},
		AchievementAllRounder: {
			match: func(m Metrics, _ BehaviorThresholds) bool {
				return m.PurchasesCompleted >= 1 && m.SalesCompleted >= 1 &&
					m.ListingsPublished >= 1 && m.ChatsStarted >= 3
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementAllRounder, Title: "Всё в одном году",
					Description: "Поиск, общение, покупки и продажи сложились в один разносторонний сценарий.",
					Reason:      fmt.Sprintf("Покупок: %d. Продаж: %d. Опубликованных объявлений: %d.", m.PurchasesCompleted, m.SalesCompleted, m.ListingsPublished),
					Shareable:   true,
				}
			},
		},
		AchievementFirstSellingSteps: {
			match: func(m Metrics, t BehaviorThresholds) bool {
				return m.ListingsCreated >= t.StartingSellerMinCreated &&
					m.ListingsCreated > m.ListingsPublished && m.SalesCompleted == 0
			},
			build: func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementFirstSellingSteps, Title: "Первые шаги в продажах",
					Description: "Сценарий продажи уже начался с создания собственных объявлений.",
					Reason:      fmt.Sprintf("Объявлений создано: %d.", m.ListingsCreated), Shareable: true,
				}
			},
		},
	}
}
