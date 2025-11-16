package statistics

import (
	"context"
	"fmt"

	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"go.uber.org/zap"
)

type Service interface {
	GetUserStats(ctx context.Context, userID string) (*UserStatsDTO, error)
	GetTeamStats(ctx context.Context, teamName string) (*TeamStatsDTO, error)
	GetGlobalStats(ctx context.Context) (*GlobalStatsDTO, error)
	GetTopReviewers(ctx context.Context, limit int) ([]TopReviewerDTO, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// GetUserStats возвращает статистику для конкретного пользователя по его айди
func (s *service) GetUserStats(ctx context.Context, userID string) (*UserStatsDTO, error) {
	logger.Debug("getting user stats", zap.String("user_id", userID))

	stats, err := s.repo.GetUserStats(ctx, userID)
	if err != nil {
		logger.Error("failed to get user stats", zap.Error(err))
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	return stats.ToDTO(), nil
}

// GetTeamStats возвращает статистику для конкретной команды по имени
func (s *service) GetTeamStats(ctx context.Context, teamName string) (*TeamStatsDTO, error) {
	logger.Debug("getting team stats", zap.String("team_name", teamName))

	stats, err := s.repo.GetTeamStats(ctx, teamName)
	if err != nil {
		logger.Error("failed to get team stats", zap.Error(err))
		return nil, fmt.Errorf("get team stats: %w", err)
	}

	return stats.ToDTO(), nil
}

// GetGlobalStats возвращает глобальную статистику по всем пользователям и командам
func (s *service) GetGlobalStats(ctx context.Context) (*GlobalStatsDTO, error) {
	logger.Debug("getting global stats")

	stats, err := s.repo.GetGlobalStats(ctx)
	if err != nil {
		logger.Error("failed to get global stats", zap.Error(err))
		return nil, fmt.Errorf("get global stats: %w", err)
	}

	return stats.ToDTO(), nil
}

// GetTopReviewers возвращает список топ рецензентов по количеству ревью
func (s *service) GetTopReviewers(ctx context.Context, limit int) ([]TopReviewerDTO, error) {
	logger.Debug("getting top reviewers", zap.Int("limit", limit))

	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	reviewers, err := s.repo.GetTopReviewers(ctx, limit)
	if err != nil {
		logger.Error("failed to get top reviewers", zap.Error(err))
		return nil, fmt.Errorf("get top reviewers: %w", err)
	}

	return TopReviewersToDTO(reviewers), nil
}
