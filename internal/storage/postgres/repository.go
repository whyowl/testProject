package postgres

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type PgRepository struct {
	txManager TransactionManager
}

func NewPgRepository(txManager TransactionManager) *PgRepository {
	return &PgRepository{txManager: txManager}
}

func (r *PgRepository) InsertWallet(ctx context.Context, walletId uuid.UUID) error {

	tx := r.txManager.GetQueryEngine(ctx)

	query := "INSERT INTO wallets (wallet_id) VALUES ($1)"

	_, err := tx.Exec(ctx, query, walletId)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.New("wallet already exists")
		}
		return err
	}

	return nil
}

func (r *PgRepository) DeleteWallet(ctx context.Context, id uuid.UUID) error {

	tx := r.txManager.GetQueryEngine(ctx)

	query := "DELETE FROM wallets WHERE wallet_id = $1"

	_, err := tx.Exec(ctx, query, id.String())
	if err != nil {
		return err
	}

	return nil
}

func (r *PgRepository) GetById(ctx context.Context, walletId uuid.UUID) (int, error) {

	tx := r.txManager.GetQueryEngine(ctx)

	query := "SELECT balance FROM wallets WHERE wallet_id = $1"
	row := tx.QueryRow(ctx, query, walletId)

	var balance int
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, errors.New("wallet not found")
		}
		return 0, err
	}
	return balance, nil
}

func (r *PgRepository) LockBalance(ctx context.Context, walletId uuid.UUID) error {
	tx := r.txManager.GetQueryEngine(ctx)

	query := "SELECT balance FROM wallets WHERE wallet_id = $1 FOR UPDATE;"

	_, err := tx.Exec(ctx, query, walletId)
	if err != nil {
		return err
	}
	return nil
}

func (r *PgRepository) UpdateBalance(ctx context.Context, walletId uuid.UUID, balanceDiff int) error {
	tx := r.txManager.GetQueryEngine(ctx)

	query := "UPDATE wallets SET balance = balance + $2 WHERE wallet_id = $1;"

	tag, err := tx.Exec(ctx, query, walletId, balanceDiff)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("wallet not found")
	}
	return nil
}
