package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Service interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*DTO, error)
	GetUser(ctx context.Context, userID string) (*DTO, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SetIsActive(ctx context.Context, userID string, isActive bool) (*DTO, error) {
	logger.Info("setting user active status",
		zap.String("user_id", userID),
		zap.Bool("is_active", isActive))

	// Проверка существования пользователя
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("failed to get user", zap.String("user_id", userID), zap.Error(err))
		return nil, err
	}

	if user == nil {
		logger.Warn("user not found", zap.String("user_id", userID))
		return nil, ErrUserNotFound
	}

	// Обновляем статус активности
	if err := s.repo.SetIsActive(ctx, userID, isActive); err != nil {
		logger.Error("failed to set user active status", zap.Error(err))
		return nil, fmt.Errorf("set is_active: %w", err)
	}

	updatedUser, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return updatedUser.ToDTO(), nil
}

func (s *service) GetUser(ctx context.Context, userID string) (*DTO, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToDTO(), nil
}
