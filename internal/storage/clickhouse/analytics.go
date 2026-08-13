package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"

	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
)

// Считает метрики. Если есть и актуально (считает количество событий) - берет из кеша, иначе идет в таблицу с сырыми данными
func (r *Repo) CalculateMetrics(ctx context.Context, profileID uuid.UUID, period model.RecapPeriod) (model.Metrics, error) {
	liveCount, err := r.countEvents(ctx, profileID, period.Year)
	if err != nil {
		return model.Metrics{}, fmt.Errorf("count events: %w", err)
	}

	// Проверяем кеш до проверки liveCount==0 - UpsertAnnualMetrics может
	// зафиксировать метрики для профиля без сырых events (например, при
	// миграции исторических данных), и такой кеш должен читаться как обычно.
	if cached, ok, err := r.cachedMetrics(ctx, profileID, period.Year); err != nil {
		return model.Metrics{}, err
	} else if ok && cached.eventCount == liveCount {
		return cached.metrics, nil
	}

	if liveCount == 0 {
		return model.Metrics{}, application.ErrMetricsNotFound
	}

	events, err := r.queryEvents(ctx, profileID, period.Year)
	if err != nil {
		return model.Metrics{}, fmt.Errorf("query events: %w", err)
	}
	metrics := analytics.AggregateEvents(events)

	if err := r.writeMetricsCache(ctx, profileID, period.Year, metrics, uint64(len(events))); err != nil {
		return model.Metrics{}, fmt.Errorf("cache aggregated metrics: %w", err)
	}
	return metrics, nil
}

func (r *Repo) countEvents(ctx context.Context, profileID uuid.UUID, year uint32) (uint64, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT count()
		FROM events
		WHERE profile_id = ? AND toYear(occurred_at) = ?
	`, profileID, uint16(year))
	if err != nil {
		return 0, fmt.Errorf("query event count: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close event count rows: %v", err)
		}
	}()

	if !rows.Next() {
		return 0, rows.Err()
	}
	var count uint64
	if err := rows.Scan(&count); err != nil {
		return 0, fmt.Errorf("scan event count: %w", err)
	}
	return count, rows.Err()
}

type cachedMetricsRow struct {
	metrics    model.Metrics
	eventCount uint64
}

func (r *Repo) cachedMetrics(ctx context.Context, profileID uuid.UUID, year uint32) (cachedMetricsRow, bool, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT metrics, event_count
		FROM annual_metrics FINAL
		WHERE profile_id = ? AND year = ?
		LIMIT 1
	`, profileID, uint16(year))
	if err != nil {
		return cachedMetricsRow{}, false, fmt.Errorf("query annual metrics cache: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close annual metrics cache rows: %v", err)
		}
	}()

	if !rows.Next() {
		return cachedMetricsRow{}, false, rows.Err()
	}
	var raw string
	var eventCount uint64
	if err := rows.Scan(&raw, &eventCount); err != nil {
		return cachedMetricsRow{}, false, fmt.Errorf("scan annual metrics cache: %w", err)
	}
	var metrics model.Metrics
	if err := json.Unmarshal([]byte(raw), &metrics); err != nil {
		return cachedMetricsRow{}, false, fmt.Errorf("decode annual metrics cache: %w", err)
	}
	return cachedMetricsRow{metrics: metrics, eventCount: eventCount}, true, nil
}

// UpsertAnnualMetrics явно фиксирует метрики профиля за год, минуя агрегацию
// сырых events. Freshness-маркер берется как текущий liveCount, поэтому
// CalculateMetrics отдаст именно эти метрики, пока набор events профиля за
// год не изменится.
func (r *Repo) UpsertAnnualMetrics(ctx context.Context, profileID uuid.UUID, year uint32, metrics model.Metrics) error {
	liveCount, err := r.countEvents(ctx, profileID, year)
	if err != nil {
		return fmt.Errorf("count events: %w", err)
	}
	return r.writeMetricsCache(ctx, profileID, year, metrics, liveCount)
}

// Записываем кэш, если новый
func (r *Repo) writeMetricsCache(ctx context.Context, profileID uuid.UUID, year uint32, metrics model.Metrics, eventCount uint64) error {
	encoded, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("encode annual metrics: %w", err)
	}
	return r.conn.Exec(ctx, `
		INSERT INTO annual_metrics (profile_id, year, metrics, event_count) VALUES (?, ?, ?, ?)
	`, profileID, uint16(year), string(encoded), eventCount)
}

// Считываем метрики, если кэш пустой
func (r *Repo) queryEvents(ctx context.Context, profileID uuid.UUID, year uint32) ([]model.Event, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT id, profile_id, event_type, occurred_at, category, ad_id, dialog_id
		FROM events
		WHERE profile_id = ? AND toYear(occurred_at) = ?
	`, profileID, uint16(year))
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close events rows: %v", err)
		}
	}()

	var events []model.Event
	for rows.Next() {
		var event model.Event
		var eventType string
		if err := rows.Scan(
			&event.ID, &event.ProfileID, &eventType, &event.OccurredAt,
			&event.Category, &event.AdID, &event.DialogID,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		event.Type = model.ActivityType(eventType)
		events = append(events, event)
	}
	return events, rows.Err()
}

// Заносим готовые евенты (для авто-заполнения)
func (r *Repo) InsertEvents(ctx context.Context, events []model.Event) error {
	if len(events) == 0 {
		return nil
	}
	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO events (id, profile_id, event_type, occurred_at, category, ad_id, dialog_id)
	`)
	if err != nil {
		return fmt.Errorf("prepare events batch: %w", err)
	}
	for _, event := range events {
		if err := batch.Append(
			event.ID, event.ProfileID, string(event.Type), event.OccurredAt,
			event.Category, event.AdID, event.DialogID,
		); err != nil {
			return fmt.Errorf("append event: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("send events batch: %w", err)
	}
	return nil
}
