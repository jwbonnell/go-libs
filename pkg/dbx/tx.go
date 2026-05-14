package dbx

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/jwbonnell/go-libs/pkg/dbx/queriers"
)

// RunInTx executes fn inside a transaction. If the context already carries a
// transaction (via SetTx), fn runs within that existing transaction and the
// outermost RunInTx manages the lifecycle. If no transaction exists, a new one
// is started, committed on nil return, and rolled back on error.
func RunInTx(ctx context.Context, db queriers.Querier, fn func(txCtx context.Context) error) error {
	if _, ok := GetTx(ctx); ok {
		return fn(ctx)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("dbx.RunInTx begin: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(SetTx(ctx, tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// AdvisoryLockRunInTx is like RunInTx but also acquires a PostgreSQL
// transaction-scoped advisory lock (pg_advisory_xact_lock) before calling fn.
// The lock is automatically released when the transaction commits or rolls back.
func AdvisoryLockRunInTx(ctx context.Context, db queriers.Querier, lockID string, fn func(txCtx context.Context) error) error {
	return RunInTx(ctx, db, func(txCtx context.Context) error {
		tx, _ := GetTx(txCtx)

		h := fnv.New64a()
		if _, err := h.Write([]byte(lockID)); err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				return fmt.Errorf("dbx.AdvisoryLockRunInTx: failed to rollback transaction: txErr=%w, err=%w", txErr, err)
			}
			return fmt.Errorf("dbx.AdvisoryLockTxBegin: %w", err)
		}

		if _, err := tx.Exec(txCtx, "SELECT pg_advisory_xact_lock($1)", int64(h.Sum64())); err != nil {
			if txErr := tx.Rollback(ctx); txErr != nil {
				return fmt.Errorf("dbx.AdvisoryLockTxBegin: failed to rollback transaction: txErr=%w, err=%w", txErr, err)
			}
			return fmt.Errorf("dbx.AdvisoryLockTxBegin: %w", err)
		}

		return fn(txCtx)
	})
}

// Deprecated: Use AdvisoryLockRunInTx instead, which handles transaction
// lifecycle automatically and correctly joins existing transactions from context.
func AdvisoryLockTxBegin(ctx context.Context, q queriers.Querier, lockID string) (*queriers.TxQuerier, error) {
	tx, err := q.Begin(ctx)
	if err != nil {
		return nil, err
	}

	h := fnv.New64a()
	_, err = h.Write([]byte(lockID))
	if err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return nil, fmt.Errorf("dbx.AdvisoryLockTxBegin: failed to rollback transaction: txErr=%w, err=%w", txErr, err)
		}
		return nil, err
	}
	lockKey := int64(h.Sum64())

	_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", lockKey)
	if err != nil {
		if txErr := tx.Rollback(ctx); txErr != nil {
			return nil, fmt.Errorf("dbx.AdvisoryLockTxBegin: failed to rollback transaction: txErr=%w, err=%w", txErr, err)
		}
		return nil, fmt.Errorf("dbx.AdvisoryLockTxBegin: %w", err)
	}
	return tx, nil
}
