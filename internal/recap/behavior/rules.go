package behavior

import (
	"fmt"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func evaluateActiveSeller(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	checks := []model.RuleCheck{
		checkGTE("listings_published", float64(m.ListingsPublished), float64(t.ActiveSellerMinListings),
			fmt.Sprintf("Опубликовано объявлений: %d (порог %d).", m.ListingsPublished, t.ActiveSellerMinListings)),
		checkGTE("sales_completed", float64(m.SalesCompleted), float64(t.ActiveSellerMinDeals),
			fmt.Sprintf("Завершено продаж: %d (порог %d).", m.SalesCompleted, t.ActiveSellerMinDeals)),
	}
	eligible := allChecksPassed(checks)
	candidate := behaviorCandidate{eligible: eligible, checks: checks}
	if !eligible {
		return candidate
	}
	evidence := evidenceFromChecks(checks)
	candidate.behavior = model.Behavior{
		Code: model.BehaviorActiveSeller, Title: "Продажи в движении",
		Description: "В течение года было много публикаций и завершённых продаж.",
		Evidence:    evidence, Reason: evidenceReason(evidence),
	}
	return candidate
}

func evaluateStartingSeller(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	gap := int64(m.ListingsCreated) - int64(m.ListingsPublished)
	checks := []model.RuleCheck{
		checkGTE("listings_created", float64(m.ListingsCreated), float64(t.StartingSellerMinCreated),
			fmt.Sprintf("Создано объявлений: %d (порог %d).", m.ListingsCreated, t.StartingSellerMinCreated)),
		checkLTE("listings_published", float64(m.ListingsPublished), float64(t.StartingSellerMaxPublished),
			fmt.Sprintf("Опубликовано объявлений: %d (допустимый максимум %d).", m.ListingsPublished, t.StartingSellerMaxPublished)),
		checkGT("creation_publication_gap", float64(gap), 0,
			fmt.Sprintf("Разница между созданными и опубликованными объявлениями: %d (должна быть больше 0).", gap)),
		checkEQ("sales_completed", float64(m.SalesCompleted), 0, "Для стартового продавца завершённых продаж в периоде быть не должно."),
	}
	eligible := allChecksPassed(checks)
	candidate := behaviorCandidate{eligible: eligible, checks: checks}
	if !eligible {
		return candidate
	}
	evidence := evidenceFromChecks(checks)
	candidate.behavior = model.Behavior{
		Code: model.BehaviorStartingSeller, Title: "Старт в продажах",
		Description: "В течение года объявления создавались чаще, чем публиковались.",
		Evidence:    evidence, Reason: evidenceReason(evidence),
	}
	return candidate
}

func evaluateDecisiveBuyer(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	checks := []model.RuleCheck{
		checkGTE("purchases_completed", float64(m.PurchasesCompleted), float64(t.DecisiveBuyerMinPurchases),
			fmt.Sprintf("Завершено покупок: %d (порог %d).", m.PurchasesCompleted, t.DecisiveBuyerMinPurchases)),
		checkGTE("chats_started", float64(m.ChatsStarted), float64(t.DecisiveBuyerMinChats),
			fmt.Sprintf("Начато диалогов: %d (порог %d).", m.ChatsStarted, t.DecisiveBuyerMinChats)),
		checkGTE("chats_with_purchase", float64(m.ChatsWithPurchase), float64(t.DecisiveBuyerMinLinkedChats),
			fmt.Sprintf("Диалогов, связанных с покупкой: %d (порог %d).", m.ChatsWithPurchase, t.DecisiveBuyerMinLinkedChats)),
		checkGTE("purchase_rate", m.PurchaseRate, t.DecisiveBuyerMinPurchaseRate,
			fmt.Sprintf("Покупкой завершилось %.0f%% начатых диалогов (порог %.0f%%).", m.PurchaseRate*100, t.DecisiveBuyerMinPurchaseRate*100)),
	}
	eligible := allChecksPassed(checks)
	candidate := behaviorCandidate{eligible: eligible, checks: checks}
	if !eligible {
		return candidate
	}
	evidence := evidenceFromChecks(checks)
	candidate.behavior = model.Behavior{
		Code: model.BehaviorDecisiveBuyer, Title: "Решительный покупатель",
		Description: "Несколько диалогов из выбранного периода были связаны с завершёнными покупками.",
		Evidence:    evidence, Reason: evidenceReason(evidence),
	}
	return candidate
}

func evaluateFindHunter(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	checks := []model.RuleCheck{
		checkGTE("total_views", float64(m.TotalViews), float64(t.FindHunterMinViews),
			fmt.Sprintf("Просмотров: %d (порог %d).", m.TotalViews, t.FindHunterMinViews)),
		checkGTE("favorites_added", float64(m.FavoritesAdded), float64(t.FindHunterMinFavorites),
			fmt.Sprintf("Добавлено в избранное: %d (порог %d).", m.FavoritesAdded, t.FindHunterMinFavorites)),
		checkGTE("repeat_rate", m.RepeatRate, t.FindHunterMinRepeatRate,
			fmt.Sprintf("Доля повторных просмотров: %.0f%% (порог %.0f%%).", m.RepeatRate*100, t.FindHunterMinRepeatRate*100)),
	}
	eligible := allChecksPassed(checks)
	candidate := behaviorCandidate{eligible: eligible, checks: checks}
	if !eligible {
		return candidate
	}
	evidence := evidenceFromChecks(checks)
	candidate.behavior = model.Behavior{
		Code: model.BehaviorFindHunter, Title: "Искушенный наблюдатель",
		Description: "В течение года объявления добавлялись в избранное и часто просматривались повторно.",
		Evidence:    evidence, Reason: evidenceReason(evidence),
	}
	return candidate
}

func evaluateResearcher(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	checks := []model.RuleCheck{
		checkGTE("total_views", float64(m.TotalViews), float64(t.ResearcherMinViews),
			fmt.Sprintf("Просмотров: %d (порог %d).", m.TotalViews, t.ResearcherMinViews)),
		checkGTE("categories_count", float64(m.CategoriesCount), float64(t.ResearcherMinCategories),
			fmt.Sprintf("Категорий с активностью: %d (порог %d).", m.CategoriesCount, t.ResearcherMinCategories)),
		checkLTE("chats_started", float64(m.ChatsStarted), float64(t.ResearcherMaxChats),
			fmt.Sprintf("Начато диалогов: %d (допустимый максимум %d).", m.ChatsStarted, t.ResearcherMaxChats)),
	}
	eligible := allChecksPassed(checks)
	candidate := behaviorCandidate{eligible: eligible, checks: checks}
	if !eligible {
		return candidate
	}
	evidence := evidenceFromChecks(checks)
	candidate.behavior = model.Behavior{
		Code: model.BehaviorResearcher, Title: "Глубокое исследование",
		Description: "За год было много просмотров в разных категориях и сравнительно мало начатых диалогов.",
		Evidence:    evidence, Reason: evidenceReason(evidence),
	}
	return candidate
}
