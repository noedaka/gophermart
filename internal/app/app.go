package app

import (
	"context"
	"database/sql"
	"gophermart/internal/client"
	"gophermart/internal/config"
	dbConfig "gophermart/internal/config/db"
	"gophermart/internal/handler"
	"gophermart/internal/middleware"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"gophermart/internal/worker"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
)

func Run() error {
	cfg, err := config.Init()
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	if err := dbConfig.InitDB(db); err != nil {
		return err
	}

	userRepo := repository.NewRepo(db)
	service := service.NewService(userRepo)
	accrualClient := client.NewClient(cfg.AccrualSystemAddress)
    accrualWorker := worker.NewWorker(userRepo, accrualClient)
	go accrualWorker.Start(context.Background())

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

	if err := http.ListenAndServe(cfg.ServerAdress, r); err != nil {
		return err
	}

	return nil
}
