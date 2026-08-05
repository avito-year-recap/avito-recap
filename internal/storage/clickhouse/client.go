package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/year-recap/internal/recap"
)

type Repo struct {
	conn driver.Conn
}

var _ recap.Repository = (*Repo)(nil)

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
