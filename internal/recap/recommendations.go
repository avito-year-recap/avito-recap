package recap

import "fmt"

const (
	publishedForImprove uint64 = 3
	chatsForContinue    uint64 = 3
)

// BuildNextAction returns one deterministic next step tied to the dominant behavior.
// Continuing an unfinished scenario is preferred over a generic discovery action.
func BuildNextAction(metrics Metrics) NextAction {
	metrics = EnrichMetrics(metrics)
	behavior := DetectBehavior(metrics)

	switch behavior.Code {
	case BehaviorStartingSeller:
		return NextAction{
			Code:        ActionFinishDraft,
			Title:       "Проверь начатые объявления",
			Description: "Открой созданные объявления и заверши те, которые ещё не опубликованы.",
			ButtonText:  "Проверить объявления",
			Reason: fmt.Sprintf(
				"В выбранном году создано %d объявлений, опубликовано %d; это повод проверить незавершённые публикации, но не точное число черновиков.",
				metrics.ListingsCreated,
				metrics.ListingsPublished,
			),
		}

	case BehaviorActiveSeller:
		return NextAction{
			Code:        ActionCreateListing,
			Title:       "Продолжи успешный сценарий",
			Description: "Создай новое объявление, пока опыт публикаций и продаж остаётся актуальным.",
			ButtonText:  "Создать объявление",
			Reason:      "В течение года было много публикаций и несколько завершённых продаж.",
		}

	case BehaviorDecisiveBuyer:
		return NextAction{
			Code:        ActionViewSimilarListings,
			Title:       "Посмотри похожие варианты",
			Description: "Подборка по главному интересу поможет быстро перейти к следующему выбору.",
			ButtonText:  "Смотреть похожее",
			Reason:      "Несколько диалогов в выбранном периоде были связаны с завершёнными покупками.",
		}

	case BehaviorResearcher:
		return NextAction{
			Code:        ActionSaveSearch,
			Title:       "Сохрани поиск",
			Description: "Новые объявления по выбранным параметрам будут легче отслеживать без повторного поиска.",
			ButtonText:  "Сохранить поиск",
			Reason:      "Просмотров и категорий много, а начатых диалогов мало — автоматическое обновление поиска сокращает повторную работу.",
		}

	case BehaviorFindHunter:
		return NextAction{
			Code:        ActionOpenFavorites,
			Title:       "Вернись к своим находкам",
			Description: "В избранном уже есть варианты, которые можно ещё раз сравнить и обсудить.",
			ButtonText:  "Открыть избранное",
			Reason:      "В течение года было несколько добавлений в избранное и много повторных просмотров.",
		}
	}

	switch {
	case metrics.ChatsStarted >= chatsForContinue:
		return NextAction{
			Code:        ActionContinueDialogs,
			Title:       "Продолжи начатые диалоги",
			Description: "В сообщениях могут оставаться вопросы и договорённости, которые стоит завершить.",
			ButtonText:  "Открыть сообщения",
			Reason:      "Диалоги были заметной частью активности, но ни один более узкий сценарий не доминирует.",
		}

	case metrics.ListingsPublished >= publishedForImprove && metrics.SalesCompleted == 0:
		return NextAction{
			Code:        ActionImproveListings,
			Title:       "Усиль свои объявления",
			Description: "Обнови фотографии или описание, чтобы повысить шанс на отклик.",
			ButtonText:  "Посмотреть объявления",
			Reason:      "Опубликовано несколько объявлений, но завершённых продаж пока нет.",
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
			Reason:      "Не найден более приоритетный незавершённый сценарий.",
		}
	}
}
