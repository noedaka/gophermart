package repository

import (
	"context"
	"database/sql"
	"gophermart/internal/model"
)

type UserRepository interface {
	GetUserByCredentials(ctx context.Context, userCredentials model.UserCredentials) (int64, error)
	CreateUserWithCredentials(ctx context.Context, userCredentials model.UserCredentials) (int64, error)
	CreateOrder(ctx context.Context, userID int64, orderNumber string) error
	GetOrdersByUserID(ctx context.Context, userID int64) ([]model.Order, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repository {
	return &Repository{db: db}
}
