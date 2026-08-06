package nextaction

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func recommendationRules(r ruleset.Ruleset) []recommendationRule {
	p := r.RecommendationPriorities
	return []recommendationRule{
		{
			name: "created-not-published-current-draft", code: model.ActionFinishDraft, priority: p.FinishDraft,
			match: func(c recommendationContext) bool {
				return hasCreationPublicationGap(c.metrics, r.Thresholds) &&
					c.state.CurrentDrafts > 0 && c.state.DraftListingID != uuid.Nil
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionFinishDraft, Title: createdNotPublishedTitle,
					Description: createdNotPublishedDescription, ButtonText: "Открыть черновик",
					Reason: fmt.Sprintf("Создано объявлений: %d; опубликовано: %d; сейчас доступно черновиков: %d.",
						c.metrics.ListingsCreated, c.metrics.ListingsPublished, c.state.CurrentDrafts),
					Target: listingActionTarget(c.state.DraftListingID),
				}
			},
		},
		{
			name: "created-not-published", code: model.ActionCreateListing, priority: p.ImproveListing,
			match: func(c recommendationContext) bool {
				return hasCreationPublicationGap(c.metrics, r.Thresholds)
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionCreateListing, Title: createdNotPublishedTitle,
					Description: createdNotPublishedDescription, ButtonText: "Создать объявление",
					Reason: fmt.Sprintf("Создано объявлений: %d; опубликовано: %d.",
						c.metrics.ListingsCreated, c.metrics.ListingsPublished),
					Target: routeActionTarget("/listings/new"),
				}
			},
		},
		{
			name: "many-favorites", code: model.ActionOpenFavorites, priority: p.OpenFavorites,
			match: func(c recommendationContext) bool {
				return c.state.FavoritesCount > 0
			},
			build: func(c recommendationContext) model.NextAction {
				reason := fmt.Sprintf("За год добавлено в избранное: %d.", c.metrics.FavoritesAdded)
				if c.state.FavoritesCount > 0 {
					reason = fmt.Sprintf("За год добавлено в избранное: %d; сейчас доступно вариантов: %d.",
						c.metrics.FavoritesAdded, c.state.FavoritesCount)
				}
				return model.NextAction{
					Code: model.ActionOpenFavorites, Title: manyFavoritesTitle,
					Description: manyFavoritesDescription, ButtonText: "Открыть избранное",
					Reason: reason, Target: routeActionTarget("/favorites"),
				}
			},
		},
		{
			name: "viewed-without-favorites", code: model.ActionExploreRecommendations, priority: p.NeutralFallback,
			match: func(recommendationContext) bool { return true },
			build: func(c recommendationContext) model.NextAction {
				reason := fmt.Sprintf("Просмотров: %d; повторных просмотров: %d; добавлений в избранное: %d.",
					c.metrics.TotalViews, c.metrics.RepeatedViews, c.metrics.FavoritesAdded)
				if c.metrics.TopCategoryCode != "" {
					return model.NextAction{
						Code: model.ActionOpenTopCategory, Title: viewedWithoutFavoritesTitle,
						Description: viewedWithoutFavoritesDescription, ButtonText: "Смотреть объявления",
						Reason: reason, Target: categoryActionTarget(c.metrics.TopCategoryCode),
					}
				}
				return model.NextAction{
					Code: model.ActionExploreRecommendations, Title: viewedWithoutFavoritesTitle,
					Description: viewedWithoutFavoritesDescription, ButtonText: "Открыть рекомендации",
					Reason: reason, Target: routeActionTarget("/recommendations"),
				}
			},
		},
	}
}

func hasCreationPublicationGap(metrics model.Metrics, thresholds ruleset.BehaviorThresholds) bool {
	return metrics.ListingsCreated >= thresholds.StartingSellerMinCreated &&
		metrics.ListingsPublished <= thresholds.StartingSellerMaxPublished &&
		metrics.ListingsCreated > metrics.ListingsPublished
}
