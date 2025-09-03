package handler

import (
	"encoding/json"
	"net/http"
	"project/internal/service"
)

type RestHandler struct {
	s *service.WalletService
}

func NewHandler(svc *service.WalletService) *RestHandler {
	return &RestHandler{
		s: svc,
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
