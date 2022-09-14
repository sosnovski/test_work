package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype/pgxtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type txKey struct{}

// injectTx injects transaction to context
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx extracts transaction from context
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

type Transactor struct {
	conn   *pgxpool.Pool
	logger *zap.Logger
}

func NewTransactor(conn *pgxpool.Pool, logger *zap.Logger) *Transactor {
	return &Transactor{conn: conn, logger: logger}
}

func (t *Transactor) WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	tx, err := t.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		if err := tx.Rollback(ctx); err != nil {
			t.logger.Error("transaction rollback", zap.Error(err))
		}
	}(tx, ctx)

	err = txFunc(injectTx(ctx, tx))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (t *Transactor) Conn(ctx context.Context) pgxtype.Querier {
	tx := extractTx(ctx)
	if tx != nil {
		return tx
	}

	return t.conn
}
