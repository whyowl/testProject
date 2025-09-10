package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"os"
	"os/signal"
	"project/internal/api"
	"project/internal/config"
	"project/internal/service"
	"project/internal/storage"
	"project/internal/storage/postgres"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sig
		cancel()
	}()

	cfg := config.Load()

	pool, err := pgxpool.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	WalletService := service.NewWalletService(InitStorage(pool))

	router := api.SetupRouter(WalletService)

	go func() {
		err := router.Run(cfg.ApiAddress)
		if err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	ctxSvr, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := router.Stop(ctxSvr); err != nil {
		log.Printf("Failed to stop server:", err)
	}
	time.Sleep(7 * time.Second)
}

func InitStorage(pool *pgxpool.Pool) storage.Facade {
	txMngr := postgres.NewTxManager(pool)
	pgRepo := postgres.NewPgRepository(txMngr)

	return storage.NewStorageFacade(txMngr, pgRepo)
}