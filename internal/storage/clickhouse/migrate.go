package clickhouse

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const migrationsDir = "migrations"

const migrationsTableDDL = `
CREATE TABLE IF NOT EXISTS schema_migrations
(
    version    String,
    applied_at DateTime DEFAULT now()
)
ENGINE = MergeTree
ORDER BY version`

// EnsureSchema applies pending migrations/*.sql files in filename order and
// records each one in schema_migrations, so a file only ever runs once.
// Every migration file holds exactly one statement — the ClickHouse native
// protocol executes one statement per query, so splitting on ";" is not an
// option here.
func (r *Repo) EnsureSchema(ctx context.Context) error {
	if err := r.conn.Exec(ctx, migrationsTableDDL); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	applied, err := r.appliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationFiles, migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		name := entry.Name()
		if _, ok := applied[name]; ok {
			continue
		}

		content, err := migrationFiles.ReadFile(migrationsDir + "/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		statement := strings.TrimSuffix(strings.TrimSpace(string(content)), ";")
		if err := r.conn.Exec(ctx, statement); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if err := r.conn.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}

func (r *Repo) appliedMigrations(ctx context.Context) (map[string]struct{}, error) {
	rows, err := r.conn.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("close schema_migrations rows: %v", err)
		}
	}()

	applied := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan schema_migrations row: %w", err)
		}
		applied[version] = struct{}{}
	}
	return applied, rows.Err()
}
