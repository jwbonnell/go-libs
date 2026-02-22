package dbx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jwbonnell/go-libs/pkg/dbx/queriers"
)

// Exec is a thin wrapper around pgx.Exec that supports passing in generic struct args and they
// are automatically converted to pgx.NamedArgs
func Exec[T any](ctx context.Context, d queriers.Querier, sql string, args T) error {
	namedArgs, err := StructToNamedArgs(args)
	_, err = d.Exec(ctx, sql, namedArgs)
	if err != nil {
		return fmt.Errorf("dbx.Querier.Exec: %w", err)
	}
	return nil
}

// ExecReturn executes a SQL statement is expected to return a single row that needs to be mapped to a struct.
func ExecReturn[S any, D any](ctx context.Context, d queriers.Querier, sql string, dest *D, args S) error {
	namedArgs, err := StructToNamedArgs(args)
	rows, err := d.Query(ctx, sql, namedArgs)
	if err != nil {
		return fmt.Errorf("dbx.Querier.Query: %w", err)
	}
	defer rows.Close()

	// Use pgx.CollectOneRow with RowToStructByName for struct mapping
	// If T is not a struct, the RowToStructByName will fail and Scan will be required.
	var got []D
	got, err = pgx.CollectRows(rows, pgx.RowToStructByName[D])
	if err != nil {
		return fmt.Errorf("pgx.CollectRows: %w", err)
	}
	if len(got) == 0 {
		return pgx.ErrNoRows
	}
	*dest = got[0]
	return nil
}

// ExecReturnMany functions the same as ExecReturn where a SQL statement is expected to return data that
// needs to be mapped to a struct. The difference is multiple rows are expected and are returned as a slice.
func ExecReturnMany[S any, D any](ctx context.Context, d queriers.Querier, sql string, dest []D, args S) error {
	namedArgs, err := StructToNamedArgs(args)
	rows, err := d.Query(ctx, sql, namedArgs)
	if err != nil {
		return fmt.Errorf("dbx.Querier.Query: %w", err)
	}
	defer rows.Close()

	// Use pgx.CollectOneRow with RowToStructByName for struct mapping
	// If T is not a struct, the RowToStructByName will fail and Scan will be required.
	dest, err = pgx.CollectRows(rows, pgx.RowToStructByName[D])
	if err != nil {
		return fmt.Errorf("pgx.CollectRows: %w", err)
	}
	return nil
}

func AdvisoryTransactionLock[T any](ctx context.Context, tx pgx.Tx, id int) error {
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", id)
	if err != nil {
		return fmt.Errorf("dbx.AdvisoryTransactionLock: %w", err)
	}
	return nil
}
