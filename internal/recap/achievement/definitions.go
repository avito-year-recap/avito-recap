package achievement

import (
	"fmt"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func achievementDefinitionFor(code model.AchievementCode) (achievementDefinition, bool) {
	definition, ok := achievementDefinitions()[code]
	return definition, ok
}

func achievementDefinitions() map[model.AchievementCode]achievementDefinition {
	definitions := map[model.AchievementCode]achievementDefinition{
		model.AchievementSuccessfulSeller: standardAchievement(
			func(m model.Metrics, r ruleset.Ruleset) bool {
				t := r.AchievementThresholds
				return m.ListingsPublished >= t.SuccessfulSellerMinPublished &&
					sellerSaleRate(m) >= t.SuccessfulSellerMinSaleRate
			},
			func(m model.Metrics) model.Achievement {
				rate := sellerSaleRate(m)
				return model.Achievement{
					Code: model.AchievementSuccessfulSeller, Title: "Мастер переговоров",
					Description: "Большая часть опубликованных объявлений дошла до продажи.",
					Reason: fmt.Sprintf(
						"Продажи составили %.0f%% от числа опубликованных объявлений (%d из %d).",
						rate*100, cappedSalesForRate(m), m.ListingsPublished,
					),
					Shareable: true,
				}
			},
		),
		model.AchievementConsistentPublisher: standardAchievement(
			func(m model.Metrics, r ruleset.Ruleset) bool {
				t := r.AchievementThresholds
				return m.ListingsPublished >= t.ConsistentPublisherMinPublished &&
					sellerSaleRate(m) >= t.ConsistentPublisherMinSaleRate
			},
			func(m model.Metrics) model.Achievement {
				rate := sellerSaleRate(m)
				return model.Achievement{
					Code: model.AchievementConsistentPublisher, Title: "Маяк стабильности",
					Description: "Высокий объём публикаций сочетался со стабильной долей продаж.",
					Reason: fmt.Sprintf(
						"Опубликовано %d объявлений; продажи составили %.0f%% (%d из %d).",
						m.ListingsPublished, rate*100, cappedSalesForRate(m), m.ListingsPublished,
					),
					Shareable: true,
				}
			},
		),
		model.AchievementDealCloser: standardAchievement(
			func(m model.Metrics, _ ruleset.Ruleset) bool { return m.PurchasesCompleted >= 3 },
			func(m model.Metrics) model.Achievement {
				return model.Achievement{
					Code: model.AchievementDealCloser, Title: "Сделка состоялась",
					Description: "Выбранные варианты несколько раз превращались в завершённые покупки.",
					Reason:      fmt.Sprintf("Покупок завершено: %d.", m.PurchasesCompleted), Shareable: true,
				}
			},
		),
		model.AchievementQuickDecision: standardAchievement(
			func(m model.Metrics, r ruleset.Ruleset) bool {
				t := r.Thresholds
				return m.PurchasesCompleted >= t.DecisiveBuyerMinPurchases &&
					m.ChatsStarted >= t.DecisiveBuyerMinChats &&
					m.ChatsWithPurchase >= t.DecisiveBuyerMinLinkedChats &&
					m.PurchaseRate >= t.DecisiveBuyerMinPurchaseRate
			},
			func(m model.Metrics) model.Achievement {
				return model.Achievement{
					Code: model.AchievementQuickDecision, Title: "Молния",
					Description: "Ты быстро переходил от диалога к подходящему выбору.",
					Reason:      fmt.Sprintf("Покупкой завершилось %.0f%% начатых диалогов.", m.PurchaseRate*100),
					Shareable:   true,
				}
			},
		),
		model.AchievementBroadInterests: standardAchievement(
			func(m model.Metrics, _ ruleset.Ruleset) bool { return m.CategoriesCount >= 6 },
			func(m model.Metrics) model.Achievement {
				return model.Achievement{
					Code: model.AchievementBroadInterests, Title: "Человек-оркестр",
					Description: "За год внимание охватило множество разных направлений.",
					Reason:      fmt.Sprintf("Категорий с активностью: %d.", m.CategoriesCount), Shareable: true,
				}
			},
		),
		model.AchievementAttentiveResearcher: standardAchievement(
			func(m model.Metrics, _ ruleset.Ruleset) bool { return m.TotalViews >= 150 },
			func(m model.Metrics) model.Achievement {
				return model.Achievement{
					Code: model.AchievementAttentiveResearcher, Title: "Стратег",
					Description: "Перед следующим шагом было внимательно изучено много вариантов.",
					Reason:      fmt.Sprintf("Просмотров объявлений: %d.", m.TotalViews), Shareable: true,
				}
			},
		),
		model.AchievementMasterOfFavorites: standardAchievement(
			func(m model.Metrics, _ ruleset.Ruleset) bool { return m.FavoritesAdded >= 20 },
			func(m model.Metrics) model.Achievement {
				return model.Achievement{
					Code: model.AchievementMasterOfFavorites, Title: "Собиратель жемчужин",
					Description: "В избранном собралась заметная коллекция интересных находок.",
					Reason:      fmt.Sprintf("В избранное добавлено: %d.", m.FavoritesAdded), Shareable: true,
				}
			},
		),
		model.AchievementAllRounder: standardAchievement(
			func(m model.Metrics, r ruleset.Ruleset) bool {
				return isBalancedSellerBuyer(m, r.AchievementThresholds)
			},
			func(m model.Metrics) model.Achievement {
				return model.Achievement{
					Code: model.AchievementAllRounder, Title: "Человек-швейцарский нож",
					Description: "Покупки и продажи сложились в один по-настоящему универсальный сценарий.",
					Reason:      fmt.Sprintf("Покупок: %d. Продаж: %d.", m.PurchasesCompleted, m.SalesCompleted),
					Shareable:   true,
				}
			},
		),
		model.AchievementFirstSellingSteps: standardAchievement(
			func(m model.Metrics, r ruleset.Ruleset) bool {
				t := r.Thresholds
				draftStart := m.ListingsCreated >= t.StartingSellerMinCreated &&
					m.ListingsCreated > m.ListingsPublished && m.SalesCompleted == 0
				earlySeller := m.SalesCompleted > 0 && m.SalesCompleted < 5
				return draftStart || earlySeller
			},
			func(m model.Metrics) model.Achievement {
				reason := fmt.Sprintf("Объявлений создано: %d.", m.ListingsCreated)
				if m.SalesCompleted > 0 {
					reason = fmt.Sprintf("Продаж завершено: %d.", m.SalesCompleted)
				}
				return model.Achievement{
					Code: model.AchievementFirstSellingSteps, Title: "Начинающий бизнесмен",
					Description: "Продажный сценарий уже начался и набирает первые результаты.",
					Reason:      reason, Shareable: true,
				}
			},
		),
	}

	for code, theme := range thematicAchievements() {
		code := code
		theme := theme
		definitions[code] = achievementDefinition{evaluate: func(m model.Metrics, r ruleset.Ruleset) (model.Achievement, bool) {
			evidence, earned := thematicEvidence(m, theme.CategoryCodes, r.AchievementThresholds)
			if !earned {
				return model.Achievement{}, false
			}
			return model.Achievement{
				Code: code, Title: theme.Title, Description: theme.Description,
				Reason: thematicReason(evidence), Strength: evidence.Signal, Shareable: true,
			}, true
		}}
	}
	return definitions
}

func sellerSaleRate(m model.Metrics) float64 {
	if m.ListingsPublished == 0 {
		return 0
	}
	return float64(cappedSalesForRate(m)) / float64(m.ListingsPublished)
}

// A sale completed this year can theoretically refer to an older listing. For
// this annual recap ratio we cap the numerator at this year's publications so
// the displayed conversion remains a meaningful percentage in [0, 100].
func cappedSalesForRate(m model.Metrics) uint64 {
	if m.SalesCompleted > m.ListingsPublished {
		return m.ListingsPublished
	}
	return m.SalesCompleted
}

func standardAchievement(match func(model.Metrics, ruleset.Ruleset) bool, build func(model.Metrics) model.Achievement) achievementDefinition {
	return achievementDefinition{evaluate: func(m model.Metrics, r ruleset.Ruleset) (model.Achievement, bool) {
		if !match(m, r) {
			return model.Achievement{}, false
		}
		return build(m), true
	}}
}
