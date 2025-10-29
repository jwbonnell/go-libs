package db

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jwbonnell/go-libs/pkg/db/queriers"
	"github.com/jwbonnell/go-libs/pkg/log"

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
func New(ctx context.Context, cfg ConnectionConfig, log *log.Logger) (*DB, error) {
	sslMode := "required"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")
	if cfg.Schema != "" {
		q.Set("search_path", cfg.Schema)
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	pgxCfg, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// optional sensible defaults
	pgxCfg.MaxConns = int32(cfg.MaxOpenConns)
	pgxCfg.MinConns = 1
	pgxCfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return &DB{
		pool: pool,
		log:  log,
	}, nil
}

func (d *DB) Status(ctx context.Context) error {
	// if a user supplied deadline is not supplied, default to a 1 second deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
	}

	for attempts := 1; ; attempts++ {
		if err := d.pool.Ping(ctx); err == nil {
			break
		}

		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	_, err := d.pool.Query(ctx, "SELECT VERSION()")
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
	}
}
