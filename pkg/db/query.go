package db

import (
	"context"
	"fmt"
	"github.com/jwbonnell/go-libs/pkg/db/queriers"

	"github.com/jackc/pgx/v5"
)

// QueryOne database record
func QueryOne[T any](ctx context.Context, q queriers.Querier, sql string, dest *T, namedArgs pgx.NamedArgs) error {
	rows, err := q.Query(ctx, sql, namedArgs)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	// Use pgx.CollectOneRow with RowToStructByName for struct mapping
	// If T is not a struct, the RowToStructByName will fail and Scan will be required.
	var got []T
	got, err = pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return fmt.Errorf("collect row: %w", err)
	}
	if len(got) == 0 {
		return pgx.ErrNoRows
	}
	*dest = got[0]
	return nil
}

// Query multiple database records
func Query[T any](ctx context.Context, q queriers.Querier, sql string, dest *[]T, namedArgs pgx.NamedArgs) error {
	rows, err := q.Query(ctx, sql, namedArgs)
	if err != nil {
		return fmt.Errorf("query named: %w", err)
	}
	defer rows.Close()

	vals, err := pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return fmt.Errorf("collect rows: %w", err)
	}
	*dest = vals
	return nil
}
