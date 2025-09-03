package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

type WalletOperationType string

const (
	Deposit  WalletOperationType = "DEPOSIT"
	Withdraw WalletOperationType = "WITHDRAW"
)

type WalletRequest struct {
	WalletID      string              `json:"walletId"`
	OperationType WalletOperationType `json:"operationType"`
	Amount        int64               `json:"amount"`
}

type CreateWalletRequest struct {
	WalletID string `json:"walletId"`
}

func (h *RestHandler) TransferFunds(w http.ResponseWriter, r *http.Request) {
	var req WalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsedWalletID, err := uuid.Parse(req.WalletID)
	if err != nil || parsedWalletID == uuid.Nil {
		http.Error(w, "invalid walletId parameter", http.StatusBadRequest)
		return
	}

	switch req.OperationType {
	case Deposit:
		if err := h.s.DepositFunds(r.Context(), parsedWalletID, int(req.Amount)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case Withdraw:
		if err := h.s.WithdrawFunds(r.Context(), parsedWalletID, int(req.Amount)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "invalid operationType parameter", http.StatusBadRequest)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "success"})
}

func (h *RestHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parsedWalletID, err := uuid.Parse(req.WalletID)
	if err != nil || parsedWalletID == uuid.Nil {
		http.Error(w, "invalid walletId parameter", http.StatusBadRequest)
		return
	}

	if err := h.s.CreateWallet(r.Context(), parsedWalletID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "success"})
}

func (h *RestHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	walletIdStr := chi.URLParam(r, "walletId")
	walletId, err := uuid.Parse(walletIdStr)
	if err != nil || walletId == uuid.Nil {
		http.Error(w, "invalid walletId parameter", http.StatusBadRequest)
		return
	}

	balance, err := h.s.GetBalance(r.Context(), walletId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"walletId": walletId.String(), "balance": balance})
}
