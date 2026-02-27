package main

import (
	"log"
	"net/http"

	"github.com/kasku/kasku/internal/config"
	"github.com/kasku/kasku/internal/db"
	"github.com/kasku/kasku/internal/handler"
	"github.com/kasku/kasku/internal/router"
	"github.com/kasku/kasku/internal/service"
	"github.com/kasku/kasku/internal/session"
	"github.com/kasku/kasku/internal/store"
)

func main() {
	cfg := config.Load()

	// 1. Open database
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	// 2. Run migrations (idempotent)
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	// 3. Session manager with custom SQLite store
	sessionMgr := session.NewManager(database, cfg.SessionSecret, cfg.SessionLifetime)

	// 4. Store layer
	stores := store.New(database)

	// 5. Service layer
	authSvc := service.NewAuthService(stores.Family, stores.User)
	walletSvc := service.NewWalletService(stores.Wallet)
	ledgerSvc := service.NewLedgerService(stores.Ledger, stores.Wallet)

	// 6. Handler layer
	handlers := &router.Handlers{
		Auth:      handler.NewAuthHandler(sessionMgr, authSvc),
		Dashboard: handler.NewDashboardHandler(sessionMgr, stores.Wallet, stores.Ledger),
		Wallet:    handler.NewWalletHandler(sessionMgr, walletSvc, stores.Wallet),
		Ledger:    handler.NewLedgerHandler(sessionMgr, ledgerSvc, stores.Wallet, stores.Ledger),
		Family:    handler.NewFamilyHandler(sessionMgr, authSvc, stores.User, cfg.AppURL),
	}

	// 7. Build HTTP router
	r := router.New(sessionMgr, stores, handlers)

	// 8. Start server
	addr := ":" + cfg.Port
	log.Printf("Kasku listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}
