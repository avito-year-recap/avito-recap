package nextaction

import (
	"fmt"

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
