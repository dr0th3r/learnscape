package utils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
)

var tracerTransaction = otel.Tracer("transaction")

type txFunc func(pgx.Tx) error

func HandleTx(ctx context.Context, db *pgxpool.Pool, txFuncs ...txFunc) error {
	_, span := tracerTransaction.Start(ctx, "handle db transaction")
	defer span.End()

	span.AddEvent("Starting database transaction")
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, f := range txFuncs {
		span.AddEvent(fmt.Sprintf("Executing function: %d", i))
		if err := f(tx); err != nil {
			return err
		}
	}

	span.AddEvent("Commiting database transaction")
	return tx.Commit(ctx)
}
