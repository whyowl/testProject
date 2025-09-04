package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
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

func newHandler(ff *fakeFacade) *RestHandler {
	ws := service.NewWalletService(ff)
	return NewHandler(ws)
}

func doJSONReq(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestTransferFunds_BadJSON(t *testing.T) {
	h := newHandler(&fakeFacade{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.TransferFunds(w, req)

	require.Equalf(t, http.StatusBadRequest, w.Code, "want 400, got %d", w.Code)
}

func TestTransferFunds_InvalidWalletID(t *testing.T) {
	h := newHandler(&fakeFacade{})
	req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
		"walletId":      "not-a-uuid",
		"operationType": "DEPOSIT",
		"amount":        10,
	})
	w := httptest.NewRecorder()

	h.TransferFunds(w, req)

	require.Equalf(t, http.StatusBadRequest, w.Code, "want 400, got %d", w.Code)
}

func TestTransferFunds_InvalidOperationType(t *testing.T) {
	h := newHandler(&fakeFacade{})
	req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
		"walletId":      uuid.New().String(),
		"operationType": "MOVE",
		"amount":        10,
	})
	w := httptest.NewRecorder()

	h.TransferFunds(w, req)

	require.Equalf(t, http.StatusBadRequest, w.Code, "want 400, got %d", w.Code)
}

func TestTransferFunds_Deposit_ErrorMapping(t *testing.T) {
	id := uuid.New()

	{
		ff := &fakeFacade{}
		h := newHandler(ff)
		ff.depositErr = errAny("amount must be positive")

		req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
			"walletId":      id.String(),
			"operationType": "DEPOSIT",
			"amount":        0,
		})
		w := httptest.NewRecorder()
		h.TransferFunds(w, req)
		require.Equalf(t, http.StatusBadRequest, w.Code, "want 400, got %d", w.Code)
	}

	{
		ff := &fakeFacade{depositErr: errAny("wallet not found")}
		h := newHandler(ff)
		req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
			"walletId":      id.String(),
			"operationType": "DEPOSIT",
			"amount":        10,
		})
		w := httptest.NewRecorder()
		h.TransferFunds(w, req)
		require.Equalf(t, http.StatusNotFound, w.Code, "want 404, got %d", w.Code)
	}

	{
		ff := &fakeFacade{depositErr: errAny("db down")}
		h := newHandler(ff)
		req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
			"walletId":      id.String(),
			"operationType": "DEPOSIT",
			"amount":        10,
		})
		w := httptest.NewRecorder()
		h.TransferFunds(w, req)
		require.Equalf(t, http.StatusInternalServerError, w.Code, "want 500, got %d", w.Code)
	}
}

func TestTransferFunds_Withdraw_ErrorMapping(t *testing.T) {
	id := uuid.New()

	{
		ff := &fakeFacade{withdrawErr: errAny("not enough balance")}
		h := newHandler(ff)
		req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
			"walletId":      id.String(),
			"operationType": "WITHDRAW",
			"amount":        100,
		})
		w := httptest.NewRecorder()
		h.TransferFunds(w, req)
		require.Equalf(t, http.StatusBadRequest, w.Code, "want 400, got %d", w.Code)
	}
}

func TestTransferFunds_Success_DepositAndWithdraw(t *testing.T) {
	id := uuid.New()

	// deposit ok
	{
		ff := &fakeFacade{}
		h := newHandler(ff)
		req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
			"walletId":      id.String(),
			"operationType": "DEPOSIT",
			"amount":        150,
		})
		w := httptest.NewRecorder()
		h.TransferFunds(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if ff.lastDepositID != id || ff.lastAmount != 150 {
			t.Fatalf("deposit not called as expected")
		}
	}

	// withdraw ok
	{
		ff := &fakeFacade{}
		h := newHandler(ff)
		req := doJSONReq(http.MethodPost, "/api/v1/wallet", map[string]any{
			"walletId":      id.String(),
			"operationType": "WITHDRAW",
			"amount":        70,
		})
		w := httptest.NewRecorder()
		h.TransferFunds(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if ff.lastWithdrawID != id || ff.lastAmount != 70 {
			t.Fatalf("withdraw not called as expected")
		}
	}
}

// -------- CreateWallet --------

func TestCreateWallet_BadJSON(t *testing.T) {
	h := newHandler(&fakeFacade{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/new", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateWallet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestCreateWallet_InvalidWalletID(t *testing.T) {
	h := newHandler(&fakeFacade{})
	req := doJSONReq(http.MethodPost, "/api/v1/wallets/new", map[string]any{"walletId": "nope"})
	w := httptest.NewRecorder()

	h.CreateWallet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestCreateWallet_Error500(t *testing.T) {
	ff := &fakeFacade{createErr: errAny("db down")}
	h := newHandler(ff)
	req := doJSONReq(http.MethodPost, "/api/v1/wallets/new", map[string]any{"walletId": uuid.New().String()})
	w := httptest.NewRecorder()

	h.CreateWallet(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestCreateWallet_Success(t *testing.T) {
	ff := &fakeFacade{}
	h := newHandler(ff)
	id := uuid.New()
	req := doJSONReq(http.MethodPost, "/api/v1/wallets/new", map[string]any{"walletId": id.String()})
	w := httptest.NewRecorder()

	h.CreateWallet(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if ff.lastCreateID != id {
		t.Fatalf("create not called as expected")
	}
}

// -------- GetBalance --------

func TestGetBalance_InvalidWalletID(t *testing.T) {
	h := newHandler(&fakeFacade{})
	r := chi.NewRouter()
	r.Get("/api/v1/wallets/{walletId}", h.GetBalance)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/not-a-uuid", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestGetBalance_Error500(t *testing.T) {
	ff := &fakeFacade{getErr: errAny("db down")}
	h := newHandler(ff)
	r := chi.NewRouter()
	r.Get("/api/v1/wallets/{walletId}", h.GetBalance)

	id := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+id.String(), nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", w.Code)
	}
}

func TestGetBalance_Success(t *testing.T) {
	ff := &fakeFacade{getBal: 555}
	h := newHandler(ff)
	r := chi.NewRouter()
	r.Get("/api/v1/wallets/{walletId}", h.GetBalance)

	id := uuid.New()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/"+id.String(), nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp struct {
		WalletID string `json:"walletId"`
		Balance  int64  `json:"balance"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.WalletID != id.String() || resp.Balance != 555 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

type errAny string

func (e errAny) Error() string { return string(e) }