package nextaction

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func recommendationRules(r ruleset.Ruleset) []recommendationRule {
	p := r.RecommendationPriorities
	t := r.RecommendationThresholds
	return []recommendationRule{
		{
			name: "finish-current-draft", code: model.ActionFinishDraft, priority: p.FinishDraft,
			match: func(c recommendationContext) bool {
				return c.state.CurrentDrafts > 0 && c.state.DraftListingID != uuid.Nil
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionFinishDraft, Title: finishDraftTitle,
					Description: finishDraftDescription, ButtonText: "Открыть черновик",
					Reason: fmt.Sprintf("Сейчас доступно черновиков: %d; выбран конкретный актуальный черновик.", c.state.CurrentDrafts),
					Target: listingActionTarget(c.state.DraftListingID),
				}
			},
		},
		{
			name: "continue-open-dialog", code: model.ActionContinueDialogs, priority: p.ContinueDialog,
			match: func(c recommendationContext) bool {
				return c.state.OpenDialogs > 0 && c.state.OpenDialogID != uuid.Nil
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionContinueDialogs, Title: continueDialogTitle,
					Description: continueDialogDescription, ButtonText: "Открыть диалог",
					Reason: fmt.Sprintf("Сейчас открыто диалогов: %d.", c.state.OpenDialogs),
					Target: dialogActionTarget(c.state.OpenDialogID),
				}
			},
		},
		{
			name: "improve-addressable-listing", code: model.ActionImproveListings, priority: p.ImproveListing,
			match: func(c recommendationContext) bool {
				return c.state.ActiveListings >= t.ImproveListingsMinActive && c.state.ActiveListingID != uuid.Nil
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionImproveListings, Title: improveListingTitle,
					Description: improveListingDescription, ButtonText: "Открыть объявление",
					Reason: fmt.Sprintf("Сейчас активно объявлений: %d; для улучшения выбран конкретный объект.", c.state.ActiveListings),
					Target: listingActionTarget(c.state.ActiveListingID),
				}
			},
		},
		{
			name: "similar-to-purchase", code: model.ActionViewSimilarListings, priority: p.SimilarToPurchase,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == model.BehaviorDecisiveBuyer && c.state.LastPurchasedListingID != uuid.Nil
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionViewSimilarListings, Title: similarPurchaseTitle,
					Description: similarPurchaseDescription, ButtonText: "Смотреть похожее",
					Reason: "Поведение соответствует решительному покупателю, а текущий snapshot содержит адресуемую последнюю покупку.",
					Target: listingActionTarget(c.state.LastPurchasedListingID),
				}
			},
		},
		{
			name: "save-research-search", code: model.ActionSaveSearch, priority: p.SaveSearch,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == model.BehaviorResearcher && c.metrics.TopCategoryCode != "" && !c.state.HasSavedSearchForTopCategory
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionSaveSearch, Title: saveSearchTitle,
					Description: saveSearchDescription, ButtonText: "Сохранить поиск",
					Reason: "Профиль много исследует предложения, а сохранённого поиска по главной категории сейчас нет.",
					Target: searchActionTarget(c.metrics.TopCategoryCode),
				}
			},
		},
		{
			name: "open-current-favorites", code: model.ActionOpenFavorites, priority: p.OpenFavorites,
			match: func(c recommendationContext) bool {
				return c.state.FavoritesCount > 0
			},
			build: func(c recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionOpenFavorites, Title: openFavoritesTitle,
					Description: openFavoritesDescription, ButtonText: "Открыть избранное",
					Reason: fmt.Sprintf("Сейчас в избранном доступно объектов: %d.", c.state.FavoritesCount),
					Target: routeActionTarget("/favorites"),
				}
			},
		},
		{
			name: "create-first-listing", code: model.ActionCreateFirstListing, priority: p.CreateForStarter,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == model.BehaviorStartingSeller &&
					!c.state.HasEverPublishedListing && c.state.CurrentDrafts == 0 && c.state.ActiveListings == 0
			},
			build: func(recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionCreateFirstListing, Title: createFirstListingTitle,
					Description: createFirstListingDescription, ButtonText: "Создать объявление",
					Reason: "Годовые данные показывают попытки начать продажу, но в текущем состоянии нет опубликованных, активных или черновых объявлений.",
					Target: routeActionTarget("/listings/new"),
				}
			},
		},
		{
			name: "create-for-seller", code: model.ActionCreateListing, priority: p.CreateForSeller,
			match: func(c recommendationContext) bool {
				isSeller := c.behavior.Code == model.BehaviorStartingSeller || c.behavior.Code == model.BehaviorActiveSeller
				return isSeller && c.state.HasEverPublishedListing && c.state.CurrentDrafts == 0 && c.state.ActiveListings == 0
			},
			build: func(recommendationContext) model.NextAction {
				return CreateListingAction("Сценарий продавца подтверждён, но сейчас нет активного объявления или черновика для продолжения.")
			},
		},
		{
			name: "open-top-category", code: model.ActionOpenTopCategory, priority: p.OpenTopCategory,
			match: func(c recommendationContext) bool { return c.metrics.TopCategoryCode != "" },
			build: func(c recommendationContext) model.NextAction {
				return OpenCategoryAction(c.metrics, "Главная категория года остаётся безопасным и понятным продолжением сценария.")
			},
		},
		{
			name: "neutral-fallback", code: model.ActionExploreRecommendations, priority: p.NeutralFallback,
			match: func(recommendationContext) bool { return true },
			build: func(recommendationContext) model.NextAction {
				return model.NextAction{
					Code: model.ActionExploreRecommendations, Title: exploreRecommendationsTitle,
					Description: exploreRecommendationsDescription, ButtonText: "Открыть рекомендации",
					Reason: "Нет достаточных и адресуемых оснований для более узкого действия.",
					Target: routeActionTarget("/recommendations"),
				}
			},
		},
	}
}
