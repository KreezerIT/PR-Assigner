package pullrequest

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"time"

	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

var (
	ErrPRAlreadyExists     = errors.New("pull request already exists")
	ErrPRNotFound          = errors.New("pull request not found")
	ErrPRMerged            = errors.New("cannot modify merged PR")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate         = errors.New("no active replacement candidate in team")
	ErrAuthorNotFound      = errors.New("author not found")
)

type Service interface {
	CreatePR(ctx context.Context, dto *CreatePRDTO) (*DTO, error)
	MergePR(ctx context.Context, prID string) (*DTO, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*ReassignResultDTO, error)
	GetUserReviews(ctx context.Context, userID string) ([]ShortDTO, error)
}

type service struct {
	prRepo   Repository
	userRepo user.Repository
}

func NewService(prRepo Repository, userRepo user.Repository) Service {
	return &service{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *service) CreatePR(ctx context.Context, dto *CreatePRDTO) (*DTO, error) {
	logger.Info("creating PR",
		zap.String("pr_id", dto.PullRequestID),
		zap.String("author_id", dto.AuthorID))

	// Проверка существования PR
	exists, err := s.prRepo.Exists(ctx, dto.PullRequestID)
	if err != nil {
		logger.Error("failed to check PR existence", zap.Error(err))
		return nil, fmt.Errorf("check PR existence: %w", err)
	}
	if exists {
		logger.Warn("PR already exists", zap.String("pr_id", dto.PullRequestID))
		return nil, ErrPRAlreadyExists
	}

	// Получение автора
	author, err := s.userRepo.GetByID(ctx, dto.AuthorID)
	if err != nil {
		logger.Error("failed to get author", zap.String("author_id", dto.AuthorID), zap.Error(err))
		return nil, err
	}
	if author == nil {
		logger.Warn("author not found", zap.String("author_id", dto.AuthorID))
		return nil, ErrAuthorNotFound
	}

	// Получение активных членов команды (без учета автора)
	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, author.TeamID, []string{dto.AuthorID})
	if err != nil {
		logger.Error("failed to get team members", zap.Error(err))
		return nil, fmt.Errorf("get team members: %w", err)
	}

	// Выбор до 2 ревьюверов случайным образом
	reviewers := selectRandomReviewers(candidates, 2)

	now := time.Now()
	pr := &PullRequest{
		PullRequestID:     dto.PullRequestID,
		PullRequestName:   dto.PullRequestName,
		AuthorID:          dto.AuthorID,
		Status:            StatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         &now,
	}

	// Создание PR в БД
	if err := s.prRepo.Create(ctx, pr); err != nil {
		logger.Error("failed to create PR", zap.Error(err))
		return nil, fmt.Errorf("create PR: %w", err)
	}

	// Метрики
	metrics.PRCreatedTotal.Inc()
	for _, reviewerID := range reviewers {
		metrics.ReviewerAssignmentsTotal.WithLabelValues(reviewerID).Inc()
	}

	logger.Info("PR created successfully",
		zap.String("pr_id", dto.PullRequestID),
		zap.Int("reviewers_count", len(reviewers)))

	return pr.ToDTO(), nil
}

func (s *service) MergePR(ctx context.Context, prID string) (*DTO, error) {
	logger.Info("merging PR", zap.String("pr_id", prID))

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		logger.Error("failed to get PR", zap.String("pr_id", prID), zap.Error(err))
		return nil, err
	}
	if pr == nil {
		logger.Warn("PR not found", zap.String("pr_id", prID))
		return nil, ErrPRNotFound
	}

	// Идемпотентный merge
	if pr.Status == StatusMerged {
		logger.Debug("PR already merged (idempotent operation)", zap.String("pr_id", prID))
		return pr.ToDTO(), nil
	}

	if err := s.prRepo.Merge(ctx, prID); err != nil {
		logger.Error("failed to merge PR", zap.Error(err))
		return nil, fmt.Errorf("merge PR: %w", err)
	}

	metrics.PRMergedTotal.Inc()

	logger.Info("PR merged successfully", zap.String("pr_id", prID))

	mergedPR, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	return mergedPR.ToDTO(), nil
}

func (s *service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*ReassignResultDTO, error) {
	logger.Info("reassigning reviewer",
		zap.String("pr_id", prID),
		zap.String("old_user_id", oldUserID))

	// Получение PR
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		logger.Error("failed to get PR", zap.Error(err))
		return nil, err
	}
	if pr == nil {
		logger.Warn("PR not found", zap.String("pr_id", prID))
		return nil, ErrPRNotFound
	}

	// Идемпотентный merge
	if pr.Status == StatusMerged {
		logger.Warn("cannot reassign on merged PR", zap.String("pr_id", prID))
		return nil, ErrPRMerged
	}

	// Проверка: oldUserID должен быть в списке ревьюверов
	if !slices.Contains(pr.AssignedReviewers, oldUserID) {
		logger.Warn("user is not assigned as reviewer",
			zap.String("pr_id", prID),
			zap.String("user_id", oldUserID))
		return nil, ErrReviewerNotAssigned
	}

	// Получение информации о заменяемом пользователе
	oldUser, err := s.userRepo.GetByID(ctx, oldUserID)
	if err != nil {
		logger.Error("failed to get old user", zap.Error(err))
		return nil, err
	}
	if oldUser == nil {
		return nil, user.ErrUserNotFound
	}

	// Получение кандидатов из команды заменяемого ревьювера (исключая автора PR и текущих ревьюверов)
	excludeIDs := make([]string, 0, len(pr.AssignedReviewers)+1)
	excludeIDs = append(excludeIDs, pr.AssignedReviewers...)
	excludeIDs = append(excludeIDs, pr.AuthorID)
	candidates, err := s.userRepo.GetActiveTeamMembers(ctx, oldUser.TeamID, excludeIDs)
	if err != nil {
		logger.Error("failed to get candidates", zap.Error(err))
		return nil, fmt.Errorf("get candidates: %w", err)
	}

	if len(candidates) == 0 {
		logger.Warn("no active candidates for reassignment",
			zap.String("pr_id", prID),
			zap.String("old_user_id", oldUserID))
		return nil, ErrNoCandidate
	}

	newReviewerID := selectRandomReviewers(candidates, 1)[0]

	// Транзакционная замена в БД (атомарный UPDATE)
	if err := s.prRepo.ReplaceReviewer(ctx, prID, oldUserID, newReviewerID); err != nil {
		logger.Error("failed to replace reviewer", zap.Error(err))
		return nil, fmt.Errorf("replace reviewer: %w", err)
	}

	metrics.ReassignmentsTotal.Inc()
	metrics.ReviewerAssignmentsTotal.WithLabelValues(newReviewerID).Inc()

	logger.Info("reviewer reassigned successfully",
		zap.String("pr_id", prID),
		zap.String("old_user_id", oldUserID),
		zap.String("new_user_id", newReviewerID))

	updatedPR, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		return nil, err
	}

	return &ReassignResultDTO{
		PR:         updatedPR.ToDTO(),
		ReplacedBy: newReviewerID,
	}, nil
}

func (s *service) GetUserReviews(ctx context.Context, userID string) ([]ShortDTO, error) {
	logger.Debug("getting user reviews", zap.String("user_id", userID))

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		logger.Error("failed to get user reviews", zap.Error(err))
		return nil, fmt.Errorf("get user reviews: %w", err)
	}

	return ShortDTOsToDTO(prs), nil
}

// selectRandomReviewers выбирает случайных ревьюверов из списка кандидатов
// перемешиваем только первые k позиций (O(k)), затем берём первые k элементов.
func selectRandomReviewers(candidates []user.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	n := len(candidates)
	k := min(n, maxCount)

	shuffled := make([]user.User, n)
	copy(shuffled, candidates)

	for i := 0; i < k; i++ {
		j := i + rand.Intn(n-i)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	result := make([]string, k)
	for i := 0; i < k; i++ {
		result[i] = shuffled[i].UserID
	}

	return result
}
