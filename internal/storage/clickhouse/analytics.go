package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
)

func (r *Repo) CalculateMetrics(ctx context.Context, profileID uuid.UUID, period model.RecapPeriod) (model.Metrics, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT metrics
		FROM annual_metrics FINAL
		WHERE profile_id = ? AND year = ?
		LIMIT 1
	`, profileID, uint16(period.Year))
	if err != nil {
		return model.Metrics{}, fmt.Errorf("query annual metrics: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return model.Metrics{}, application.ErrMetricsNotFound
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return model.Metrics{}, fmt.Errorf("scan annual metrics: %w", err)
	}
	var metrics model.Metrics
	if err := json.Unmarshal([]byte(raw), &metrics); err != nil {
		return model.Metrics{}, fmt.Errorf("decode annual metrics: %w", err)
	}
	return metrics, rows.Err()
}

// UpsertAnnualMetrics implements bootstrap.SeedStorage.
func (r *Repo) UpsertAnnualMetrics(ctx context.Context, profileID uuid.UUID, year uint32, metrics model.Metrics) error {
	encoded, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("encode annual metrics: %w", err)
	}
	if err := r.conn.Exec(ctx, `
		INSERT INTO annual_metrics (profile_id, year, metrics) VALUES (?, ?, ?)
	`, profileID, uint16(year), string(encoded)); err != nil {
		return fmt.Errorf("insert annual metrics: %w", err)
	}
	return nil
}
