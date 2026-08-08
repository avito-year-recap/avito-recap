package nextaction

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

func CreateListingAction(reason string) model.NextAction {
	return model.NextAction{
		Code: model.ActionCreateListing, Title: createListingTitle,
		Description: createListingDescription, ButtonText: "Создать объявление",
		Reason: reason, Target: routeActionTarget("/listings/new"),
	}
}

func OpenCategoryAction(metrics model.Metrics, reason string) model.NextAction {
	description := openTopCategoryDescription
	if metrics.TopCategory != "" {
		description = fmt.Sprintf("Вернись в категорию «%s» и проверь новые варианты.", metrics.TopCategory)
	}
	return model.NextAction{
		Code: model.ActionOpenTopCategory, Title: openTopCategoryTitle,
		Description: description, ButtonText: "Открыть категорию",
		Reason: reason, Target: categoryActionTarget(metrics.TopCategoryCode),
	}
}

const (
	finishDraftTitle       = "Заверши начатое объявление"
	finishDraftDescription = "Открой актуальный черновик и подготовь его к публикации."

	continueDialogTitle       = "Продолжи актуальный диалог"
	continueDialogDescription = "Вернись к открытому разговору и продолжи договорённость."

	improveListingTitle       = "Усиль активное объявление"
	improveListingDescription = "Обнови фотографии или описание конкретного активного объявления."

	similarPurchaseTitle       = "Посмотри похожие варианты"
	similarPurchaseDescription = "Открой подборку, похожую на недавнюю покупку."

	saveSearchTitle       = "Сохрани поиск"
	saveSearchDescription = "Новые объявления в главной категории будет проще отслеживать без повторного поиска."

	openFavoritesTitle       = "Вернись к своим находкам"
	openFavoritesDescription = "В избранном есть актуальные варианты, которые можно ещё раз сравнить."

	createFirstListingTitle       = "Опубликуй первое объявление"
	createFirstListingDescription = "Ты уже пробовал создавать объявления — следующий шаг можно начать с новой публикации."

	createListingTitle       = "Создай новое объявление"
	createListingDescription = "Начни новый сценарий продажи с актуального объявления."

	openTopCategoryTitle       = "Посмотри новые предложения"
	openTopCategoryDescription = "Вернись в главную категорию года и проверь новые варианты."

	exploreRecommendationsTitle       = "Посмотри персональные рекомендации"
	exploreRecommendationsDescription = "Открой подборку новых вариантов и выбери актуальный сценарий."
)

func routeActionTarget(route string) model.ActionTarget {
	return model.ActionTarget{Route: &model.RouteTarget{Route: route}}
}
func categoryActionTarget(code string) model.ActionTarget {
	return model.ActionTarget{Category: &model.CategoryTarget{CategoryCode: code}}
}
func listingActionTarget(id uuid.UUID) model.ActionTarget {
	return model.ActionTarget{Listing: &model.ListingTarget{ListingID: id}}
}
func dialogActionTarget(id uuid.UUID) model.ActionTarget {
	return model.ActionTarget{Dialog: &model.DialogTarget{DialogID: id}}
}
func searchActionTarget(code string) model.ActionTarget {
	return model.ActionTarget{Search: &model.SearchTarget{CategoryCode: code}}
}
