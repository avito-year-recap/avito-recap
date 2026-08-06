package recap

import (
	"fmt"
	"sort"
	"strings"
)

// maxAchievements is a hard product invariant. The ruleset may choose a lower
// limit, but no recap can ever award more than three achievements.
const maxAchievements = 3

type achievementDefinition struct {
	evaluate func(Metrics, Ruleset) (Achievement, bool)
}

func BuildAchievements(metrics Metrics) []Achievement {
	return DefaultRuleset().BuildAchievements(metrics)
}

// BuildAchievements evaluates the complete catalogue and then assembles a
// three-slot portfolio. Balanced seller-buyers receive the versatility persona
// plus one selling and one buying award. Seller-only profiles receive exactly
// one strongest seller persona. Other profiles receive the strongest award in
// each category, up to the global limit.
func (r Ruleset) BuildAchievements(metrics Metrics) []Achievement {
	metrics = EnrichMetrics(metrics)
	candidates := make([]Achievement, 0, len(r.AchievementPolicy.Rules))

	for _, configured := range r.AchievementPolicy.Rules {
		definition, ok := achievementDefinitionFor(configured.Code)
		if !ok {
			continue
		}
		candidate, earned := definition.evaluate(metrics, r)
		if !earned {
			continue
		}
		candidate.Category = configured.Category
		candidate.Priority = configured.Priority
		candidates = append(candidates, candidate)
	}

	sort.Slice(candidates, func(i, j int) bool { return achievementLess(candidates[i], candidates[j]) })
	limit := r.AchievementPolicy.MaxAwarded
	if limit > maxAchievements {
		limit = maxAchievements
	}
	if limit < 0 {
		limit = 0
	}
	return selectAchievementPortfolio(metrics, r.AchievementThresholds, candidates, limit)
}

func selectAchievementPortfolio(metrics Metrics, thresholds AchievementThresholds, candidates []Achievement, limit int) []Achievement {
	if limit == 0 || len(candidates) == 0 {
		return nil
	}

	if isBalancedSellerBuyer(metrics, thresholds) {
		selected := make([]Achievement, 0, limit)
		selected = appendFirstCode(selected, candidates, AchievementAllRounder, limit)
		selected = appendBestCategory(selected, candidates, AchievementCategorySelling, limit)
		selected = appendBestCategory(selected, candidates, AchievementCategoryBuying, limit)
		selected = fillFromBestCategories(selected, candidates, limit)
		sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
		return selected
	}

	// A profile that only sold should get one clear seller identity rather than
	// an overloaded list of unrelated awards.
	if metrics.SalesCompleted > 0 && metrics.PurchasesCompleted == 0 {
		selected := appendBestCategory(nil, candidates, AchievementCategorySelling, 1)
		return selected
	}

	// Buyer-only profiles may receive up to three distinct category personas.
	// This makes the recap feel personal without forcing a seller identity onto
	// someone who only used buying scenarios.
	if metrics.PurchasesCompleted > 0 && metrics.SalesCompleted == 0 {
		selected := make([]Achievement, 0, limit)
		for _, candidate := range candidates {
			if len(selected) >= limit {
				break
			}
			if candidate.Category == AchievementCategoryInterest {
				selected = append(selected, candidate)
			}
		}
		selected = fillFromBestCategories(selected, candidates, limit)
		sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
		return selected
	}

	if metrics.SalesCompleted > metrics.PurchasesCompleted {
		selected := appendBestCategory(nil, candidates, AchievementCategorySelling, limit)
		// For seller-dominant mixed profiles, additional earned selling awards are
		// preferred before unrelated categories.
		for _, candidate := range candidates {
			if len(selected) >= limit {
				break
			}
			if candidate.Category == AchievementCategorySelling && !containsAchievement(selected, candidate.Code) {
				selected = append(selected, candidate)
			}
		}
		selected = fillFromBestCategories(selected, candidates, limit)
		sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
		return selected
	}

	selected := fillFromBestCategories(nil, candidates, limit)
	sort.Slice(selected, func(i, j int) bool { return achievementLess(selected[i], selected[j]) })
	return selected
}

func isBalancedSellerBuyer(metrics Metrics, thresholds AchievementThresholds) bool {
	if metrics.PurchasesCompleted < thresholds.BalancedMinPurchases ||
		metrics.SalesCompleted < thresholds.BalancedMinSales {
		return false
	}
	maximum := metrics.PurchasesCompleted
	minimum := metrics.SalesCompleted
	if maximum < minimum {
		maximum, minimum = minimum, maximum
	}
	if maximum == 0 {
		return false
	}
	return float64(maximum-minimum)/float64(maximum) <= thresholds.BalancedMaxDifferenceRate
}

func appendFirstCode(selected, candidates []Achievement, code AchievementCode, limit int) []Achievement {
	if len(selected) >= limit {
		return selected
	}
	for _, candidate := range candidates {
		if candidate.Code == code && !containsAchievement(selected, code) {
			return append(selected, candidate)
		}
	}
	return selected
}

func appendBestCategory(selected, candidates []Achievement, category AchievementCategory, limit int) []Achievement {
	if len(selected) >= limit {
		return selected
	}
	for _, candidate := range candidates {
		if candidate.Category == category && !containsAchievement(selected, candidate.Code) {
			return append(selected, candidate)
		}
	}
	return selected
}

func fillFromBestCategories(selected, candidates []Achievement, limit int) []Achievement {
	seenCategories := make(map[AchievementCategory]struct{}, len(selected))
	for _, achievement := range selected {
		seenCategories[achievement.Category] = struct{}{}
	}
	for _, candidate := range candidates {
		if len(selected) >= limit {
			break
		}
		if containsAchievement(selected, candidate.Code) {
			continue
		}
		if _, exists := seenCategories[candidate.Category]; exists {
			continue
		}
		selected = append(selected, candidate)
		seenCategories[candidate.Category] = struct{}{}
	}
	return selected
}

func containsAchievement(values []Achievement, code AchievementCode) bool {
	for _, value := range values {
		if value.Code == code {
			return true
		}
	}
	return false
}

// achievementLess defines a deterministic total ordering: higher product
// priority wins, then stronger measured evidence, then stable code.
func achievementLess(left, right Achievement) bool {
	if left.Priority != right.Priority {
		return left.Priority > right.Priority
	}
	if left.Strength != right.Strength {
		return left.Strength > right.Strength
	}
	return left.Code < right.Code
}

func achievementDefinitionFor(code AchievementCode) (achievementDefinition, bool) {
	definition, ok := achievementDefinitions()[code]
	return definition, ok
}

func achievementDefinitions() map[AchievementCode]achievementDefinition {
	definitions := map[AchievementCode]achievementDefinition{
		AchievementSuccessfulSeller: standardAchievement(
			func(m Metrics, _ Ruleset) bool { return m.SalesCompleted >= 5 },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementSuccessfulSeller, Title: "Мастер переговоров",
					Description: "Сделки уверенно доходили до результата.",
					Reason:      fmt.Sprintf("Продаж завершено: %d.", m.SalesCompleted), Shareable: true,
				}
			},
		),
		AchievementConsistentPublisher: standardAchievement(
			func(m Metrics, _ Ruleset) bool { return m.ListingsPublished >= 5 && m.SalesCompleted >= 1 },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementConsistentPublisher, Title: "Маяк стабильности",
					Description: "Объявления появлялись регулярно и поддерживали стабильный ритм продаж.",
					Reason:      fmt.Sprintf("Объявлений опубликовано: %d.", m.ListingsPublished), Shareable: true,
				}
			},
		),
		AchievementDealCloser: standardAchievement(
			func(m Metrics, _ Ruleset) bool { return m.PurchasesCompleted >= 3 },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementDealCloser, Title: "Сделка состоялась",
					Description: "Выбранные варианты несколько раз превращались в завершённые покупки.",
					Reason:      fmt.Sprintf("Покупок завершено: %d.", m.PurchasesCompleted), Shareable: true,
				}
			},
		),
		AchievementQuickDecision: standardAchievement(
			func(m Metrics, r Ruleset) bool {
				t := r.Thresholds
				return m.PurchasesCompleted >= t.DecisiveBuyerMinPurchases &&
					m.ChatsStarted >= t.DecisiveBuyerMinChats &&
					m.ChatsWithPurchase >= t.DecisiveBuyerMinLinkedChats &&
					m.PurchaseRate >= t.DecisiveBuyerMinPurchaseRate
			},
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementQuickDecision, Title: "Молния",
					Description: "Ты быстро переходил от диалога к подходящему выбору.",
					Reason:      fmt.Sprintf("Покупкой завершилось %.0f%% начатых диалогов.", m.PurchaseRate*100),
					Shareable:   true,
				}
			},
		),
		AchievementBroadInterests: standardAchievement(
			func(m Metrics, _ Ruleset) bool { return m.CategoriesCount >= 6 },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementBroadInterests, Title: "Человек-оркестр",
					Description: "За год внимание охватило множество разных направлений.",
					Reason:      fmt.Sprintf("Категорий с активностью: %d.", m.CategoriesCount), Shareable: true,
				}
			},
		),
		AchievementAttentiveResearcher: standardAchievement(
			func(m Metrics, _ Ruleset) bool { return m.TotalViews >= 150 },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementAttentiveResearcher, Title: "Стратег",
					Description: "Перед следующим шагом было внимательно изучено много вариантов.",
					Reason:      fmt.Sprintf("Просмотров объявлений: %d.", m.TotalViews), Shareable: true,
				}
			},
		),
		AchievementMasterOfFavorites: standardAchievement(
			func(m Metrics, _ Ruleset) bool { return m.FavoritesAdded >= 20 },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementMasterOfFavorites, Title: "Собиратель жемчужин",
					Description: "В избранном собралась заметная коллекция интересных находок.",
					Reason:      fmt.Sprintf("В избранное добавлено: %d.", m.FavoritesAdded), Shareable: true,
				}
			},
		),
		AchievementAllRounder: standardAchievement(
			func(m Metrics, r Ruleset) bool { return isBalancedSellerBuyer(m, r.AchievementThresholds) },
			func(m Metrics) Achievement {
				return Achievement{
					Code: AchievementAllRounder, Title: "Человек-швейцарский нож",
					Description: "Покупки и продажи сложились в один по-настоящему универсальный сценарий.",
					Reason:      fmt.Sprintf("Покупок: %d. Продаж: %d.", m.PurchasesCompleted, m.SalesCompleted),
					Shareable:   true,
				}
			},
		),
		AchievementFirstSellingSteps: standardAchievement(
			func(m Metrics, r Ruleset) bool {
				t := r.Thresholds
				draftStart := m.ListingsCreated >= t.StartingSellerMinCreated &&
					m.ListingsCreated > m.ListingsPublished && m.SalesCompleted == 0
				earlySeller := m.SalesCompleted > 0 && m.SalesCompleted < 5
				return draftStart || earlySeller
			},
			func(m Metrics) Achievement {
				reason := fmt.Sprintf("Объявлений создано: %d.", m.ListingsCreated)
				if m.SalesCompleted > 0 {
					reason = fmt.Sprintf("Продаж завершено: %d.", m.SalesCompleted)
				}
				return Achievement{
					Code: AchievementFirstSellingSteps, Title: "Начинающий бизнесмен",
					Description: "Продажный сценарий уже начался и набирает первые результаты.",
					Reason:      reason, Shareable: true,
				}
			},
		),
	}

	for code, theme := range thematicAchievements() {
		code := code
		theme := theme
		definitions[code] = achievementDefinition{evaluate: func(m Metrics, r Ruleset) (Achievement, bool) {
			evidence, earned := thematicEvidence(m, theme.CategoryCodes, r.AchievementThresholds)
			if !earned {
				return Achievement{}, false
			}
			return Achievement{
				Code: code, Title: theme.Title, Description: theme.Description,
				Reason: thematicReason(evidence), Strength: evidence.Signal, Shareable: false,
			}, true
		}}
	}
	return definitions
}

func standardAchievement(match func(Metrics, Ruleset) bool, build func(Metrics) Achievement) achievementDefinition {
	return achievementDefinition{evaluate: func(m Metrics, r Ruleset) (Achievement, bool) {
		if !match(m, r) {
			return Achievement{}, false
		}
		return build(m), true
	}}
}

type thematicAchievement struct {
	Title         string
	Description   string
	CategoryCodes []string
}

func thematicAchievements() map[AchievementCode]thematicAchievement {
	return map[AchievementCode]thematicAchievement{
		AchievementStyleIcon: {
			Title: "Икона стиля", Description: "Одежда, аксессуары и красота стали заметной частью твоих находок.",
			CategoryCodes: []string{CategoryWomensFashion, CategoryBeautyCosmetics},
		},
		AchievementFashionableMan: {
			Title: "Модник", Description: "Мужская одежда и детали образа часто оказывались в центре внимания.",
			CategoryCodes: []string{CategoryMensFashion},
		},
		AchievementTraveler: {
			Title: "Путешественник", Description: "Снаряжение для маршрутов и отдыха стало одним из главных интересов года.",
			CategoryCodes: []string{CategoryOutdoorTravel},
		},
		AchievementForTheSoul: {
			Title: "Для души", Description: "Товары для дачи и сада помогали создавать своё уютное пространство.",
			CategoryCodes: []string{CategoryGarden},
		},
		AchievementBookworm: {
			Title: "Книжный червь", Description: "Книги и новые истории регулярно попадали в поле внимания.",
			CategoryCodes: []string{CategoryBooks},
		},
		AchievementBeautyConnoisseur: {
			Title: "Ценитель прекрасного", Description: "Украшения и выразительные детали стали заметным интересом года.",
			CategoryCodes: []string{CategoryJewelry},
		},
		AchievementInTheRhythmOfMusic: {
			Title: "В ритме музыки", Description: "Музыкальные товары задавали особое настроение твоим находкам.",
			CategoryCodes: []string{CategoryMusic},
		},
		AchievementWorldOfPlay: {
			Title: "Мир игры", Description: "Игрушки, куклы и коллекционные находки добавляли году немного волшебства.",
			CategoryCodes: []string{CategoryToysDolls},
		},
		AchievementMasterCraft: {
			Title: "Дело мастера", Description: "Инструменты и полезное оборудование часто становились частью планов.",
			CategoryCodes: []string{CategoryTools},
		},
		AchievementCaringOwner: {
			Title: "Заботливый хозяин", Description: "Товары для животных стали заметной частью твоих выборов.",
			CategoryCodes: []string{CategoryPets},
		},
		AchievementLittleDiscoveries: {
			Title: "Для маленьких открытий", Description: "Детские товары сопровождали новые идеи, игры и открытия.",
			CategoryCodes: []string{CategoryKids},
		},
	}
}

type thematicAchievementEvidence struct {
	Categories         []string
	Views              uint64
	FavoritesAdded     uint64
	PurchasesCompleted uint64
	Signal             uint64
	DominanceRate      float64
}

func thematicEvidence(metrics Metrics, categoryCodes []string, thresholds AchievementThresholds) (thematicAchievementEvidence, bool) {
	allowed := make(map[string]struct{}, len(categoryCodes))
	for _, code := range categoryCodes {
		allowed[code] = struct{}{}
	}

	var evidence thematicAchievementEvidence
	var totalSignal uint64
	for _, activity := range metrics.CategoryActivities {
		signal := categoryActivitySignal(activity)
		totalSignal += signal
		if _, ok := allowed[activity.CategoryCode]; !ok {
			continue
		}
		evidence.Categories = append(evidence.Categories, activity.Category)
		evidence.Views += activity.Views
		evidence.FavoritesAdded += activity.FavoritesAdded
		evidence.PurchasesCompleted += activity.PurchasesCompleted
		evidence.Signal += signal
	}
	if evidence.Signal == 0 || totalSignal == 0 {
		return thematicAchievementEvidence{}, false
	}
	evidence.DominanceRate = float64(evidence.Signal) / float64(totalSignal)
	volumeReached := evidence.Views >= thresholds.ThematicMinViews ||
		evidence.FavoritesAdded >= thresholds.ThematicMinFavorites ||
		evidence.PurchasesCompleted >= thresholds.ThematicMinPurchases
	if !volumeReached || evidence.DominanceRate < thresholds.ThematicMinDominanceRate {
		return thematicAchievementEvidence{}, false
	}
	sort.Strings(evidence.Categories)
	return evidence, true
}

func categoryActivitySignal(activity CategoryActivity) uint64 {
	return activity.Views + activity.FavoritesAdded*4 + activity.PurchasesCompleted*12
}

func thematicReason(evidence thematicAchievementEvidence) string {
	return fmt.Sprintf(
		"Категории: %s. Просмотров: %d. Добавлений в избранное: %d. Покупок: %d. Доля тематической активности: %.0f%%.",
		strings.Join(evidence.Categories, ", "), evidence.Views, evidence.FavoritesAdded,
		evidence.PurchasesCompleted, evidence.DominanceRate*100,
	)
}
