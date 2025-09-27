package service

import (
	"context"
	"gophermart/internal/model"
	"gophermart/internal/repository"
)

type Service interface {
	Login(ctx context.Context, credentials model.UserCredentials) (int64, error)
	Register(ctx context.Context, credentials model.UserCredentials) (int64, error)
	CreateOrder(ctx context.Context, userID int64, orderNumber string) error
	GetOrders(ctx context.Context, userID int64) ([]model.Order, error)
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	CreateWithdrawal(ctx context.Context, userID int64, withdrawal model.Withdrawal) error
	GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalData, error)
}

type service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) Service {
	return &service{userRepo: userRepo}
}

func (s *service) Login(ctx context.Context, credentials model.UserCredentials) (int64, error) {
	return s.userRepo.GetUserByCredentials(ctx, credentials)
}

func (s *service) Register(ctx context.Context, credentials model.UserCredentials) (int64, error) {
	return s.userRepo.CreateUserWithCredentials(ctx, credentials)
}

func (s *service) CreateOrder(ctx context.Context, userID int64, orderNumber string) error {
	return s.userRepo.CreateOrder(ctx, userID, orderNumber)
}

func (s *service) GetOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	return s.userRepo.GetOrdersByUserID(ctx, userID)
}

func (s *service) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	return s.userRepo.GetBalance(ctx, userID)
}

func (s *service) CreateWithdrawal(ctx context.Context, userID int64, withdrawal model.Withdrawal) error {
	return s.userRepo.CreateWithdrawal(ctx, userID, withdrawal)
}

func (s *service) GetWithdrawals(ctx context.Context, userID int64) ([]model.WithdrawalData, error) {
	return s.userRepo.GetWithdrawals(ctx, userID)
}
