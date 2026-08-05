package recap

import "fmt"

const (
	activeSellerMinListings          uint64  = 5
	activeSellerMinDeals             uint64  = 3
	startingSellerMinCreated         uint64  = 3
	startingSellerMinDrafts          uint64  = 2
	startingSellerMaxPublicationRate float64 = 0.60
	decisiveBuyerMinPurchases        uint64  = 3
	decisiveBuyerMinChats            uint64  = 5
	decisiveBuyerMinPurchaseRate     float64 = 0.20
	findHunterMinViews               uint64  = 20
	findHunterMinFavorite            float64 = 0.15
	findHunterMinRepeat              float64 = 0.20
	researcherMinViews               uint64  = 100
	researcherMinCategories          uint64  = 5
	researcherMaxChatRate            float64 = 0.05
)

// DetectBehavior is deterministic: equal counters and rules version produce equal behavior.
// More specific scenarios stay above broader ones.
func DetectBehavior(metrics Metrics) Behavior {
	metrics = EnrichMetrics(metrics)
	drafts := uint64(0)
	if metrics.ListingsCreated > metrics.ListingsPublished {
		drafts = metrics.ListingsCreated - metrics.ListingsPublished
	}

	switch {
	case metrics.ListingsPublished >= activeSellerMinListings &&
		metrics.SalesCompleted >= activeSellerMinDeals:
		return Behavior{
			Code:        BehaviorActiveSeller,
			Title:       "Продажи в движении",
			Description: "Объявления регулярно публиковались, а сделки доходили до результата.",
			Reason: fmt.Sprintf(
				"Объявлений опубликовано: %d. Продаж завершено: %d.",
				metrics.ListingsPublished,
				metrics.SalesCompleted,
			),
		}

	case metrics.ListingsCreated >= startingSellerMinCreated &&
		drafts >= startingSellerMinDrafts &&
		metrics.PublicationRate < startingSellerMaxPublicationRate &&
		metrics.SalesCompleted == 0:
		return Behavior{
			Code:        BehaviorStartingSeller,
			Title:       "Старт в продажах",
			Description: "Создание объявлений уже началось, но часть публикаций ещё ждёт завершения.",
			Reason: fmt.Sprintf(
				"Объявлений создано: %d. Опубликовано: %d. Черновиков осталось: %d.",
				metrics.ListingsCreated,
				metrics.ListingsPublished,
				drafts,
			),
		}

	case metrics.PurchasesCompleted >= decisiveBuyerMinPurchases &&
		metrics.ChatsStarted >= decisiveBuyerMinChats &&
		metrics.PurchaseRate >= decisiveBuyerMinPurchaseRate:
		return Behavior{
			Code:        BehaviorDecisiveBuyer,
			Title:       "Быстрый выбор",
			Description: "После просмотра вариантов общение быстро переходило к покупке.",
			Reason: fmt.Sprintf(
				"Диалогов начато: %d. Покупок завершено: %d.",
				metrics.ChatsStarted,
				metrics.PurchasesCompleted,
			),
		}

	case metrics.TotalViews >= findHunterMinViews &&
		metrics.FavoriteRate >= findHunterMinFavorite &&
		metrics.RepeatRate >= findHunterMinRepeat:
		return Behavior{
			Code:        BehaviorFindHunter,
			Title:       "Охота за находками",
			Description: "Интересные варианты сохранялись, а к объявлениям часто возвращались для сравнения.",
			Reason: fmt.Sprintf(
				"В избранное добавлено: %d. Повторных просмотров: %d.",
				metrics.FavoritesAdded,
				metrics.RepeatedViews,
			),
		}

	case metrics.TotalViews >= researcherMinViews &&
		metrics.CategoriesCount >= researcherMinCategories &&
		metrics.ChatRate < researcherMaxChatRate:
		return Behavior{
			Code:        BehaviorResearcher,
			Title:       "Глубокое исследование",
			Description: "Много вариантов сравнивалось в разных категориях до перехода к общению.",
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
