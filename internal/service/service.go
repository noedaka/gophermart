
package service

import (
	"context"
	"gophermart/internal/model"
	"gophermart/internal/repository"
)

type Service interface {
	Login(ctx context.Context, credentials model.UserCredentials) (int64, error)
	Register(ctx context.Context, credentials model.UserCredentials) (int64, error)
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