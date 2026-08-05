package recap

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type behaviorCandidate struct {
	behavior Behavior
	eligible bool
	tieBreak uint32
}

type behaviorRule struct {
	code     BehaviorCode
	tieBreak uint32
	evaluate func(Metrics, BehaviorThresholds) behaviorCandidate
}

// DetectBehavior is a compatibility wrapper around the explicit default ruleset.
func DetectBehavior(metrics Metrics) Behavior {
	return DefaultRuleset().DetectBehavior(metrics)
}

// DetectBehavior evaluates every behavior rule, then selects the highest score.
// Slice order is not business logic: ties are resolved by an explicit tie-break rank.
func (r Ruleset) DetectBehavior(metrics Metrics) Behavior {
	metrics = EnrichMetrics(metrics)
	candidates := make([]behaviorCandidate, 0, len(r.behaviorRules()))
	for _, rule := range r.behaviorRules() {
		candidate := rule.evaluate(metrics, r.Thresholds)
		candidate.tieBreak = rule.tieBreak
		if candidate.eligible {
			candidates = append(candidates, candidate)
		}
	}

	return selectBehaviorCandidate(candidates)
}

func selectBehaviorCandidate(candidates []behaviorCandidate) Behavior {
	if len(candidates) == 0 {
		return Behavior{
			Code:        BehaviorUniversal,
			Title:       "Разные сценарии",
			Description: "В течение года использовались разные возможности площадки без одного доминирующего сценария.",
			Reason:      "Ни один специализированный сценарий не набрал минимальный набор доказательств.",
			Score:       0,
		}
	}

	ordered := append([]behaviorCandidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].behavior.Score != ordered[j].behavior.Score {
			return ordered[i].behavior.Score > ordered[j].behavior.Score
		}
		if ordered[i].tieBreak != ordered[j].tieBreak {
			return ordered[i].tieBreak > ordered[j].tieBreak
		}
		return ordered[i].behavior.Code < ordered[j].behavior.Code
	})
	return ordered[0].behavior
}

func (r Ruleset) behaviorRules() []behaviorRule {
	return []behaviorRule{
		{code: BehaviorResearcher, tieBreak: 10, evaluate: evaluateResearcher},
		{code: BehaviorFindHunter, tieBreak: 20, evaluate: evaluateFindHunter},
		{code: BehaviorStartingSeller, tieBreak: 30, evaluate: evaluateStartingSeller},
		{code: BehaviorDecisiveBuyer, tieBreak: 40, evaluate: evaluateDecisiveBuyer},
		{code: BehaviorActiveSeller, tieBreak: 50, evaluate: evaluateActiveSeller},
	}
}

func evaluateActiveSeller(m Metrics, t BehaviorThresholds) behaviorCandidate {
	if m.ListingsPublished < t.ActiveSellerMinListings || m.SalesCompleted < t.ActiveSellerMinDeals {
		return behaviorCandidate{}
	}
	evidence := []BehaviorEvidence{
		ruleWeightEvidence(120),
		evidenceCount("listings_published", m.ListingsPublished, t.ActiveSellerMinListings, 45,
			fmt.Sprintf("Опубликовано объявлений: %d (порог %d).", m.ListingsPublished, t.ActiveSellerMinListings)),
		evidenceCount("sales_completed", m.SalesCompleted, t.ActiveSellerMinDeals, 55,
			fmt.Sprintf("Завершено продаж: %d (порог %d).", m.SalesCompleted, t.ActiveSellerMinDeals)),
	}
	return behaviorCandidate{eligible: true, behavior: Behavior{
		Code: BehaviorActiveSeller, Title: "Продажи в движении",
		Description: "В течение года было много публикаций и завершённых продаж.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateStartingSeller(m Metrics, t BehaviorThresholds) behaviorCandidate {
	if m.ListingsCreated < t.StartingSellerMinCreated ||
		m.ListingsPublished > t.StartingSellerMaxPublished ||
		m.ListingsCreated <= m.ListingsPublished || m.SalesCompleted != 0 {
		return behaviorCandidate{}
	}
	gap := m.ListingsCreated - m.ListingsPublished
	evidence := []BehaviorEvidence{
		ruleWeightEvidence(80),
		evidenceCount("listings_created", m.ListingsCreated, t.StartingSellerMinCreated, 40,
			fmt.Sprintf("Создано объявлений: %d (порог %d).", m.ListingsCreated, t.StartingSellerMinCreated)),
		evidenceCount("creation_publication_gap", gap, 1, 40,
			fmt.Sprintf("Созданий больше публикаций на %d; это годовой сигнал, а не число текущих черновиков.", gap)),
		{Metric: "sales_completed", Actual: 0, Threshold: 0, Points: 20, Detail: "Завершённых продаж в периоде не было."},
	}
	return behaviorCandidate{eligible: true, behavior: Behavior{
		Code: BehaviorStartingSeller, Title: "Старт в продажах",
		Description: "В течение года объявления создавались чаще, чем публиковались.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateDecisiveBuyer(m Metrics, t BehaviorThresholds) behaviorCandidate {
	if m.PurchasesCompleted < t.DecisiveBuyerMinPurchases ||
		m.ChatsStarted < t.DecisiveBuyerMinChats ||
		m.ChatsWithPurchase < t.DecisiveBuyerMinLinkedChats ||
		m.PurchaseRate < t.DecisiveBuyerMinPurchaseRate {
		return behaviorCandidate{}
	}
	evidence := []BehaviorEvidence{
		ruleWeightEvidence(100),
		evidenceCount("purchases_completed", m.PurchasesCompleted, t.DecisiveBuyerMinPurchases, 25,
			fmt.Sprintf("Завершено покупок: %d (порог %d).", m.PurchasesCompleted, t.DecisiveBuyerMinPurchases)),
		evidenceCount("chats_started", m.ChatsStarted, t.DecisiveBuyerMinChats, 15,
			fmt.Sprintf("Начато диалогов: %d (порог %d).", m.ChatsStarted, t.DecisiveBuyerMinChats)),
		evidenceCount("chats_with_purchase", m.ChatsWithPurchase, t.DecisiveBuyerMinLinkedChats, 30,
			fmt.Sprintf("Диалогов, связанных с покупкой: %d (порог %d).", m.ChatsWithPurchase, t.DecisiveBuyerMinLinkedChats)),
		evidenceRate("purchase_rate", m.PurchaseRate, t.DecisiveBuyerMinPurchaseRate, 30,
			fmt.Sprintf("Покупкой завершилось %.0f%% связанных диалогов (порог %.0f%%).", m.PurchaseRate*100, t.DecisiveBuyerMinPurchaseRate*100)),
	}
	return behaviorCandidate{eligible: true, behavior: Behavior{
		Code: BehaviorDecisiveBuyer, Title: "Решительный покупатель",
		Description: "Несколько диалогов из выбранного периода были связаны с завершёнными покупками.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateFindHunter(m Metrics, t BehaviorThresholds) behaviorCandidate {
	if m.TotalViews < t.FindHunterMinViews || m.FavoritesAdded < t.FindHunterMinFavorites ||
		m.RepeatRate < t.FindHunterMinRepeatRate {
		return behaviorCandidate{}
	}
	evidence := []BehaviorEvidence{
		ruleWeightEvidence(40),
		evidenceCount("total_views", m.TotalViews, t.FindHunterMinViews, 25,
			fmt.Sprintf("Просмотров: %d (порог %d).", m.TotalViews, t.FindHunterMinViews)),
		evidenceCount("favorites_added", m.FavoritesAdded, t.FindHunterMinFavorites, 30,
			fmt.Sprintf("Добавлено в избранное: %d (порог %d).", m.FavoritesAdded, t.FindHunterMinFavorites)),
		evidenceRate("repeat_rate", m.RepeatRate, t.FindHunterMinRepeatRate, 45,
			fmt.Sprintf("Доля повторных просмотров: %.0f%% (порог %.0f%%).", m.RepeatRate*100, t.FindHunterMinRepeatRate*100)),
	}
	return behaviorCandidate{eligible: true, behavior: Behavior{
		Code: BehaviorFindHunter, Title: "Охота за находками",
		Description: "В течение года объявления добавлялись в избранное и часто просматривались повторно.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func evaluateResearcher(m Metrics, t BehaviorThresholds) behaviorCandidate {
	if m.TotalViews < t.ResearcherMinViews || m.CategoriesCount < t.ResearcherMinCategories ||
		m.ChatsStarted > t.ResearcherMaxChats {
		return behaviorCandidate{}
	}
	evidence := []BehaviorEvidence{
		ruleWeightEvidence(20),
		evidenceCount("total_views", m.TotalViews, t.ResearcherMinViews, 40,
			fmt.Sprintf("Просмотров: %d (порог %d).", m.TotalViews, t.ResearcherMinViews)),
		evidenceCount("categories_count", m.CategoriesCount, t.ResearcherMinCategories, 40,
			fmt.Sprintf("Категорий с активностью: %d (порог %d).", m.CategoriesCount, t.ResearcherMinCategories)),
		evidenceInverseCount("chats_started", m.ChatsStarted, t.ResearcherMaxChats, 20,
			fmt.Sprintf("Начато диалогов: %d (допустимый максимум %d).", m.ChatsStarted, t.ResearcherMaxChats)),
	}
	return behaviorCandidate{eligible: true, behavior: Behavior{
		Code: BehaviorResearcher, Title: "Глубокое исследование",
		Description: "За год было много просмотров в разных категориях и сравнительно мало начатых диалогов.",
		Score:       evidenceScore(evidence), Evidence: evidence, Reason: evidenceReason(evidence),
	}}
}

func ruleWeightEvidence(points uint32) BehaviorEvidence {
	return BehaviorEvidence{Metric: "ruleset_weight", Actual: 1, Threshold: 1, Points: points, Detail: "Применён явный вес сценария из версии правил."}
}

func evidenceCount(metric string, actual, threshold uint64, maxPoints uint32, detail string) BehaviorEvidence {
	return BehaviorEvidence{Metric: metric, Actual: float64(actual), Threshold: float64(threshold), Points: scaledPoints(float64(actual), float64(threshold), maxPoints), Detail: detail}
}

func evidenceRate(metric string, actual, threshold float64, maxPoints uint32, detail string) BehaviorEvidence {
	return BehaviorEvidence{Metric: metric, Actual: actual, Threshold: threshold, Points: scaledPoints(actual, threshold, maxPoints), Detail: detail}
}

func evidenceInverseCount(metric string, actual, maximum uint64, maxPoints uint32, detail string) BehaviorEvidence {
	points := maxPoints
	if maximum > 0 && actual > 0 {
		ratio := float64(actual) / float64(maximum)
		points = uint32(math.Round(float64(maxPoints) * math.Max(0, 1-ratio/2)))
	}
	return BehaviorEvidence{Metric: metric, Actual: float64(actual), Threshold: float64(maximum), Points: points, Detail: detail}
}

func scaledPoints(actual, threshold float64, maxPoints uint32) uint32 {
	if threshold <= 0 || actual <= 0 {
		return 0
	}
	value := (actual / threshold) * (float64(maxPoints) / 2)
	if value > float64(maxPoints) {
		value = float64(maxPoints)
	}
	return uint32(math.Round(value))
}

func evidenceScore(evidence []BehaviorEvidence) uint32 {
	var score uint32
	for _, item := range evidence {
		score += item.Points
	}
	return score
}

func evidenceReason(evidence []BehaviorEvidence) string {
	parts := make([]string, 0, len(evidence))
	for _, item := range evidence {
		parts = append(parts, item.Detail)
	}
	return strings.Join(parts, " ")
}
