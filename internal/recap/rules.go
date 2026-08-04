package recap

import "fmt"

const (
	activeSellerMinListings uint64  = 5
	activeSellerMinDeals    uint64  = 3
	findHunterMinViews      uint64  = 20
	findHunterMinFavorite   float64 = 0.15
	findHunterMinRepeat     float64 = 0.20
	researcherMinViews      uint64  = 100
	researcherMinCategories uint64  = 5
	researcherMaxChatRate   float64 = 0.05
)

// DetectBehavior is deterministic: equal metrics and rules version produce equal behavior.
// More specific behavior rules must stay above broader rules.
func DetectBehavior(metrics Metrics) Behavior {
	switch {
	case metrics.ListingsPublished >= activeSellerMinListings &&
		metrics.SalesCompleted >= activeSellerMinDeals:
		return Behavior{
			Code:        BehaviorActiveSeller,
			Title:       "Активный продавец",
			Description: "Ты регулярно публиковал объявления и доводил сделки до результата.",
			Reason: fmt.Sprintf(
				"Опубликовано %d объявлений и завершено %d сделок.",
				metrics.ListingsPublished,
				metrics.SalesCompleted,
			),
		}

	case metrics.TotalViews >= findHunterMinViews &&
		metrics.FavoriteRate >= findHunterMinFavorite &&
		metrics.RepeatRate >= findHunterMinRepeat:
		return Behavior{
			Code:        BehaviorFindHunter,
			Title:       "Охотник за находками",
			Description: "Ты сохранял интересные варианты и возвращался к ним, чтобы сравнить.",
			Reason: fmt.Sprintf(
				"В избранное добавлено %d объявлений, повторных просмотров — %d.",
				metrics.FavoritesAdded,
				metrics.RepeatedViews,
			),
		}

	case metrics.TotalViews >= researcherMinViews &&
		metrics.CategoriesCount >= researcherMinCategories &&
		metrics.ChatRate < researcherMaxChatRate:
		return Behavior{
			Code:        BehaviorResearcher,
			Title:       "Исследователь",
			Description: "Ты внимательно изучал рынок и сравнивал варианты в разных категориях.",
			Reason: fmt.Sprintf(
				"Просмотрено %d объявлений в %d категориях, начато %d диалогов.",
				metrics.TotalViews,
				metrics.CategoriesCount,
				metrics.ChatsStarted,
			),
		}

	default:
		return Behavior{
			Code:        BehaviorUniversal,
			Title:       "Универсальный пользователь",
			Description: "Ты использовал разные возможности площадки и не ограничивался одним сценарием.",
			Reason:      "Активность не попала под один доминирующий сценарий.",
		}
	}
}
