package storage

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_storage.go -package=mocks project/internal/storage WalletRepo,Facade
//go:generate mockgen -destination=mocks/mock_postgres.go -package=mocks project/internal/storage/postgres TransactionManager

type WalletRepo interface {
	LockBalance(ctx context.Context, walletId uuid.UUID) error
	UpdateBalance(ctx context.Context, walletId uuid.UUID, balanceDiff int64) error
	GetById(ctx context.Context, walletId uuid.UUID) (int64, error)
	InsertWallet(ctx context.Context, walletId uuid.UUID) error
}
