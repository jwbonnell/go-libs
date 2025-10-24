package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jwbonnell/go-libs/pkg/db/queriers"
)

func Exec[T any](ctx context.Context, d queriers.Querier, sql string, args T) error {
	namedArgs, err := StructToNamedArgs(args)
	rows, err := d.Query(ctx, sql, namedArgs)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	defer rows.Close()
	return nil
}

func AdvisoryTransactionLock[T any](ctx context.Context, tx pgx.Tx, id int) error {
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", id)
	return err
}
