package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"project/internal/storage"

	"github.com/google/uuid"
)

type mockFacade struct {
	OnDeposit  func(ctx context.Context, walletId uuid.UUID, amount int64) error
	OnWithdraw func(ctx context.Context, walletId uuid.UUID, amount int64) error
	OnGetByID  func(ctx context.Context, walletId uuid.UUID) (int64, error)
	OnCreate   func(ctx context.Context, walletId uuid.UUID) error

	depositCalls  int
	withdrawCalls int
	getByIDCalls  int
	createCalls   int
}

var _ storage.Facade = (*mockFacade)(nil)

func (m *mockFacade) Deposit(ctx context.Context, walletId uuid.UUID, amount int64) error {
	m.depositCalls++
	if m.OnDeposit != nil {
		return m.OnDeposit(ctx, walletId, amount)
	}
	return nil
}

func (m *mockFacade) Withdraw(ctx context.Context, walletId uuid.UUID, amount int64) error {
	m.withdrawCalls++
	if m.OnWithdraw != nil {
		return m.OnWithdraw(ctx, walletId, amount)
	}
	return nil
}

func (m *mockFacade) GetByID(ctx context.Context, walletId uuid.UUID) (int64, error) {
	m.getByIDCalls++
	if m.OnGetByID != nil {
		return m.OnGetByID(ctx, walletId)
	}
	return 0, nil
}

func (m *mockFacade) Create(ctx context.Context, walletId uuid.UUID) error {
	m.createCalls++
	if m.OnCreate != nil {
		return m.OnCreate(ctx, walletId)
	}
	return nil
}

func TestDepositFunds(t *testing.T) {
	ws := NewWalletService(&mockFacade{})

	t.Run("amount must be positive", func(t *testing.T) {
		err := ws.DepositFunds(context.Background(), uuid.New(), 0)

		require.Equalf(t, "amount must be positive", err.Error(), "want error 'amount must be positive', got %v", err)
	})

	t.Run("ok path calls repo", func(t *testing.T) {
		m := &mockFacade{}
		called := false
		m.OnDeposit = func(ctx context.Context, id uuid.UUID, a int64) error {
			called = true
			if a != 100 {
				return errors.New("wrong amount")
			}
			return nil
		}
		ws.Repo = m
		err := ws.DepositFunds(context.Background(), uuid.New(), 100)

		require.NoError(t, err, "unexpected error: %v", err)
		require.Equalf(t, true, called, "repo.Deposit wasn't called")
		require.Equalf(t, 1, m.depositCalls, "repo.Deposit wasn't called exactly once")
	})
}

func TestWithdrawFunds(t *testing.T) {
	ws := NewWalletService(&mockFacade{})

	t.Run("amount must be positive", func(t *testing.T) {
		err := ws.WithdrawFunds(context.Background(), uuid.New(), -10)

		require.Equalf(t, "amount must be positive", err.Error(), "want error 'amount must be positive', got %v", err)
	})

	t.Run("ok path calls repo", func(t *testing.T) {
		m := &mockFacade{}
		called := false
		m.OnWithdraw = func(ctx context.Context, id uuid.UUID, a int64) error {
			called = true
			if a != 50 {
				return errors.New("wrong amount")
			}
			return nil
		}
		ws.Repo = m
		err := ws.WithdrawFunds(context.Background(), uuid.New(), 50)

		require.NoError(t, err, "unexpected error: %v", err)
		require.Equalf(t, true, called, "repo.Withdraw wasn't called")
		require.Equalf(t, 1, m.withdrawCalls, "repo.Withdraw wasn't called exactly once")
	})
}

func TestGetBalance(t *testing.T) {
	m := &mockFacade{}
	want := int64(777)
	m.OnGetByID = func(ctx context.Context, id uuid.UUID) (int64, error) {
		return want, nil
	}
	ws := NewWalletService(m)

	got, err := ws.GetBalance(context.Background(), uuid.New())

	require.NoError(t, err)

	require.Equal(t, want, got)

	require.Equalf(t, 1, m.getByIDCalls, "repo.GetByID wasn't called exactly once")
}

func TestCreateWallet(t *testing.T) {
	ws := NewWalletService(&mockFacade{})

	t.Run("nil uuid error", func(t *testing.T) {
		err := ws.CreateWallet(context.Background(), uuid.Nil)
		require.Equalf(t, "walletId parameter is required", err.Error(), "want error 'walletId parameter is required', got %v", err)
	})

	t.Run("ok path calls repo", func(t *testing.T) {
		m := &mockFacade{}
		ws.Repo = m
		id := uuid.New()
		err := ws.CreateWallet(context.Background(), id)
		require.NoError(t, err, "unexpected error: %v", err)
		require.Equalf(t, 1, m.createCalls, "repo.Create wasn't called exactly once")
	})
}
