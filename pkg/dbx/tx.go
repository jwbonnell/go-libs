package dbx

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/jwbonnell/go-libs/pkg/dbx/queriers"
)

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
