package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"project/internal/storage/postgres"
)

type Facade interface {
	Deposit(ctx context.Context, walletId uuid.UUID, amount int) error
	Withdraw(ctx context.Context, walletId uuid.UUID, amount int) error
	GetByID(ctx context.Context, walletId uuid.UUID) (int, error)
	Create(ctx context.Context, walletId uuid.UUID) error
}

type StorageFacade struct {
	txManager    postgres.TransactionManager
	pgRepository *postgres.PgRepository
}

func NewStorageFacade(txManager postgres.TransactionManager, pgRepository *postgres.PgRepository) Facade {
	return &StorageFacade{
		txManager:    txManager,
		pgRepository: pgRepository,
	}
}

func (f *StorageFacade) Deposit(ctx context.Context, walletId uuid.UUID, amount int) error {
	return f.txManager.RunSerializable(ctx, func(ctxTx context.Context) error {

		if err := f.pgRepository.LockBalance(ctxTx, walletId); err != nil {
			return err
		}

		if err := f.pgRepository.UpdateBalance(ctxTx, walletId, amount); err != nil {
			return err
		}

		return nil
	})
}

func (f *StorageFacade) Withdraw(ctx context.Context, walletId uuid.UUID, amount int) error {
	return f.txManager.RunSerializable(ctx, func(ctxTx context.Context) error {

		if err := f.pgRepository.LockBalance(ctxTx, walletId); err != nil {
			return err
		}

		balance, err := f.pgRepository.GetById(ctxTx, walletId)
		if err != nil {
			return err
		}

		if balance < amount {
			return fmt.Errorf("not enough balance: %d < %d", balance, amount)
		}

		if err := f.pgRepository.UpdateBalance(ctxTx, walletId, -amount); err != nil {
			return err
		}

		return nil
	})
}

func (f *StorageFacade) GetByID(ctx context.Context, walletId uuid.UUID) (int, error) {
	return f.pgRepository.GetById(ctx, walletId)
}

func (f *StorageFacade) Create(ctx context.Context, walletId uuid.UUID) error {
	return f.pgRepository.InsertWallet(ctx, walletId)
}
