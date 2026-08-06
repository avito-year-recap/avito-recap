package nextaction

import (
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
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
