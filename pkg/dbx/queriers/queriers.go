package queriers

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jwbonnell/go-libs/pkg/logx"
)

// Querier is a minimal adapter interface that abstracts over a pgx connection
// pool and a transaction. Use it when you want the same code to run inside or
// outside a transaction.
//
// Query executes a SQL query that returns rows. Callers are responsible for
// closing the returned pgx.Rows (or use helpers like pgx.CollectRows that
// close rows for you).
//
// Exec executes a statement (INSERT/UPDATE/DELETE) and returns a CommandTag.
// Inspect CommandTag.RowsAffected() to see how many rows were modified.
//
// Begin starts a transaction and returns a *TxQuerier that wraps the started
// transaction. Implementations may return an error if a transaction cannot be
// started.
type Querier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (*TxQuerier, error)
}

// PoolQuerier is a thin wrapper around *pgxpool.Pool that implements Querier.
//
// Fields:
//   - Q: underlying connection pool (required).
//   - Log: optional logger that callers or higher-level helpers can use.
type PoolQuerier struct {
	Q   *pgxpool.Pool
	Log *logx.Logger
}

// Query forwards the call to the underlying pool's Query method.
// The caller must close the returned rows.
func (pq *PoolQuerier) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return pq.Q.Query(ctx, sql, args...)
}

// QueryRow is a convenience helper forwarding to pool.QueryRow for single-row queries.
func (pq *PoolQuerier) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return pq.Q.QueryRow(ctx, sql, args...)
}

// Exec forwards the call to the underlying pool's Exec method and returns the CommandTag.
func (pq *PoolQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pq.Q.Exec(ctx, sql, args...)
}

// Begin starts a transaction on the pool and returns a TxQuerier that wraps it.
func (pq *PoolQuerier) Begin(ctx context.Context) (*TxQuerier, error) {
	tx, err := pq.Q.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &TxQuerier{
		q:   tx,
		log: pq.Log,
	}, nil
}

// TxQuerier wraps a pgx.Tx and implements Querier for use inside transactions.
//
// The wrapper stores the pgx.Tx (interface) directly (not a pointer to an
// interface). The optional logx field is carried through from PoolQuerier when
// transactions are started.
type TxQuerier struct {
	q   pgx.Tx
	log *logx.Logger
}

// Query forwards to the underlying transaction's Query method.
// Callers must close the returned rows.
func (tq *TxQuerier) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return tq.q.Query(ctx, sql, args...)
}

// Exec forwards to the underlying transaction's Exec method and returns the CommandTag.
func (tq *TxQuerier) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return tq.q.Exec(ctx, sql, args...)
}

// Begin starts a nested transaction (savepoint) on the current transaction, if
// supported by pgx. It returns a new TxQuerier wrapping the nested transaction.
func (tq *TxQuerier) Begin(ctx context.Context) (*TxQuerier, error) {
	tx, err := tq.q.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &TxQuerier{
		q:   tx,
		log: tq.log,
	}, nil
}

// Commit commits the underlying transaction.
func (tq *TxQuerier) Commit(ctx context.Context) error {
	return tq.q.Commit(ctx)
}

// Rollback rolls back the underlying transaction.
func (tq *TxQuerier) Rollback(ctx context.Context) error {
	return tq.q.Rollback(ctx)
}
