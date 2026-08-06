package nextaction

import "github.com/year-recap/internal/recap/model"

func CreateListingAction(reason string) model.NextAction {
	return model.NextAction{
		Code: model.ActionCreateListing, Title: createdNotPublishedTitle,
		Description: createdNotPublishedDescription, ButtonText: "Создать объявление",
		Reason: reason, Target: routeActionTarget("/listings/new"),
	}
}

// openCategoryAction is kept for compatibility with callers that need a category target.
func OpenCategoryAction(metrics model.Metrics, reason string) model.NextAction {
	return model.NextAction{
		Code: model.ActionOpenTopCategory, Title: viewedWithoutFavoritesTitle,
		Description: viewedWithoutFavoritesDescription, ButtonText: "Смотреть объявления",
		Reason: reason, Target: categoryActionTarget(metrics.TopCategoryCode),
	}
}
