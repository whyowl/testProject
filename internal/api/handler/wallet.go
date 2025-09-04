package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var req WalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	parsedWalletID, err := uuid.Parse(req.WalletID)
	if err != nil || parsedWalletID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "invalid walletId parameter")
		return
	}

	switch req.OperationType {
	case Deposit:
		if err := h.s.DepositFunds(ctx, parsedWalletID, req.Amount); err != nil {
			status := http.StatusInternalServerError
			switch err.Error() {
			case "wallet not found":
				status = http.StatusNotFound
			case "amount must be positive":
				status = http.StatusBadRequest
			}
			respondError(w, status, err.Error())
			return
		}
	case Withdraw:
		if err := h.s.WithdrawFunds(ctx, parsedWalletID, req.Amount); err != nil {
			status := http.StatusInternalServerError
			switch err.Error() {
			case "wallet not found":
				status = http.StatusNotFound
			case "amount must be positive":
				status = http.StatusBadRequest
			case "not enough balance":
				status = http.StatusBadRequest
			}
			respondError(w, status, err.Error())
			return
		}
	default:
		respondError(w, http.StatusBadRequest, "invalid operationType parameter")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "success"})
}

func (h *RestHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	var req CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	parsedWalletID, err := uuid.Parse(req.WalletID)
	if err != nil || parsedWalletID == uuid.Nil {
		respondError(w, http.StatusBadRequest, "invalid walletId parameter")
		return
	}

	if err := h.s.CreateWallet(ctx, parsedWalletID); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "wallet already exists" {
			status = http.StatusConflict
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"status": "success"})
}

func (h *RestHandler) GetBalance(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	walletIdStr := chi.URLParam(r, "walletId")
	walletId, err := uuid.Parse(walletIdStr)
	if err != nil || walletId == uuid.Nil {
		respondError(w, http.StatusBadRequest, "invalid walletId parameter")
		return
	}

	balance, err := h.s.GetBalance(ctx, walletId)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "wallet not found" {
			status = http.StatusNotFound
		}
		respondError(w, status, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"walletId": walletId.String(), "balance": balance})
}
