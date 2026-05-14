package dbx

import (
	"context"

	"github.com/jwbonnell/go-libs/pkg/dbx/queriers"
)

type ctxKey int

const txKey ctxKey = 1

// SetTx stores a TxQuerier in the context.
func SetTx(ctx context.Context, tx *queriers.TxQuerier) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTx retrieves the TxQuerier from context, if present.
func GetTx(ctx context.Context) (*queriers.TxQuerier, bool) {
	tx, ok := ctx.Value(txKey).(*queriers.TxQuerier)
	return tx, ok
}

// ResolveQuerier returns the transaction from context if present,
// otherwise returns the provided fallback querier.
func ResolveQuerier(ctx context.Context, fallback queriers.Querier) queriers.Querier {
	if tx, ok := GetTx(ctx); ok {
		return tx
	}
	return fallback
}
