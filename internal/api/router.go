package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"project/internal/api/handler"
	"project/internal/service"
)

type Router struct {
	r *chi.Mux
	s *http.Server
}

func SetupRouter(s *service.WalletService) *Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	h := handler.NewHandler(s)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/wallet", h.TransferFunds)
		r.Get("/wallets/{walletId}", h.GetBalance)
		r.Post("/wallets/new", h.CreateWallet)
	})

	return &Router{r: r}
}

func (router *Router) Run(addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: router.r,
	}
	router.s = srv

	return router.s.ListenAndServe()
}

func (router *Router) Stop(ctx context.Context) error { //TODO CHECK
	err := router.s.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
