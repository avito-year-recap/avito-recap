package achievement

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

type thematicAchievement struct {
	Title         string
	Description   string
	CategoryCodes []string
}

func thematicAchievements() map[model.AchievementCode]thematicAchievement {
	return map[model.AchievementCode]thematicAchievement{
		model.AchievementStyleIcon: {
			Title: "Икона стиля", Description: "Одежда, аксессуары и красота стали заметной частью твоих находок.",
			CategoryCodes: []string{analytics.CategoryWomensFashion, analytics.CategoryBeautyCosmetics},
		},
		model.AchievementFashionableMan: {
			Title: "Модник", Description: "Мужская одежда и детали образа часто оказывались в центре внимания.",
			CategoryCodes: []string{analytics.CategoryMensFashion},
		},
		model.AchievementTraveler: {
			Title: "Путешественник", Description: "Снаряжение для маршрутов и отдыха стало одним из главных интересов года.",
			CategoryCodes: []string{analytics.CategoryOutdoorTravel},
		},
		model.AchievementForTheSoul: {
			Title: "Для души", Description: "Товары для дачи и сада помогали создавать своё уютное пространство.",
			CategoryCodes: []string{analytics.CategoryGarden},
		},
		model.AchievementBookworm: {
			Title: "Книжный червь", Description: "Книги и новые истории регулярно попадали в поле внимания.",
			CategoryCodes: []string{analytics.CategoryBooks},
		},
		model.AchievementBeautyConnoisseur: {
			Title: "Ценитель прекрасного", Description: "Украшения и выразительные детали стали заметным интересом года.",
			CategoryCodes: []string{analytics.CategoryJewelry},
		},
		model.AchievementInTheRhythmOfMusic: {
			Title: "В ритме музыки", Description: "Музыкальные товары задавали особое настроение твоим находкам.",
			CategoryCodes: []string{analytics.CategoryMusic},
		},
		model.AchievementWorldOfPlay: {
			Title: "Мир игры", Description: "Игрушки, куклы и коллекционные находки добавляли году немного волшебства.",
			CategoryCodes: []string{analytics.CategoryToysDolls},
		},
		model.AchievementMasterCraft: {
			Title: "Дело мастера", Description: "Инструменты и полезное оборудование часто становились частью планов.",
			CategoryCodes: []string{analytics.CategoryTools},
		},
		model.AchievementCaringOwner: {
			Title: "Заботливый хозяин", Description: "Товары для животных стали заметной частью твоих выборов.",
			CategoryCodes: []string{analytics.CategoryPets},
		},
		model.AchievementLittleDiscoveries: {
			Title: "Для маленьких открытий", Description: "Детские товары сопровождали новые идеи, игры и открытия.",
			CategoryCodes: []string{analytics.CategoryKids},
		},
		model.AchievementDecisiveStep: {
			Title: "Решительный шаг", Description: "Недвижимость стала заметным направлением года — ты изучал варианты и приближался к важному решению.",
			CategoryCodes: []string{analytics.CategoryRealEstate},
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

func thematicEvidence(metrics model.Metrics, categoryCodes []string, thresholds ruleset.AchievementThresholds) (thematicAchievementEvidence, bool) {
	allowed := make(map[string]struct{}, len(categoryCodes))
	for _, code := range categoryCodes {
		allowed[code] = struct{}{}
	}

	var evidence thematicAchievementEvidence
	var totalSignal uint64
	for _, activity := range metrics.CategoryActivities {
		signal := categoryActivitySignal(activity)
		totalSignal = saturatingAdd(totalSignal, signal)
		if _, ok := allowed[activity.CategoryCode]; !ok {
			continue
		}
		evidence.Categories = append(evidence.Categories, activity.Category)
		evidence.Views = saturatingAdd(evidence.Views, activity.Views)
		evidence.FavoritesAdded = saturatingAdd(evidence.FavoritesAdded, activity.FavoritesAdded)
		evidence.PurchasesCompleted = saturatingAdd(evidence.PurchasesCompleted, activity.PurchasesCompleted)
		evidence.Signal = saturatingAdd(evidence.Signal, signal)
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

func categoryActivitySignal(activity model.CategoryActivity) uint64 {
	return saturatingAdd(
		activity.Views,
		saturatingAdd(
			saturatingMultiply(activity.FavoritesAdded, 4),
			saturatingMultiply(activity.PurchasesCompleted, 12),
		),
	)
}

func saturatingAdd(left, right uint64) uint64 {
	if math.MaxUint64-left < right {
		return math.MaxUint64
	}
	return left + right
}

func saturatingMultiply(value, multiplier uint64) uint64 {
	if multiplier != 0 && value > math.MaxUint64/multiplier {
		return math.MaxUint64
	}
	return value * multiplier
}

func thematicReason(evidence thematicAchievementEvidence) string {
	return fmt.Sprintf(
		"Категории: %s. Просмотров: %d. Добавлений в избранное: %d. Покупок: %d. Доля тематической активности: %.0f%%.",
		strings.Join(evidence.Categories, ", "), evidence.Views, evidence.FavoritesAdded,
		evidence.PurchasesCompleted, evidence.DominanceRate*100,
	)
}
