package clickhouse

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/year-recap/internal/recap"
)

// Сохраняем сырое собитие
func (r *Repo) InsertEvents(ctx context.Context, events []recap.Event) error {
	batch, err := r.conn.PrepareBatch(ctx, `
		INSERT INTO events
			(event_id, user_id, event_type, event_time, category, subcategory, city, price, ad_id, dialog_id, search_query, metadata)
	`)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, e := range events {
		id := uuid.New()
		if e.ID != "" {
			if parsed, err := uuid.Parse(e.ID); err == nil {
				id = parsed
			}
		}

		metadata := e.Metadata
		if metadata == nil {
			metadata = map[string]string{}
		}

		if err := batch.Append(
			id,
			uint64(e.UserID),
			string(e.Type),
			e.OccurredAt,
			e.Category,
			e.Subcategory,
			e.City,
			e.Price,
			e.AdID,
			e.DialogID,
			e.SearchQuery,
			metadata,
		); err != nil {
			return fmt.Errorf("append event: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}
	return nil
}

// Получаем id всех активных пользователей в этом году
func (r *Repo) GetActiveUserIDs(ctx context.Context, year int) ([]recap.UserID, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT DISTINCT user_id
		FROM recap_category_month_agg
		WHERE year = ?
	`, uint16(year))
	if err != nil {
		return nil, fmt.Errorf("query active user ids: %w", err)
	}
	defer rows.Close()

	var out []recap.UserID
	for rows.Next() {
		var uid uint64
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan active user id: %w", err)
		}
		out = append(out, recap.UserID(uid))
	}
	return out, rows.Err()
}
