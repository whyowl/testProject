package storage

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"project/internal/storage/mocks"
)

func TestDeposit_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := uuid.New()

	tm := mocks.NewMockTransactionManager(ctrl)
	repo := mocks.NewMockWalletRepo(ctrl)

	tm.
		EXPECT().
		RunSerializable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctxTx context.Context) error) error {
			return fn(ctx)
		})

	gomock.InOrder(
		repo.EXPECT().LockBalance(gomock.Any(), id).Return(nil),
		repo.EXPECT().UpdateBalance(gomock.Any(), id, int64(150)).Return(nil),
	)

	f := NewStorageFacade(tm, repo)

	if err := f.Deposit(ctx, id, 150); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithdraw_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := uuid.New()

	tm := mocks.NewMockTransactionManager(ctrl)
	repo := mocks.NewMockWalletRepo(ctrl)

	tm.
		EXPECT().
		RunSerializable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctxTx context.Context) error) error {
			return fn(ctx)
		})

	gomock.InOrder(
		repo.EXPECT().LockBalance(gomock.Any(), id).Return(nil),
		repo.EXPECT().GetById(gomock.Any(), id).Return(int64(200), nil),
		repo.EXPECT().UpdateBalance(gomock.Any(), id, int64(-150)).Return(nil),
	)

	f := NewStorageFacade(tm, repo)

	require.NoError(t, f.Withdraw(ctx, id, 150))
}

func TestWithdraw_NotEnoughBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := uuid.New()

	tm := mocks.NewMockTransactionManager(ctrl)
	repo := mocks.NewMockWalletRepo(ctrl)

	tm.
		EXPECT().
		RunSerializable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctxTx context.Context) error) error {
			return fn(ctx)
		})

	gomock.InOrder(
		repo.EXPECT().LockBalance(gomock.Any(), id).Return(nil),
		repo.EXPECT().GetById(gomock.Any(), id).Return(int64(100), nil),
	)

	f := NewStorageFacade(tm, repo)

	require.EqualError(t, f.Withdraw(ctx, id, 150), "not enough balance: 100 < 150")
}

func TestWithdraw_LockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := uuid.New()

	tm := mocks.NewMockTransactionManager(ctrl)
	repo := mocks.NewMockWalletRepo(ctrl)

	tm.
		EXPECT().
		RunSerializable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctxTx context.Context) error) error {
			return fn(ctx)
		})

	gomock.InOrder(
		repo.EXPECT().LockBalance(gomock.Any(), id).Return(errAny("lock-fail")),
	)

	f := NewStorageFacade(tm, repo)

	require.EqualError(t, f.Withdraw(ctx, id, 150), "lock-fail")
}

func TestDeposit_LockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := uuid.New()

	tm := mocks.NewMockTransactionManager(ctrl)
	repo := mocks.NewMockWalletRepo(ctrl)

	tm.
		EXPECT().
		RunSerializable(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(ctxTx context.Context) error) error {
			return fn(ctx)
		})

	gomock.InOrder(
		repo.EXPECT().LockBalance(gomock.Any(), id).Return(errAny("lock-fail")),
	)

	f := NewStorageFacade(tm, repo)

	require.EqualError(t, f.Deposit(ctx, id, 150), "lock-fail")
}

func TestDeposit_TxErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	id := uuid.New()

	tm := mocks.NewMockTransactionManager(ctrl)
	repo := mocks.NewMockWalletRepo(ctrl)

	tm.
		EXPECT().
		RunSerializable(gomock.Any(), gomock.Any()).
		Return(errAny("tx-fail"))

	f := NewStorageFacade(tm, repo)

	require.EqualError(t, f.Deposit(ctx, id, 150), "tx-fail")
}

type errAny string

func (e errAny) Error() string { return string(e) }
