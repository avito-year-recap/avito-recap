package clickhouse

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"

	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
)

func (r *Repo) GetRecapByKey(ctx context.Context, key model.RecapKey) (model.Recap, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT recap
		FROM recaps
		WHERE profile_id = ? AND year = ? AND rules_version = ? AND rules_digest = ?
		LIMIT 1
	`, key.ProfileID, uint16(key.Year), key.RulesVersion, key.RulesDigest)
	if err != nil {
		return model.Recap{}, fmt.Errorf("query recap by key: %w", err)
	}
	defer rows.Close()
	return scanOneRecap(rows, application.ErrRecapNotFound)
}

func (r *Repo) GetRecap(ctx context.Context, recapID uuid.UUID) (model.Recap, error) {
	rows, err := r.conn.Query(ctx, `SELECT recap FROM recaps WHERE id = ? LIMIT 1`, recapID)
	if err != nil {
		return model.Recap{}, fmt.Errorf("query recap: %w", err)
	}
	defer rows.Close()
	return scanOneRecap(rows, application.ErrRecapNotFound)
}

func (r *Repo) GetRecapByShareID(ctx context.Context, shareID uuid.UUID) (model.Recap, error) {
	rows, err := r.conn.Query(ctx, `SELECT recap FROM recaps WHERE share_id = ? LIMIT 1`, shareID)
	if err != nil {
		return model.Recap{}, fmt.Errorf("query recap by share id: %w", err)
	}
	defer rows.Close()
	return scanOneRecap(rows, application.ErrRecapNotFound)
}

// CreateRecapIfAbsent is check-then-insert, not a single atomic statement:
// ClickHouse's MergeTree family has no unique constraint or transactional
// upsert to enforce "one row per idempotency key" at write time the way a row
// store would. For a single-writer demo deployment the race window between
// the read and the insert is acceptable; a multi-writer deployment would need
// an external lock (or a dedicated OLTP store) in front of this key.
func (r *Repo) CreateRecapIfAbsent(ctx context.Context, key model.RecapKey, value model.Recap) (model.Recap, error) {
	if existing, err := r.GetRecapByKey(ctx, key); err == nil {
		return existing, nil
	} else if !errors.Is(err, application.ErrRecapNotFound) {
		return model.Recap{}, fmt.Errorf("check existing recap: %w", err)
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return model.Recap{}, fmt.Errorf("encode recap: %w", err)
	}
	if err := r.conn.Exec(ctx, `
		INSERT INTO recaps (id, share_id, profile_id, year, rules_version, rules_digest, recap)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		value.ID, value.ShareID, value.Profile.ID, uint16(value.Year),
		value.RulesVersion, value.RulesDigest, string(encoded),
	); err != nil {
		return model.Recap{}, fmt.Errorf("insert recap: %w", err)
	}
	return value, nil
}

func scanOneRecap(rows driver.Rows, notFound error) (model.Recap, error) {
	if !rows.Next() {
		return model.Recap{}, notFound
	}
	var raw string
	if err := rows.Scan(&raw); err != nil {
		return model.Recap{}, fmt.Errorf("scan recap: %w", err)
	}
	var value model.Recap
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return model.Recap{}, fmt.Errorf("decode recap: %w", err)
	}
	return value, rows.Err()
}
