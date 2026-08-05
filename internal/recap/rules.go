package recap

import "fmt"

const (
	activeSellerMinListings      uint64  = 5
	activeSellerMinDeals         uint64  = 3
	startingSellerMinCreated     uint64  = 3
	startingSellerMaxPublished   uint64  = 2
	decisiveBuyerMinPurchases    uint64  = 3
	decisiveBuyerMinChats        uint64  = 5
	decisiveBuyerMinLinkedChats  uint64  = 3
	decisiveBuyerMinPurchaseRate float64 = 0.20
	findHunterMinViews           uint64  = 20
	findHunterMinFavorites       uint64  = 3
	findHunterMinRepeatRate      float64 = 0.20
	researcherMinViews           uint64  = 100
	researcherMinCategories      uint64  = 5
	researcherMaxChats           uint64  = 4
)

// DetectBehavior is deterministic: equal counters and rules version produce equal behavior.
//
// Annual counters can belong to different listings and different event cohorts. For example,
// a listing may be viewed in December and added to favorites in January. Therefore rules must
// not interpret favorites/views, chats/views, published/created, or sales/published as funnels.
// A ratio is used only when its numerator explicitly represents a linked subset of the
// denominator, as ChatsWithPurchase does for ChatsStarted.
func DetectBehavior(metrics Metrics) Behavior {
	metrics = EnrichMetrics(metrics)

	switch {
	case metrics.ListingsPublished >= activeSellerMinListings &&
		metrics.SalesCompleted >= activeSellerMinDeals:
		return Behavior{
			Code:        BehaviorActiveSeller,
			Title:       "Продажи в движении",
			Description: "В течение года было много публикаций и завершённых продаж.",
			Reason: fmt.Sprintf(
				"Объявлений опубликовано: %d. Продаж завершено: %d.",
				metrics.ListingsPublished,
				metrics.SalesCompleted,
			),
		}

	case metrics.ListingsCreated >= startingSellerMinCreated &&
		metrics.ListingsPublished <= startingSellerMaxPublished &&
		metrics.ListingsCreated > metrics.ListingsPublished &&
		metrics.SalesCompleted == 0:
		return Behavior{
			Code:        BehaviorStartingSeller,
			Title:       "Старт в продажах",
			Description: "В течение года объявления создавались чаще, чем публиковались.",
			Reason: fmt.Sprintf(
				"Объявлений создано: %d. Опубликовано: %d. Завершённых продаж: %d.",
				metrics.ListingsCreated,
				metrics.ListingsPublished,
				metrics.SalesCompleted,
			),
		}

	case metrics.PurchasesCompleted >= decisiveBuyerMinPurchases &&
		metrics.ChatsStarted >= decisiveBuyerMinChats &&
		metrics.ChatsWithPurchase >= decisiveBuyerMinLinkedChats &&
		metrics.PurchaseRate >= decisiveBuyerMinPurchaseRate:
		return Behavior{
			Code:        BehaviorDecisiveBuyer,
			Title:       "Решительный покупатель",
			Description: "Несколько диалогов из выбранного периода были связаны с завершёнными покупками.",
			Reason: fmt.Sprintf(
				"Диалогов начато: %d. Диалогов с покупкой: %d. Покупок завершено: %d.",
				metrics.ChatsStarted,
				metrics.ChatsWithPurchase,
				metrics.PurchasesCompleted,
			),
		}

	case metrics.TotalViews >= findHunterMinViews &&
		metrics.FavoritesAdded >= findHunterMinFavorites &&
		metrics.RepeatRate >= findHunterMinRepeatRate:
		return Behavior{
			Code:        BehaviorFindHunter,
			Title:       "Охота за находками",
			Description: "В течение года объявления добавлялись в избранное и часто просматривались повторно.",
			Reason: fmt.Sprintf(
				"В избранное добавлено: %d. Повторных просмотров: %d из %d.",
				metrics.FavoritesAdded,
				metrics.RepeatedViews,
				metrics.TotalViews,
			),
		}

	case metrics.TotalViews >= researcherMinViews &&
		metrics.CategoriesCount >= researcherMinCategories &&
		metrics.ChatsStarted <= researcherMaxChats:
		return Behavior{
			Code:        BehaviorResearcher,
			Title:       "Глубокое исследование",
			Description: "За год было много просмотров в разных категориях и сравнительно мало начатых диалогов.",
			Reason: fmt.Sprintf(
				"Просмотров: %d. Категорий с активностью: %d. Диалогов начато: %d.",
				metrics.TotalViews,
				metrics.CategoriesCount,
				metrics.ChatsStarted,
			),
		}

	default:
		return Behavior{
			Code:        BehaviorUniversal,
			Title:       "Разные сценарии",
			Description: "В течение года использовались разные возможности площадки без одного доминирующего сценария.",
			Reason:      "Активность распределилась между поиском, общением, покупками и продажами.",
		}
	}
}
