package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/year-recap/internal/bootstrap"
	"github.com/year-recap/internal/recap/application"
)

type Repo struct {
	conn driver.Conn
}

var (
	_ application.ProfileStorage     = (*Repo)(nil)
	_ application.AnalyticsStorage   = (*Repo)(nil)
	_ application.ActionStateStorage = (*Repo)(nil)
	_ application.RecapStorage       = (*Repo)(nil)
	_ bootstrap.SeedStorage          = (*Repo)(nil)
)

func Connect(ctx context.Context, dsn string) (*Repo, error) {
	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &Repo{conn: conn}, nil
}

func (r *Repo) Close() error {
	return r.conn.Close()
}

// Создает таблицы под профили и евенты
var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS profiles
	(
		id           UUID,
		code         String,
		display_name String,
		description  String,
		avatar_url   String,
		updated_at   DateTime DEFAULT now()
	)
	ENGINE = ReplacingMergeTree(updated_at)
	ORDER BY id`,

	`CREATE TABLE IF NOT EXISTS events
	(
		id          UUID,
		profile_id  UUID,
		event_type  LowCardinality(String),
		occurred_at DateTime,
		category    LowCardinality(String),
		ad_id       Nullable(UInt64),
		dialog_id   Nullable(UInt64)
	)
	ENGINE = MergeTree
	PARTITION BY toYYYYMM(occurred_at)
	ORDER BY (profile_id, occurred_at)
	TTL occurred_at + INTERVAL 3 YEAR`,

	`CREATE TABLE IF NOT EXISTS annual_metrics
	(
		profile_id  UUID,
		year        UInt16,
		metrics     String,
		event_count UInt64,
		updated_at  DateTime DEFAULT now()
	)
	ENGINE = ReplacingMergeTree(updated_at)
	ORDER BY (profile_id, year)
	TTL updated_at + INTERVAL 3 YEAR`,

	// Текущее состояние профиля
	`CREATE TABLE IF NOT EXISTS actionable_state
	(
		profile_id UUID,
		state      String,
		updated_at DateTime DEFAULT now()
	)
	ENGINE = ReplacingMergeTree(updated_at)
	ORDER BY profile_id`,

	// Неизменяемые карточки
	`CREATE TABLE IF NOT EXISTS recaps
	(
		id            UUID,
		share_id      UUID,
		profile_id    UUID,
		year          UInt16,
		rules_version String,
		rules_digest  String,
		recap         String,
		created_at    DateTime DEFAULT now()
	)
	ENGINE = MergeTree
	ORDER BY (profile_id, year, rules_version, rules_digest)`,
}

var alterStatements = []string{
	`ALTER TABLE annual_metrics MODIFY TTL updated_at + INTERVAL 3 YEAR`,
}

// Проверка, что создали все необходимые таблицы
func (r *Repo) EnsureSchema(ctx context.Context) error {
	for _, statement := range schemaStatements {
		if err := r.conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}
	for _, statement := range alterStatements {
		if err := r.conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("apply schema migration: %w", err)
		}
	}
	return nil
}
