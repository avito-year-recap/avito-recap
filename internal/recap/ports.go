package recap

import (
	"context"
	"time"
)

// Итоговые данные по одному пользователю за год: количество действий.
// Используется как для карточки "итоговые данные".
type YearSummary struct {
	TotalEvents  uint64
	ActiveDays   uint64
	FirstEventAt time.Time
	LastEventAt  time.Time
	CountByType  map[EventType]uint64
}

// Количество действий для одной категории, всегда возвращается отсортированным
type CategoryStat struct {
	Category string
	Count    uint64
}

// Количество действий за один календарный месяц (1-12)
type MonthStat struct {
	Month int
	Count uint64
}

// Количество поисковых запросов, возвращаемое в порядке убывания по Count.
type SearchStat struct {
	Query string
	Count uint64
}

// Описывает, когда в течение дня/недели пользователь активен.
// Используется для карточки "ритм" и для правил, основанных на времени, таких как образ "ночной совы"
type ActivityRhythm struct {
	CountByHour    [24]uint64
	CountByWeekday [7]uint64
}

// Категория, к которой пользователь проявил сильный интерес,
// (повторно добавленные в избранное), но так и не реализовал ее
type MissedOpportunity struct {
	Category           string
	FavoritesCount     uint64
	FollowThroughCount uint64
}

type Repository interface {
    // Добавляем событие
	InsertEvents(ctx context.Context, events []Event) error

    // Сводка по году: идёт в карточку "Год в цифрах" и как вход для правил персоны/ачивок.
	GetYearSummary(ctx context.Context, userID UserID, year int) (YearSummary, error)
	// Разбивка активности по категориям, отсортированная по убыванию
	GetCategoryBreakdown(ctx context.Context, userID UserID, year int) ([]CategoryStat, error)
	// Активность по месяцам (1-12) — используется, чтобы найти пиковый месяц ("активнее всего вы были в марте").
	GetMonthlyActivity(ctx context.Context, userID UserID, year int) ([]MonthStat, error)
	// Топ поисковых запросов пользователя за год — для карточки "что вы искали"
	GetTopSearches(ctx context.Context, userID UserID, year int, limit int) ([]SearchStat, error)
	// Распределение активности по часам суток и дням недели
	GetActivityRhythm(ctx context.Context, userID UserID, year int) (ActivityRhythm, error)
	// Категории, где пользователь активно добавлял в избранное (≥3 раза), но
    // так и не написал продавцу, не позвонил и не купил
	GetMissedOpportunities(ctx context.Context, userID UserID, year int) ([]MissedOpportunity, error)

	// Число действий ровно по одной заданной категории.
	GetCategoryActionCount(ctx context.Context, userID UserID, year int, category string) (uint64, error)

	// Число действий ровно одного заданного типа.
	GetEventTypeActionCount(ctx context.Context, userID UserID, year int, eventType EventType) (uint64, error)

	// GetViewRepeatStats — сколько всего было просмотров и сколько из них уникальных объявлений за год
	GetViewRepeatStats(ctx context.Context, userID UserID, year int) (totalViews uint64, uniqueAdsViewed uint64, err error)

	// GetChatsStartedCount — число уникальных диалогов за год
	GetChatsStartedCount(ctx context.Context, userID UserID, year int) (uint64, error)

	// GetChatsWithPurchaseCount — число уникальных диалогов за год, которые привели к покупке
	GetChatsWithPurchaseCount(ctx context.Context, userID UserID, year int) (uint64, error)

	// Последнее активное объявление
	GetActiveListingIDs(ctx context.Context, userID UserID) ([]uint64, error)

	// GetDraftListingID — id самого недавно черновика
	GetDraftListingID(ctx context.Context, userID UserID) (*uint64, error)

	// Список активных пользователей (по событиям за год)
	GetActiveUserIDs(ctx context.Context, year int) ([]UserID, error)

    // Сохранение рассчитаной карточки
	SaveRecap(ctx context.Context, r Recap) error

	// Получение рассчитаной карточки
	GetCachedRecap(ctx context.Context, userID UserID, year int) (Recap, bool, error)
}
