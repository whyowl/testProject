package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"project/internal/service"
)

type fakeFacade struct {
	depositErr  error
	withdrawErr error
	getBal      int64
	getErr      error
	createErr   error

	lastDepositID  uuid.UUID
	lastWithdrawID uuid.UUID
	lastGetID      uuid.UUID
	lastCreateID   uuid.UUID
	lastAmount     int64
}

func (f *fakeFacade) Deposit(ctx context.Context, walletId uuid.UUID, amount int64) error {
	f.lastDepositID = walletId
	f.lastAmount = amount
	return f.depositErr
}
func (f *fakeFacade) Withdraw(ctx context.Context, walletId uuid.UUID, amount int64) error {
	f.lastWithdrawID = walletId
	f.lastAmount = amount
	return f.withdrawErr
}
func (f *fakeFacade) GetByID(ctx context.Context, walletId uuid.UUID) (int64, error) {
	f.lastGetID = walletId
	if f.getErr != nil {
		return 0, f.getErr
	}
	return f.getBal, nil
}
func (f *fakeFacade) Create(ctx context.Context, walletId uuid.UUID) error {
	f.lastCreateID = walletId
	return f.createErr
}

func newTestServer() (*Router, *fakeFacade) {
	ff := &fakeFacade{getBal: 123}
	ws := service.NewWalletService(ff)
	rt := SetupRouter(ws)
	return rt, ff
}

func doReq(r http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var rc io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rc = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rc)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRootOK(t *testing.T) {
	rt, _ := newTestServer()
	w := doReq(rt.r, http.MethodGet, "/", nil)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "OK", w.Body.String())
}

func TestTransferFunds_Deposit_Success(t *testing.T) {
	rt, ff := newTestServer()
	id := uuid.New()

	w := doReq(rt.r, http.MethodPost, "/api/v1/wallet", map[string]any{
		"walletId":      id.String(),
		"operationType": "DEPOSIT",
		"amount":        150,
	})

	require.Equal(t, http.StatusOK, w.Code)
	require.Equalf(t, id, ff.lastDepositID, "facade not called as expected: id=%v amount=%d", ff.lastDepositID, ff.lastAmount)
	require.Equalf(t, int64(150), ff.lastAmount, "facade not called as expected: id=%v amount=%d", ff.lastDepositID, ff.lastAmount)
}

func TestTransferFunds_Withdraw_Success(t *testing.T) {
	rt, ff := newTestServer()
	id := uuid.New()

	w := doReq(rt.r, http.MethodPost, "/api/v1/wallet", map[string]any{
		"walletId":      id.String(),
		"operationType": "WITHDRAW",
		"amount":        70,
	})

	require.Equal(t, http.StatusOK, w.Code)
	require.Equalf(t, id, ff.lastWithdrawID, "facade not called as expected: id=%v amount=%d", ff.lastWithdrawID, ff.lastAmount)
	require.Equalf(t, int64(70), ff.lastAmount, "facade not called as expected: id=%v amount=%d", ff.lastWithdrawID, ff.lastAmount)
}

func TestGetBalance_Success(t *testing.T) {
	rt, ff := newTestServer()
	id := uuid.New()
	ff.getBal = 777

	w := doReq(rt.r, http.MethodGet, "/api/v1/wallets/"+id.String(), nil)

	require.Equalf(t, http.StatusOK, w.Code, "want 200, got %d body=%s", w.Code, w.Body.Bytes())
	require.Equalf(t, id, ff.lastGetID, "GetByID not called as expected")

	var resp struct {
		WalletID string `json:"walletId"`
		Balance  int64  `json:"balance"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	require.Equalf(t, id.String(), resp.WalletID, "unexpected walletId in response")
	require.Equalf(t, int64(777), resp.Balance, "unexpected balance in response")
}

func TestCreateWallet_Success(t *testing.T) {
	rt, ff := newTestServer()
	id := uuid.New()

	w := doReq(rt.r, http.MethodPost, "/api/v1/wallets/new", map[string]any{
		"walletId": id.String(),
	})

	require.Equalf(t, http.StatusOK, w.Code, "want 200, got %d body=%s", w.Code, w.Body.Bytes())
	require.Equalf(t, id, ff.lastCreateID, "Create not called as expected")
}

func TestUnknownRoute_404(t *testing.T) {
	rt, _ := newTestServer()
	w := doReq(rt.r, http.MethodGet, "/nope", nil)
	require.Equalf(t, http.StatusNotFound, w.Code, "want 404, got %d body=%s", w.Code, w.Body.Bytes())
}
