package recap

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

const (
	manyFavoritesTitle       = "Твои фавориты всё ещё ждут"
	manyFavoritesDescription = "В избранном осталось несколько актуальных вариантов. Возможно, среди них всё ещё есть подходящий."

	createdNotPublishedTitle       = "Объявления почти готовы"
	createdNotPublishedDescription = "Часть созданных объявлений не дошла до публикации."

	viewedWithoutFavoritesTitle       = "Кажется, тебе это понравилось"
	viewedWithoutFavoritesDescription = "К некоторым объявлениям ты возвращался несколько раз. Добавляй их в избранное, чтобы сравнивать в одном месте."
)

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

// BuildNextAction evaluates an explicit priority table. The user-facing output
// is deliberately restricted to three product-approved variants.
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
			name: "created-not-published-current-draft", code: ActionFinishDraft, priority: p.FinishDraft,
			match: func(c recommendationContext) bool {
				return hasCreationPublicationGap(c.metrics, r.Thresholds) &&
					c.state.CurrentDrafts > 0 && c.state.DraftListingID != uuid.Nil
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{
					Code: ActionFinishDraft, Title: createdNotPublishedTitle,
					Description: createdNotPublishedDescription, ButtonText: "Открыть черновик",
					Reason: fmt.Sprintf("Создано объявлений: %d; опубликовано: %d; сейчас доступно черновиков: %d.",
						c.metrics.ListingsCreated, c.metrics.ListingsPublished, c.state.CurrentDrafts),
					Target: listingActionTarget(c.state.DraftListingID),
				}
			},
		},
		{
			name: "created-not-published", code: ActionCreateListing, priority: p.ImproveListing,
			match: func(c recommendationContext) bool {
				return hasCreationPublicationGap(c.metrics, r.Thresholds)
			},
			build: func(c recommendationContext) NextAction {
				return NextAction{
					Code: ActionCreateListing, Title: createdNotPublishedTitle,
					Description: createdNotPublishedDescription, ButtonText: "Создать объявление",
					Reason: fmt.Sprintf("Создано объявлений: %d; опубликовано: %d.",
						c.metrics.ListingsCreated, c.metrics.ListingsPublished),
					Target: routeActionTarget("/listings/new"),
				}
			},
		},
		{
			name: "many-favorites", code: ActionOpenFavorites, priority: p.OpenFavorites,
			match: func(c recommendationContext) bool {
				return c.state.FavoritesCount > 0
			},
			build: func(c recommendationContext) NextAction {
				reason := fmt.Sprintf("За год добавлено в избранное: %d.", c.metrics.FavoritesAdded)
				if c.state.FavoritesCount > 0 {
					reason = fmt.Sprintf("За год добавлено в избранное: %d; сейчас доступно вариантов: %d.",
						c.metrics.FavoritesAdded, c.state.FavoritesCount)
				}
				return NextAction{
					Code: ActionOpenFavorites, Title: manyFavoritesTitle,
					Description: manyFavoritesDescription, ButtonText: "Открыть избранное",
					Reason: reason, Target: routeActionTarget("/favorites"),
				}
			},
		},
		{
			name: "viewed-without-favorites", code: ActionExploreRecommendations, priority: p.NeutralFallback,
			match: func(recommendationContext) bool { return true },
			build: func(c recommendationContext) NextAction {
				reason := fmt.Sprintf("Просмотров: %d; повторных просмотров: %d; добавлений в избранное: %d.",
					c.metrics.TotalViews, c.metrics.RepeatedViews, c.metrics.FavoritesAdded)
				if c.metrics.TopCategoryCode != "" {
					return NextAction{
						Code: ActionOpenTopCategory, Title: viewedWithoutFavoritesTitle,
						Description: viewedWithoutFavoritesDescription, ButtonText: "Смотреть объявления",
						Reason: reason, Target: categoryActionTarget(c.metrics.TopCategoryCode),
					}
				}
				return NextAction{
					Code: ActionExploreRecommendations, Title: viewedWithoutFavoritesTitle,
					Description: viewedWithoutFavoritesDescription, ButtonText: "Открыть рекомендации",
					Reason: reason, Target: routeActionTarget("/recommendations"),
				}
			},
		},
	}
}

func hasCreationPublicationGap(metrics Metrics, thresholds BehaviorThresholds) bool {
	return metrics.ListingsCreated >= thresholds.StartingSellerMinCreated &&
		metrics.ListingsPublished <= thresholds.StartingSellerMaxPublished &&
		metrics.ListingsCreated > metrics.ListingsPublished
}

// createListingAction is kept for compatibility with integrity and hardening tests.
func createListingAction(reason string) NextAction {
	return NextAction{
		Code: ActionCreateListing, Title: createdNotPublishedTitle,
		Description: createdNotPublishedDescription, ButtonText: "Создать объявление",
		Reason: reason, Target: routeActionTarget("/listings/new"),
	}
}

// openCategoryAction is kept for compatibility with callers that need a category target.
func openCategoryAction(metrics Metrics, reason string) NextAction {
	return NextAction{
		Code: ActionOpenTopCategory, Title: viewedWithoutFavoritesTitle,
		Description: viewedWithoutFavoritesDescription, ButtonText: "Смотреть объявления",
		Reason: reason, Target: categoryActionTarget(metrics.TopCategoryCode),
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
