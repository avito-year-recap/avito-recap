package recap

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	publishedForImprove uint64 = 3
)

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

// BuildNextAction uses historical metrics only for context and current state for
// executable unfinished work. Every returned action has a structured target.
func (r Ruleset) BuildNextAction(metrics Metrics, state ActionableState, behavior Behavior) NextAction {
	metrics = EnrichMetrics(metrics)
	state = normalizeActionableState(state)

	if state.CurrentDrafts > 0 && state.DraftListingID != uuid.Nil {
		return NextAction{
			Code:        ActionFinishDraft,
			Title:       "Заверши начатое объявление",
			Description: "Открой актуальный черновик и подготовь его к публикации.",
			ButtonText:  "Открыть черновик",
			Reason:      fmt.Sprintf("Сейчас доступно черновиков: %d. Рекомендация основана на текущем состоянии, а не на разнице годовых счётчиков.", state.CurrentDrafts),
			Target:      listingActionTarget(state.DraftListingID),
		}
	}

	if state.OpenDialogs > 0 && state.OpenDialogID != uuid.Nil {
		return NextAction{
			Code:        ActionContinueDialogs,
			Title:       "Продолжи актуальный диалог",
			Description: "Вернись к открытому разговору и заверши договорённость.",
			ButtonText:  "Открыть диалог",
			Reason:      fmt.Sprintf("Сейчас открыто диалогов: %d.", state.OpenDialogs),
			Target:      dialogActionTarget(state.OpenDialogID),
		}
	}

	switch behavior.Code {
	case BehaviorActiveSeller:
		return createListingAction("В течение года было много публикаций и несколько завершённых продаж.")

	case BehaviorDecisiveBuyer:
		if state.LastPurchasedListingID != uuid.Nil {
			return NextAction{
				Code:        ActionViewSimilarListings,
				Title:       "Посмотри похожие варианты",
				Description: "Открой подборку, похожую на недавнюю покупку.",
				ButtonText:  "Смотреть похожее",
				Reason:      "Текущий target связан с последним завершённым сценарием покупки.",
				Target:      listingActionTarget(state.LastPurchasedListingID),
			}
		}
		if metrics.TopCategoryCode != "" {
			return openCategoryAction(metrics, "Главная категория года даёт безопасный контекст для следующего выбора.")
		}

	case BehaviorResearcher:
		if metrics.TopCategoryCode != "" && !state.HasSavedSearchForTopCategory {
			return NextAction{
				Code:        ActionSaveSearch,
				Title:       "Сохрани поиск",
				Description: "Новые объявления в главной категории будет проще отслеживать без повторного поиска.",
				ButtonText:  "Сохранить поиск",
				Reason:      "Просмотров и категорий много, а сохранённого поиска по главному интересу сейчас нет.",
				Target:      searchActionTarget(metrics.TopCategoryCode),
			}
		}
		if metrics.TopCategoryCode != "" {
			return openCategoryAction(metrics, "Поиск уже сохранён, поэтому следующий полезный шаг — открыть свежие предложения.")
		}

	case BehaviorFindHunter:
		if state.FavoritesCount > 0 {
			return NextAction{
				Code:        ActionOpenFavorites,
				Title:       "Вернись к своим находкам",
				Description: "В избранном есть актуальные варианты, которые можно ещё раз сравнить.",
				ButtonText:  "Открыть избранное",
				Reason:      fmt.Sprintf("Сейчас в избранном доступно объектов: %d.", state.FavoritesCount),
				Target:      routeActionTarget("/favorites"),
			}
		}
		if metrics.TopCategoryCode != "" {
			return openCategoryAction(metrics, "Исторический интерес сохранился, но актуальное избранное пусто.")
		}

	case BehaviorStartingSeller:
		// Historical creation/publication gaps do not prove that a draft still exists.
		return createListingAction("Текущих адресуемых черновиков нет, поэтому предлагается новый сценарий создания объявления.")
	}

	if state.ActiveListings >= publishedForImprove && state.ActiveListingID != uuid.Nil && metrics.SalesCompleted == 0 {
		return NextAction{
			Code:        ActionImproveListings,
			Title:       "Усиль активное объявление",
			Description: "Обнови фотографии или описание конкретного активного объявления.",
			ButtonText:  "Открыть объявление",
			Reason:      fmt.Sprintf("Сейчас активно объявлений: %d, а завершённых продаж в выбранном году не было.", state.ActiveListings),
			Target:      listingActionTarget(state.ActiveListingID),
		}
	}

	if metrics.TopCategoryCode != "" {
		return openCategoryAction(metrics, "Эта категория была самой просматриваемой за год.")
	}

	if !state.HasEverPublishedListing {
		return createListingAction("Текущее состояние подтверждает, что профиль ещё не публиковал объявления.")
	}

	return NextAction{
		Code:        ActionExploreRecommendations,
		Title:       "Посмотри персональные рекомендации",
		Description: "Открой подборку новых вариантов и выбери актуальный сценарий.",
		ButtonText:  "Открыть рекомендации",
		Reason:      "Ни исторические метрики, ни текущее состояние не дают основания для более узкого действия.",
		Target:      routeActionTarget("/recommendations"),
	}
}

func createListingAction(reason string) NextAction {
	return NextAction{
		Code:        ActionCreateListing,
		Title:       "Создай новое объявление",
		Description: "Начни новый сценарий продажи с актуального объявления.",
		ButtonText:  "Создать объявление",
		Reason:      reason,
		Target:      routeActionTarget("/listings/new"),
	}
}

func openCategoryAction(metrics Metrics, reason string) NextAction {
	return NextAction{
		Code:        ActionOpenTopCategory,
		Title:       "Посмотри новые предложения",
		Description: fmt.Sprintf("Вернись в категорию «%s» и проверь новые варианты.", metrics.TopCategory),
		ButtonText:  "Открыть категорию",
		Reason:      reason,
		Target:      categoryActionTarget(metrics.TopCategoryCode),
	}
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
