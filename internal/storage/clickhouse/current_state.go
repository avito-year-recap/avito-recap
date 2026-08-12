package clickhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
)

func (r *Repo) GetActionableState(ctx context.Context, profileID uuid.UUID, asOf time.Time) (model.ActionableState, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT state
		FROM actionable_state FINAL
		WHERE profile_id = ?
		LIMIT 1
	`, profileID)
	if err != nil {
		return model.ActionableState{}, fmt.Errorf("query actionable state: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return model.ActionableState{}, application.ErrRecapNotFound
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return model.ActionableState{}, fmt.Errorf("scan actionable state: %w", err)
	}
	var state model.ActionableState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return model.ActionableState{}, fmt.Errorf("decode actionable state: %w", err)
	}
	// CapturedAt is when this snapshot was read, not when the underlying facts
	// (drafts/dialogs/favorites) were last written — matches the in-memory
	// storage adapter's contract.
	state.CapturedAt = asOf.UTC()
	return state, rows.Err()
}

// PutActionableState implements bootstrap.SeedStorage.
func (r *Repo) PutActionableState(ctx context.Context, profileID uuid.UUID, _ time.Time, state model.ActionableState) error {
	encoded, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode actionable state: %w", err)
	}
	if err := r.conn.Exec(ctx, `
		INSERT INTO actionable_state (profile_id, state) VALUES (?, ?)
	`, profileID, string(encoded)); err != nil {
		return fmt.Errorf("insert actionable state: %w", err)
	}
	return nil
}
