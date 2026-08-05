package recap

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

const publishedForImprove uint64 = 3

type recommendationContext struct {
	metrics  Metrics
	state    ActionableState
	behavior Behavior
}

type recommendationRule struct {
	name     string
	code     ActionCode
	priority int
	match    func(recommendationContext) bool
	build    func(recommendationContext) NextAction
}

// BuildNextAction is a compatibility wrapper. New code should pass an explicit
// ActionableState to Ruleset.BuildNextAction.
func BuildNextAction(metrics Metrics, states ...ActionableState) NextAction {
	state := ActionableState{}
	if len(states) > 0 {
		state = states[0]
	}
	ruleset := DefaultRuleset()
	behavior := ruleset.DetectBehavior(metrics)
	return ruleset.BuildNextAction(metrics, state, behavior)
}

// BuildNextAction evaluates an explicit priority table. Product ordering is
// data in the Ruleset and therefore part of its digest, not slice/switch order.
func (r Ruleset) BuildNextAction(metrics Metrics, state ActionableState, behavior Behavior) NextAction {
	ctx := recommendationContext{
		metrics: EnrichMetrics(metrics), state: normalizeActionableState(state), behavior: behavior,
	}
	rules := r.recommendationRules()
	matched := make([]recommendationRule, 0, len(rules))
	for _, rule := range rules {
		if rule.match(ctx) {
			matched = append(matched, rule)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].priority != matched[j].priority {
			return matched[i].priority > matched[j].priority
		}
		if matched[i].code != matched[j].code {
			return matched[i].code < matched[j].code
		}
		return matched[i].name < matched[j].name
	})
	return matched[0].build(ctx)
}

func (r Ruleset) recommendationRules() []recommendationRule {
	p := r.RecommendationPriorities
	return []recommendationRule{
		{
			name: "finish-current-draft", code: ActionFinishDraft, priority: p.FinishDraft,
			match: func(c recommendationContext) bool {
				return c.state.CurrentDrafts > 0 && c.state.DraftListingID != uuid.Nil
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{Code: ActionFinishDraft, Title: "Заверши начатое объявление",
					Description: "Открой актуальный черновик и подготовь его к публикации.", ButtonText: "Открыть черновик",
					Reason: fmt.Sprintf("Сейчас доступно черновиков: %d. Данные получены из текущего адресуемого snapshot.", c.state.CurrentDrafts),
					Target: listingActionTarget(c.state.DraftListingID)}
			},
		},
		{
			name: "continue-open-dialog", code: ActionContinueDialogs, priority: p.ContinueDialog,
			match: func(c recommendationContext) bool { return c.state.OpenDialogs > 0 && c.state.OpenDialogID != uuid.Nil },
			build: func(c recommendationContext) NextAction {
				return NextAction{Code: ActionContinueDialogs, Title: "Продолжи актуальный диалог",
					Description: "Вернись к открытому разговору и заверши договорённость.", ButtonText: "Открыть диалог",
					Reason: fmt.Sprintf("Сейчас открыто диалогов: %d.", c.state.OpenDialogs), Target: dialogActionTarget(c.state.OpenDialogID)}
			},
		},
		{
			name: "improve-addressable-listing", code: ActionImproveListings, priority: p.ImproveListing,
			match: func(c recommendationContext) bool {
				return c.state.ActiveListings >= publishedForImprove && c.state.ActiveListingID != uuid.Nil
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{Code: ActionImproveListings, Title: "Усиль активное объявление",
					Description: "Обнови фотографии или описание конкретного активного объявления.", ButtonText: "Открыть объявление",
					Reason: fmt.Sprintf("Сейчас активно объявлений: %d; выбран конкретный адресуемый объект.", c.state.ActiveListings),
					Target: listingActionTarget(c.state.ActiveListingID)}
			},
		},
		{
			name: "similar-to-purchase", code: ActionViewSimilarListings, priority: p.SimilarToPurchase,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == BehaviorDecisiveBuyer && c.state.LastPurchasedListingID != uuid.Nil
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{Code: ActionViewSimilarListings, Title: "Посмотри похожие варианты",
					Description: "Открой подборку, похожую на недавнюю покупку.", ButtonText: "Смотреть похожее",
					Reason: "Target связан с последним завершённым сценарием покупки.",
					Target: listingActionTarget(c.state.LastPurchasedListingID)}
			},
		},
		{
			name: "save-research-search", code: ActionSaveSearch, priority: p.SaveSearch,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == BehaviorResearcher && c.metrics.TopCategoryCode != "" && !c.state.HasSavedSearchForTopCategory
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{Code: ActionSaveSearch, Title: "Сохрани поиск",
					Description: "Новые объявления в главной категории будет проще отслеживать без повторного поиска.", ButtonText: "Сохранить поиск",
					Reason: "Просмотров и категорий много, а сохранённого поиска по главному интересу сейчас нет.",
					Target: searchActionTarget(c.metrics.TopCategoryCode)}
			},
		},
		{
			name: "open-current-favorites", code: ActionOpenFavorites, priority: p.OpenFavorites,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == BehaviorFindHunter && c.state.FavoritesCount > 0
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{Code: ActionOpenFavorites, Title: "Вернись к своим находкам",
					Description: "В избранном есть актуальные варианты, которые можно ещё раз сравнить.", ButtonText: "Открыть избранное",
					Reason: fmt.Sprintf("Сейчас в избранном доступно объектов: %d.", c.state.FavoritesCount), Target: routeActionTarget("/favorites")}
			},
		},
		{
			name: "create-for-starting-seller", code: ActionCreateListing, priority: p.CreateForStarter,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == BehaviorStartingSeller && c.state.ActiveListings == 0
			},
			build: func(recommendationContext) NextAction {
				return createListingAction("Исторические метрики подтверждают сценарий продавца, но актуального адресуемого черновика или активного объявления нет.")
			},
		},
		{
			name: "create-for-active-seller", code: ActionCreateListing, priority: p.CreateForSeller,
			match: func(c recommendationContext) bool {
				return c.behavior.Code == BehaviorActiveSeller && c.state.ActiveListings == 0
			},
			build: func(recommendationContext) NextAction {
				return createListingAction("Годовой сценарий продавца подтверждён, а текущих активных объявлений нет.")
			},
		},
		{
			name: "open-top-category", code: ActionOpenTopCategory, priority: p.OpenTopCategory,
			match: func(c recommendationContext) bool { return c.metrics.TopCategoryCode != "" },
			build: func(c recommendationContext) NextAction {
				return openCategoryAction(c.metrics, "Категория была самой просматриваемой за завершённый год и остаётся нейтральным продолжением сценария.")
			},
		},
		{
			name: "neutral-fallback", code: ActionExploreRecommendations, priority: p.NeutralFallback,
			match: func(recommendationContext) bool { return true },
			build: func(recommendationContext) NextAction {
				return NextAction{Code: ActionExploreRecommendations, Title: "Посмотри персональные рекомендации",
					Description: "Открой подборку новых вариантов и выбери актуальный сценарий.", ButtonText: "Открыть рекомендации",
					Reason: "Нет достаточных и адресуемых оснований для более узкого действия.", Target: routeActionTarget("/recommendations")}
			},
		},
	}
}

func createListingAction(reason string) NextAction {
	return NextAction{Code: ActionCreateListing, Title: "Создай новое объявление",
		Description: "Начни новый сценарий продажи с актуального объявления.", ButtonText: "Создать объявление",
		Reason: reason, Target: routeActionTarget("/listings/new")}
}

func openCategoryAction(metrics Metrics, reason string) NextAction {
	return NextAction{Code: ActionOpenTopCategory, Title: "Посмотри новые предложения",
		Description: fmt.Sprintf("Вернись в категорию «%s» и проверь новые варианты.", metrics.TopCategory), ButtonText: "Открыть категорию",
		Reason: reason, Target: categoryActionTarget(metrics.TopCategoryCode)}
}

func routeActionTarget(route string) ActionTarget {
	return ActionTarget{Route: &RouteTarget{Route: route}}
}
func categoryActionTarget(code string) ActionTarget {
	return ActionTarget{Category: &CategoryTarget{CategoryCode: code}}
}
func listingActionTarget(id uuid.UUID) ActionTarget {
	return ActionTarget{Listing: &ListingTarget{ListingID: id}}
}
func dialogActionTarget(id uuid.UUID) ActionTarget {
	return ActionTarget{Dialog: &DialogTarget{DialogID: id}}
}
func searchActionTarget(code string) ActionTarget {
	return ActionTarget{Search: &SearchTarget{CategoryCode: code}}
}
