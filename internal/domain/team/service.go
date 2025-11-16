package team

import (
	"context"
	"errors"
	"fmt"

	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"go.uber.org/zap"
)

var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")
)

type Service interface {
	CreateTeam(ctx context.Context, dto *CreateTeamDTO) (*DTO, error)
	GetTeam(ctx context.Context, teamName string) (*DTO, error)
}

type service struct {
	teamRepo Repository
	userRepo user.Repository
}

func NewService(teamRepo Repository, userRepo user.Repository) Service {
	return &service{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *service) CreateTeam(ctx context.Context, dto *CreateTeamDTO) (*DTO, error) {
	logger.Info("creating team", zap.String("team_name", dto.TeamName))

	// Проверка существования команды
	exists, err := s.teamRepo.Exists(ctx, dto.TeamName)
	if err != nil {
		logger.Error("failed to check team existence", zap.Error(err))
		return nil, fmt.Errorf("check team existence: %w", err)
	}

	if exists {
		logger.Warn("team already exists", zap.String("team_name", dto.TeamName))
		return nil, ErrTeamAlreadyExists
	}

	teamModel := FromCreateDTO(dto)

	// Сохранение команды
	if err := s.teamRepo.Create(ctx, teamModel); err != nil {
		logger.Error("failed to create team", zap.Error(err))
		return nil, fmt.Errorf("create team: %w", err)
	}

	// Получаем ID созданной команды
	teamID, err := s.teamRepo.GetIDByName(ctx, dto.TeamName)
	if err != nil {
		logger.Error("failed to get team ID", zap.Error(err))
		return nil, fmt.Errorf("get team ID: %w", err)
	}

	// Создание/обновление (Upsert) пользователей команды
	for _, member := range dto.Members {
		userModel := &user.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamID:   teamID,
			IsActive: member.IsActive,
		}

		if err := s.userRepo.Upsert(ctx, userModel); err != nil {
			logger.Error("failed to upsert user",
				zap.String("user_id", member.UserID),
				zap.Error(err))
			return nil, fmt.Errorf("upsert user %s: %w", member.UserID, err)
		}
	}

	logger.Info("team created successfully", zap.String("team_name", dto.TeamName))

	team, err := s.teamRepo.GetByName(ctx, dto.TeamName)
	if err != nil {
		return nil, err
	}

	return team.ToDTO(), nil
}

func (s *service) GetTeam(ctx context.Context, teamName string) (*DTO, error) {
	logger.Debug("getting team", zap.String("team_name", teamName))

	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		logger.Error("failed to get team", zap.String("team_name", teamName), zap.Error(err))
		return nil, err
	}

	if team == nil {
		logger.Warn("team not found", zap.String("team_name", teamName))
		return nil, ErrTeamNotFound
	}

	return team.ToDTO(), nil
}
