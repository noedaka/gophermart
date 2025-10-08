package app

import (
	"context"
	"database/sql"
	"gophermart/internal/client"
	"gophermart/internal/config"
	dbconfig "gophermart/internal/config/db"
	"gophermart/internal/handler"
	"gophermart/internal/middleware"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"gophermart/internal/worker"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
)

const workersCount = 3

func Run() error {
	cfg, err := config.Init()
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	if err := dbconfig.Migrate(db); err != nil {
		return err
	}

	userRepo := repository.NewRepo(db)
	service := service.NewService(userRepo)
	accrualClient := client.NewClient(cfg.AccrualSystemAddress)
	accrualWorker := worker.NewWorker(userRepo, accrualClient, workersCount)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go accrualWorker.Start(ctx)

	handler := handler.NewHandler(service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {

		r.Route("/api", func(r chi.Router) {
			r.Route("/user", func(r chi.Router) {
				r.Post("/register", handler.RegisterHandler)
				r.Post("/login", handler.LoginHandler)

				r.Group(func(r chi.Router) {
					r.Use(middleware.AuthMiddleware)

					r.Post("/orders", handler.CreateOrderHandler)
					r.Get("/orders", handler.GetOrdersHandler)
					r.Get("/balance", handler.GetBalanceHandler)
					r.Post("/balance/withdraw", handler.CreateWithdrawalHandler)
					r.Get("/withdrawals", handler.GetWithdrawalsHandler)
				})
			})
		})
	})

	server := &http.Server{
		Addr:    cfg.ServerAdress,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown серва
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// Останавливаем воркеры 
	log.Println("Stopping workers...")
	cancel()

	time.Sleep(1 * time.Second)

	return nil
}
