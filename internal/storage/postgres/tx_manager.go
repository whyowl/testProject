package postgres

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type txManagerKey struct{}

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (tm *TxManager) RunSerializable(ctx context.Context, fn func(ctxTx context.Context) error) error {
	options := pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}
	return tm.beginFunc(ctx, options, fn)
}

func (tm *TxManager) RunReadUncommitted(ctx context.Context, fn func(ctxTx context.Context) error) error {
	options := pgx.TxOptions{
		IsoLevel:   pgx.ReadUncommitted,
		AccessMode: pgx.ReadOnly,
	}
	return tm.beginFunc(ctx, options, fn)
}

func (tm *TxManager) beginFunc(ctx context.Context, txOptions pgx.TxOptions, fn func(ctxTx context.Context) error) error {
	tx, err := tm.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ctx = context.WithValue(ctx, txManagerKey{}, tx)
	if err := fn(ctx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (tm *TxManager) GetQueryEngine(ctx context.Context) QueryEngine {
	tx, ok := ctx.Value(txManagerKey{}).(QueryEngine)
	if ok && tx != nil {
		return tx
	}
	return tm.pool
}
