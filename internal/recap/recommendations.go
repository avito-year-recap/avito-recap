package recap

import "fmt"

const (
	favoritesForAction  uint64 = 10
	publishedForImprove uint64 = 3
	chatsForContinue    uint64 = 3
)

// BuildNextAction returns exactly one deterministic next step.
// Rules are ordered by the expected value of continuing an unfinished scenario.
func BuildNextAction(metrics Metrics) NextAction {
	switch {
	case metrics.ListingsCreated > metrics.ListingsPublished:
		drafts := metrics.ListingsCreated - metrics.ListingsPublished

		return NextAction{
			Code:        ActionFinishDraft,
			Title:       "Заверши начатое объявление",
			Description: fmt.Sprintf("У тебя осталось незавершённых черновиков: %d.", drafts),
			ButtonText:  "Продолжить публикацию",
			Reason:      "Количество созданных объявлений больше количества опубликованных.",
		}

	case metrics.ListingsPublished >= publishedForImprove && metrics.SalesCompleted == 0:
		return NextAction{
			Code:        ActionImproveListings,
			Title:       "Усиль свои объявления",
			Description: "Попробуй обновить фотографии или описание, чтобы получить больше откликов.",
			ButtonText:  "Посмотреть объявления",
			Reason:      "Опубликовано несколько объявлений, но завершённых сделок пока нет.",
		}

	case metrics.ChatsStarted >= chatsForContinue &&
		metrics.PurchasesCompleted == 0 && metrics.SalesCompleted == 0:
		return NextAction{
			Code:        ActionContinueChats,
			Title:       "Продолжи начатые диалоги",
			Description: "В переписках могут оставаться сценарии, которые ещё можно завершить.",
			ButtonText:  "Открыть сообщения",
			Reason:      "Есть начатые диалоги, но нет завершённых сделок.",
		}

	case metrics.FavoritesAdded >= favoritesForAction && metrics.ChatsStarted == 0:
		return NextAction{
			Code:        ActionOpenFavorites,
			Title:       "Вернись к своим находкам",
			Description: "В избранном уже есть варианты, с которыми можно продолжить.",
			ButtonText:  "Открыть избранное",
			Reason:      "Есть много сохранённых объявлений, но ещё не начато ни одного диалога.",
		}

	case metrics.TopCategory != "":
		return NextAction{
			Code:        ActionOpenTopCategory,
			Title:       "Посмотри новые предложения",
			Description: fmt.Sprintf("Вернись в категорию «%s» и проверь новые варианты.", metrics.TopCategory),
			ButtonText:  "Открыть категорию",
			Reason:      "Эта категория была самой просматриваемой за год.",
		}

	default:
		return NextAction{
			Code:        ActionCreateFirstListing,
			Title:       "Попробуй новый сценарий",
			Description: "Создай первое объявление и узнай, насколько быстро найдётся покупатель.",
			ButtonText:  "Создать объявление",
			Reason:      "Не найден более приоритетный незавершённый пользовательский сценарий.",
		}
	}
}
