package behavior

import (
	"fmt"

	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func evaluateActiveSeller(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	if m.ListingsPublished < t.ActiveSellerMinListings || m.SalesCompleted < t.ActiveSellerMinDeals {
		return behaviorCandidate{}
	}
	evidence := []model.BehaviorEvidence{
		evidenceCount("listings_published", m.ListingsPublished, t.ActiveSellerMinListings, 45,
			fmt.Sprintf("Опубликовано объявлений: %d (порог %d).", m.ListingsPublished, t.ActiveSellerMinListings)),
		evidenceCount("sales_completed", m.SalesCompleted, t.ActiveSellerMinDeals, 55,
			fmt.Sprintf("Завершено продаж: %d (порог %d).", m.SalesCompleted, t.ActiveSellerMinDeals)),
	}
	return behaviorCandidate{eligible: true, behavior: model.Behavior{
		Code: model.BehaviorActiveSeller, Title: "Продажи в движении",
		Description: "В течение года было много публикаций и завершённых продаж.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateStartingSeller(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	if m.ListingsCreated < t.StartingSellerMinCreated ||
		m.ListingsPublished > t.StartingSellerMaxPublished ||
		m.ListingsCreated <= m.ListingsPublished || m.SalesCompleted != 0 {
		return behaviorCandidate{}
	}
	gap := m.ListingsCreated - m.ListingsPublished
	evidence := []model.BehaviorEvidence{
		evidenceCount("listings_created", m.ListingsCreated, t.StartingSellerMinCreated, 40,
			fmt.Sprintf("Создано объявлений: %d (порог %d).", m.ListingsCreated, t.StartingSellerMinCreated)),
		evidenceCount("creation_publication_gap", gap, 1, 40,
			fmt.Sprintf("Созданий больше публикаций на %d; это годовой сигнал, а не число текущих черновиков.", gap)),
		{Metric: "sales_completed", Actual: 0, Threshold: 0, Points: 20, Detail: "Завершённых продаж в периоде не было."},
	}
	return behaviorCandidate{eligible: true, behavior: model.Behavior{
		Code: model.BehaviorStartingSeller, Title: "Старт в продажах",
		Description: "В течение года объявления создавались чаще, чем публиковались.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateDecisiveBuyer(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	if m.PurchasesCompleted < t.DecisiveBuyerMinPurchases ||
		m.ChatsStarted < t.DecisiveBuyerMinChats ||
		m.ChatsWithPurchase < t.DecisiveBuyerMinLinkedChats ||
		m.PurchaseRate < t.DecisiveBuyerMinPurchaseRate {
		return behaviorCandidate{}
	}
	evidence := []model.BehaviorEvidence{
		evidenceCount("purchases_completed", m.PurchasesCompleted, t.DecisiveBuyerMinPurchases, 25,
			fmt.Sprintf("Завершено покупок: %d (порог %d).", m.PurchasesCompleted, t.DecisiveBuyerMinPurchases)),
		evidenceCount("chats_started", m.ChatsStarted, t.DecisiveBuyerMinChats, 15,
			fmt.Sprintf("Начато диалогов: %d (порог %d).", m.ChatsStarted, t.DecisiveBuyerMinChats)),
		evidenceCount("chats_with_purchase", m.ChatsWithPurchase, t.DecisiveBuyerMinLinkedChats, 30,
			fmt.Sprintf("Диалогов, связанных с покупкой: %d (порог %d).", m.ChatsWithPurchase, t.DecisiveBuyerMinLinkedChats)),
		evidenceRate("purchase_rate", m.PurchaseRate, t.DecisiveBuyerMinPurchaseRate, 30,
			fmt.Sprintf("Покупкой завершилось %.0f%% начатых диалогов (порог %.0f%%).", m.PurchaseRate*100, t.DecisiveBuyerMinPurchaseRate*100)),
	}
	return behaviorCandidate{eligible: true, behavior: model.Behavior{
		Code: model.BehaviorDecisiveBuyer, Title: "Решительный покупатель",
		Description: "Несколько диалогов из выбранного периода были связаны с завершёнными покупками.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateFindHunter(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	if m.TotalViews < t.FindHunterMinViews || m.FavoritesAdded < t.FindHunterMinFavorites ||
		m.RepeatRate < t.FindHunterMinRepeatRate {
		return behaviorCandidate{}
	}
	evidence := []model.BehaviorEvidence{
		evidenceCount("total_views", m.TotalViews, t.FindHunterMinViews, 25,
			fmt.Sprintf("Просмотров: %d (порог %d).", m.TotalViews, t.FindHunterMinViews)),
		evidenceCount("favorites_added", m.FavoritesAdded, t.FindHunterMinFavorites, 30,
			fmt.Sprintf("Добавлено в избранное: %d (порог %d).", m.FavoritesAdded, t.FindHunterMinFavorites)),
		evidenceRate("repeat_rate", m.RepeatRate, t.FindHunterMinRepeatRate, 45,
			fmt.Sprintf("Доля повторных просмотров: %.0f%% (порог %.0f%%).", m.RepeatRate*100, t.FindHunterMinRepeatRate*100)),
	}
	return behaviorCandidate{eligible: true, behavior: model.Behavior{
		Code: model.BehaviorFindHunter, Title: "Искушенный наблюдатель",
		Description: "В течение года объявления добавлялись в избранное и часто просматривались повторно.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateResearcher(m model.Metrics, t ruleset.BehaviorThresholds) behaviorCandidate {
	if m.TotalViews < t.ResearcherMinViews || m.CategoriesCount < t.ResearcherMinCategories ||
		m.ChatsStarted > t.ResearcherMaxChats {
		return behaviorCandidate{}
	}
	evidence := []model.BehaviorEvidence{
		evidenceCount("total_views", m.TotalViews, t.ResearcherMinViews, 40,
			fmt.Sprintf("Просмотров: %d (порог %d).", m.TotalViews, t.ResearcherMinViews)),
		evidenceCount("categories_count", m.CategoriesCount, t.ResearcherMinCategories, 40,
			fmt.Sprintf("Категорий с активностью: %d (порог %d).", m.CategoriesCount, t.ResearcherMinCategories)),
		evidenceInverseCount("chats_started", m.ChatsStarted, t.ResearcherMaxChats, 20,
			fmt.Sprintf("Начато диалогов: %d (допустимый максимум %d).", m.ChatsStarted, t.ResearcherMaxChats)),
	}
	return behaviorCandidate{eligible: true, behavior: model.Behavior{
		Code: model.BehaviorResearcher, Title: "Глубокое исследование",
		Description: "За год было много просмотров в разных категориях и сравнительно мало начатых диалогов.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}
