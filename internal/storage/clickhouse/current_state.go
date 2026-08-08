// Текущее состояние объявлений пользователя
package clickhouse

import (
	"context"
	"fmt"

	"github.com/year-recap/internal/recap"
)

const adStatusEventsSQL = `'ad_created', 'ad_published', 'ad_sold'`

// GetActiveListingIDs возвращает id объявлений (объявление сейчас активно, не продано).
func (r *Repo) GetActiveListingIDs(ctx context.Context, userID recap.UserID) ([]uint64, error) {
	rows, err := r.conn.Query(ctx, fmt.Sprintf(`
		SELECT ad_id
		FROM (
			SELECT ad_id, argMax(event_type, event_time) AS latest_status
			FROM events
			WHERE user_id = ? AND ad_id IS NOT NULL AND event_type IN (%s)
			GROUP BY ad_id
		)
		WHERE latest_status = ?
	`, adStatusEventsSQL), uint64(userID), string(recap.EventAdPublished))
	if err != nil {
		return nil, fmt.Errorf("query active listing ids: %w", err)
	}
	defer rows.Close()

	var out []uint64
	for rows.Next() {
		var adID uint64
		if err := rows.Scan(&adID); err != nil {
			return nil, fmt.Errorf("scan active listing id: %w", err)
		}
		out = append(out, adID)
	}
	return out, rows.Err()
}

// GetDraftListingID возвращает id самого недавно черновика
func (r *Repo) GetDraftListingID(ctx context.Context, userID recap.UserID) (*uint64, error) {
	rows, err := r.conn.Query(ctx, fmt.Sprintf(`
		SELECT ad_id
		FROM (
			SELECT
				ad_id,
				argMax(event_type, event_time) AS latest_status,
				max(event_time) AS latest_event_time
			FROM events
			WHERE user_id = ? AND ad_id IS NOT NULL AND event_type IN (%s)
			GROUP BY ad_id
		)
		WHERE latest_status = ?
		ORDER BY latest_event_time DESC
		LIMIT 1
	`, adStatusEventsSQL), uint64(userID), string(recap.EventAdCreated))
	if err != nil {
		return nil, fmt.Errorf("query draft listing id: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, rows.Err()
	}

	var adID uint64
	if err := rows.Scan(&adID); err != nil {
		return nil, fmt.Errorf("scan draft listing id: %w", err)
	}
	return &adID, nil
}
