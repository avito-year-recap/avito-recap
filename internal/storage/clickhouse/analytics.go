// Запросы для агрегации данных
package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/year-recap/internal/recap"
)

// Сводка по году: идёт в карточку "Год в цифрах" и как вход для правил персоны/ачивок.
func (r *Repo) GetYearSummary(ctx context.Context, userID recap.UserID, year int) (recap.YearSummary, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	var totals struct {
		ActiveDays   uint64    `ch:"active_days"`
		FirstEventAt time.Time `ch:"first_event_at"`
		LastEventAt  time.Time `ch:"last_event_at"`
		TotalEvents  uint64    `ch:"total_events"`
	}
	err := r.conn.QueryRow(ctx, `
		SELECT
			uniqExact(toDate(event_time)) AS active_days,
			min(event_time) AS first_event_at,
			max(event_time) AS last_event_at,
			count() AS total_events
		FROM events
		WHERE user_id = ? AND event_time >= ? AND event_time < ?
	`, uint64(userID), start, end).ScanStruct(&totals)
	if err != nil {
		return recap.YearSummary{}, fmt.Errorf("query year totals: %w", err)
	}

	rows, err := r.conn.Query(ctx, `
		SELECT event_type, countMerge(cnt) AS total
		FROM recap_category_month_agg
		WHERE user_id = ? AND year = ?
		GROUP BY event_type
	`, uint64(userID), uint16(year))
	if err != nil {
		return recap.YearSummary{}, fmt.Errorf("query event type totals: %w", err)
	}
	defer rows.Close()

	countByType := make(map[recap.EventType]uint64)
	for rows.Next() {
		var (
			eventType string
			total     uint64
		)
		if err := rows.Scan(&eventType, &total); err != nil {
			return recap.YearSummary{}, fmt.Errorf("scan event type total: %w", err)
		}
		countByType[recap.EventType(eventType)] = total
	}
	if err := rows.Err(); err != nil {
		return recap.YearSummary{}, fmt.Errorf("iterate event type totals: %w", err)
	}

	return recap.YearSummary{
		TotalEvents:  totals.TotalEvents,
		ActiveDays:   totals.ActiveDays,
		FirstEventAt: totals.FirstEventAt,
		LastEventAt:  totals.LastEventAt,
		CountByType:  countByType,
	}, nil
}

// Разбивка активности по категориям, отсортированная по убыванию —
// Идёт в карточку "Ваша тема года".
func (r *Repo) GetCategoryBreakdown(ctx context.Context, userID recap.UserID, year int) ([]recap.CategoryStat, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT category, countMerge(cnt) AS total
		FROM recap_category_month_agg
		WHERE user_id = ? AND year = ?
		GROUP BY category
		ORDER BY total DESC
	`, uint64(userID), uint16(year))
	if err != nil {
		return nil, fmt.Errorf("query category breakdown: %w", err)
	}
	defer rows.Close()

	var out []recap.CategoryStat
	for rows.Next() {
		var c recap.CategoryStat
		if err := rows.Scan(&c.Category, &c.Count); err != nil {
			return nil, fmt.Errorf("scan category stat: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Активность по месяцам (1-12) — используется, чтобы найти пиковый месяц ("активнее всего вы были в марте").
func (r *Repo) GetMonthlyActivity(ctx context.Context, userID recap.UserID, year int) ([]recap.MonthStat, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT month, countMerge(cnt) AS total
		FROM recap_category_month_agg
		WHERE user_id = ? AND year = ?
		GROUP BY month
		ORDER BY month
	`, uint64(userID), uint16(year))
	if err != nil {
		return nil, fmt.Errorf("query monthly activity: %w", err)
	}
	defer rows.Close()

	var out []recap.MonthStat
	for rows.Next() {
		var (
			month uint8
			total uint64
		)
		if err := rows.Scan(&month, &total); err != nil {
			return nil, fmt.Errorf("scan month stat: %w", err)
		}
		out = append(out, recap.MonthStat{Month: int(month), Count: total})
	}
	return out, rows.Err()
}

// Топ поисковых запросов пользователя за год — для карточки "что вы искали"
func (r *Repo) GetTopSearches(ctx context.Context, userID recap.UserID, year int, limit int) ([]recap.SearchStat, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	rows, err := r.conn.Query(ctx, `
		SELECT search_query, count() AS total
		FROM events
		WHERE user_id = ?
		  AND event_type = ?
		  AND event_time >= ? AND event_time < ?
		  AND search_query IS NOT NULL
		GROUP BY search_query
		ORDER BY total DESC
		LIMIT ?
	`, uint64(userID), string(recap.EventSearch), start, end, limit)
	if err != nil {
		return nil, fmt.Errorf("query top searches: %w", err)
	}
	defer rows.Close()

	var out []recap.SearchStat
	for rows.Next() {
		var s recap.SearchStat
		if err := rows.Scan(&s.Query, &s.Count); err != nil {
			return nil, fmt.Errorf("scan search stat: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// Распределение активности по часам суток и дням недели
func (r *Repo) GetActivityRhythm(ctx context.Context, userID recap.UserID, year int) (recap.ActivityRhythm, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	var rhythm recap.ActivityRhythm

	hourRows, err := r.conn.Query(ctx, `
		SELECT toHour(event_time) AS hour, count() AS total
		FROM events
		WHERE user_id = ? AND event_time >= ? AND event_time < ?
		GROUP BY hour
	`, uint64(userID), start, end)
	if err != nil {
		return rhythm, fmt.Errorf("query hourly rhythm: %w", err)
	}
	for hourRows.Next() {
		var hour uint8
		var total uint64
		if err := hourRows.Scan(&hour, &total); err != nil {
			hourRows.Close()
			return rhythm, fmt.Errorf("scan hourly rhythm: %w", err)
		}
		if hour < 24 {
			rhythm.CountByHour[hour] = total
		}
	}
	if err := hourRows.Err(); err != nil {
		hourRows.Close()
		return rhythm, fmt.Errorf("iterate hourly rhythm: %w", err)
	}
	hourRows.Close()

	dayRows, err := r.conn.Query(ctx, `
		SELECT toDayOfWeek(event_time) AS weekday, count() AS total
		FROM events
		WHERE user_id = ? AND event_time >= ? AND event_time < ?
		GROUP BY weekday
	`, uint64(userID), start, end)
	if err != nil {
		return rhythm, fmt.Errorf("query weekday rhythm: %w", err)
	}
	for dayRows.Next() {
		var weekday uint8
		var total uint64
		if err := dayRows.Scan(&weekday, &total); err != nil {
			dayRows.Close()
			return rhythm, fmt.Errorf("scan weekday rhythm: %w", err)
		}
		if weekday >= 1 && weekday <= 7 {
			rhythm.CountByWeekday[weekday-1] = total
		}
	}
	if err := dayRows.Err(); err != nil {
		dayRows.Close()
		return rhythm, fmt.Errorf("iterate weekday rhythm: %w", err)
	}
	dayRows.Close()

	return rhythm, nil
}

// Категории, где пользователь активно добавлял в избранное (≥3 раза), но
// так и не написал продавцу, не позвонил и не купил
func (r *Repo) GetMissedOpportunities(ctx context.Context, userID recap.UserID, year int) ([]recap.MissedOpportunity, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			category,
			sumIf(total, event_type = 'favorite_add') AS favorites,
			sumIf(total, event_type IN ('message_sent', 'call_made', 'purchase_made')) AS follow_through
		FROM (
			SELECT category, event_type, countMerge(cnt) AS total
			FROM recap_category_month_agg
			WHERE user_id = ? AND year = ?
			GROUP BY category, event_type
		)
		GROUP BY category
		HAVING favorites >= 3 AND follow_through = 0
		ORDER BY favorites DESC
		LIMIT 5
	`, uint64(userID), uint16(year))
	if err != nil {
		return nil, fmt.Errorf("query missed opportunities: %w", err)
	}
	defer rows.Close()

	var out []recap.MissedOpportunity
	for rows.Next() {
		var m recap.MissedOpportunity
		if err := rows.Scan(&m.Category, &m.FavoritesCount, &m.FollowThroughCount); err != nil {
			return nil, fmt.Errorf("scan missed opportunity: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Получаем количество действий по одной конкретной категории
func (r *Repo) GetCategoryActionCount(ctx context.Context, userID recap.UserID, year int, category string) (uint64, error) {
	var total uint64
	err := r.conn.QueryRow(ctx, `
		SELECT countMerge(cnt) AS total
		FROM recap_category_month_agg
		WHERE user_id = ? AND year = ? AND category = ?
	`, uint64(userID), uint16(year), category).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query category action count: %w", err)
	}
	return total, nil
}

// Получаем количество действий одного конкретного типа
func (r *Repo) GetEventTypeActionCount(ctx context.Context, userID recap.UserID, year int, eventType recap.EventType) (uint64, error) {
	var total uint64
	err := r.conn.QueryRow(ctx, `
		SELECT countMerge(cnt) AS total
		FROM recap_category_month_agg
		WHERE user_id = ? AND year = ? AND event_type = ?
	`, uint64(userID), uint16(year), string(eventType)).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query event type action count: %w", err)
	}
	return total, nil
}

// Считаем просмотры и уникальные объявления среди них за год
func (r *Repo) GetViewRepeatStats(ctx context.Context, userID recap.UserID, year int) (uint64, uint64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	var stats struct {
		TotalViews      uint64 `ch:"total_views"`
		UniqueAdsViewed uint64 `ch:"unique_ads_viewed"`
	}
	err := r.conn.QueryRow(ctx, `
		SELECT
			count() AS total_views,
			uniqExact(ad_id) AS unique_ads_viewed
		FROM events
		WHERE user_id = ?
		  AND event_type = ?
		  AND ad_id IS NOT NULL
		  AND event_time >= ? AND event_time < ?
	`, uint64(userID), string(recap.EventView), start, end).ScanStruct(&stats)
	if err != nil {
		return 0, 0, fmt.Errorf("query view repeat stats: %w", err)
	}
	return stats.TotalViews, stats.UniqueAdsViewed, nil
}

// Считаем уникальные диалоги (не сообщения!) за год
func (r *Repo) GetChatsStartedCount(ctx context.Context, userID recap.UserID, year int) (uint64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	var total uint64
	err := r.conn.QueryRow(ctx, `
		SELECT uniqExact(dialog_id) AS total
		FROM events
		WHERE user_id = ?
		  AND event_type = ?
		  AND dialog_id IS NOT NULL
		  AND event_time >= ? AND event_time < ?
	`, uint64(userID), string(recap.EventMessageSent), start, end).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query chats started count: %w", err)
	}
	return total, nil
}

// Считаем диалоги, приведшие к покупке за год — уникальные dialog_id на событиях purchase_made.
func (r *Repo) GetChatsWithPurchaseCount(ctx context.Context, userID recap.UserID, year int) (uint64, error) {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)

	var total uint64
	err := r.conn.QueryRow(ctx, `
		SELECT uniqExact(dialog_id) AS total
		FROM events
		WHERE user_id = ?
		  AND event_type = ?
		  AND dialog_id IS NOT NULL
		  AND event_time >= ? AND event_time < ?
	`, uint64(userID), string(recap.EventPurchaseMade), start, end).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query chats with purchase count: %w", err)
	}
	return total, nil
}
