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
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	CreateWithdrawal(ctx context.Context, userID int64, withdrawal model.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalData, error)
	GetOrdersForProcessing(ctx context.Context, limit int) ([]model.Order, error)
	UpdateOrderAccrual(ctx context.Context, order model.AccrualResponse) error
}

type Repository struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repository {
	return &Repository{db: db}
}
