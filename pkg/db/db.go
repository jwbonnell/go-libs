package db

import (
	"context"
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/db/queriers"
	"github.com/jwbonnell/go-libs/pkg/log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgxpool.Pool and provides simple helpers.
type DB struct {
	pool *pgxpool.Pool
	log  *log.Logger
}

func (d *DB) Pool() *queriers.PoolQuerier {
	return &queriers.PoolQuerier{
		Q:   d.pool,
		Log: d.log,
	}
}

// New creates a new DB pool. connString is a standard PG connection string.
func New(ctx context.Context, connString string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// optional sensible defaults
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{pool: pool}, nil
}

func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}
