package app

import (
	"database/sql"
	"gophermart/internal/config"
	dbConfig "gophermart/internal/config/db"
	"gophermart/internal/handler"
	"gophermart/internal/middleware"
	"gophermart/internal/repository"
	"gophermart/internal/service"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Service service.Service
}

func NewApp(db *sql.DB) *App {
	userRepo := repository.NewRepo(db)
	service := service.NewService(userRepo)

	return &App{
		Service: service,
	}
}

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

	app := NewApp(db)

	handler := handler.NewHandler(app.Service)

	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {

		r.Route("/api", func(r chi.Router) {
			r.Route("/user", func(r chi.Router) {
				r.Post("/register", handler.RegisterHandler)
				r.Post("/login", handler.LoginHandler)

				r.Group(func(r chi.Router) {
					r.Use(middleware.AuthMiddleware)
				})
			})
		})
	})

	if err := http.ListenAndServe(cfg.ServerAdress, r); err != nil {
		return err
	}

	return nil
}
