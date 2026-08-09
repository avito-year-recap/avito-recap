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

// schemaStatements are executed in order on every startup. Every statement is
// idempotent (CREATE ... IF NOT EXISTS) so re-running them against an already
// provisioned database is a no-op. The database itself ("recap") is created by
// the ClickHouse container from CLICKHOUSE_DB, not here.
//
// Tables store the domain structs (model.Metrics/model.ActionableState/
// model.Recap) as JSON rather than as individually typed columns. All three
// already round-trip through encoding/json (model.Recap even carries custom
// Card.MarshalJSON/UnmarshalJSON for its closed-union payloads), and the
// application layer is the only writer, so there is no independent SQL-level
// consumer that would need typed columns. This keeps the schema stable while
// the domain model evolves.
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

	// events is the source of truth for a profile's yearly activity.
	// annual_metrics below is a cache of AggregateEvents(events) keyed by
	// (profile_id, year) — never written directly with a hand-computed
	// Metrics value. CalculateMetrics fills it lazily on first read and
	// revalidates it against events on every read after that (see
	// event_count on annual_metrics below) rather than trusting a cache hit
	// forever, so events landing after the first read are not silently lost.
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

	// event_count is the freshness marker: it records how many events for
	// (profile_id, year) existed in `events` when this row was computed. A
	// read compares it against the live count and recomputes on mismatch,
	// which is what actually invalidates the cache — the row's presence
	// alone is not treated as proof it is still correct.
	`CREATE TABLE IF NOT EXISTS annual_metrics
	(
		profile_id  UUID,
		year        UInt16,
		metrics     String,
		event_count UInt64,
		updated_at  DateTime DEFAULT now()
	)
	ENGINE = ReplacingMergeTree(updated_at)
	ORDER BY (profile_id, year)`,

	`CREATE TABLE IF NOT EXISTS actionable_state
	(
		profile_id UUID,
		state      String,
		updated_at DateTime DEFAULT now()
	)
	ENGINE = ReplacingMergeTree(updated_at)
	ORDER BY profile_id`,

	// Recaps are immutable once generated (see application.Service.Generate),
	// so this table is a plain MergeTree, not a ReplacingMergeTree: nothing
	// should ever overwrite a stored row, only CreateRecapIfAbsent's
	// check-then-insert semantics decide whether a new one is written.
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

// EnsureSchema provisions every table the application layer depends on.
func (r *Repo) EnsureSchema(ctx context.Context) error {
	for _, statement := range schemaStatements {
		if err := r.conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}
	return nil
}
