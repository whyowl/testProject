package service

import (
	"context"
	"errors"
	"project/internal/storage"

	"github.com/google/uuid"
)

type WalletService struct {
	Repo storage.Facade
}

func NewWalletService(repo storage.Facade) *WalletService {
	return &WalletService{
		Repo: repo,
	}
}

func (ws *WalletService) DepositFunds(ctx context.Context, walletId uuid.UUID, amount int64) error {

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if err := ws.Repo.Deposit(ctx, walletId, amount); err != nil {
		return err
	}

	return nil
}

func (ws *WalletService) WithdrawFunds(ctx context.Context, walletId uuid.UUID, amount int64) error {

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	if err := ws.Repo.Withdraw(ctx, walletId, amount); err != nil {
		return err
	}

	return nil
}

func (ws *WalletService) GetBalance(ctx context.Context, walletId uuid.UUID) (int64, error) {
	return ws.Repo.GetByID(ctx, walletId)
}

func (ws *WalletService) CreateWallet(ctx context.Context, walletId uuid.UUID) error {
	if walletId == uuid.Nil {
		return errors.New("walletId parameter is required")
	}

	if err := ws.Repo.Create(ctx, walletId); err != nil {
		return err
	}

	return nil
}
