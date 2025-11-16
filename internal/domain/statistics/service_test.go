package statistics

import (
	"context"
	"errors"
	"testing"

	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStatisticsRepository struct {
	mock.Mock
}

func (m *MockStatisticsRepository) GetUserStats(ctx context.Context, userID string) (*UserStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserStats), args.Error(1)
}

func (m *MockStatisticsRepository) GetTeamStats(ctx context.Context, teamName string) (*TeamStats, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TeamStats), args.Error(1)
}

func (m *MockStatisticsRepository) GetGlobalStats(ctx context.Context) (*GlobalStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GlobalStats), args.Error(1)
}

func (m *MockStatisticsRepository) GetTopReviewers(ctx context.Context, limit int) ([]TopReviewer, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]TopReviewer), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestGetUserStats_Success(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)

	ctx := context.Background()
	expectedStats := &UserStats{
		UserID:           "u1",
		Username:         "Alice",
		TeamName:         "backend",
		TotalAssignments: 10,
		ActiveReviews:    3,
		CompletedReviews: 7,
	}

	repo.On("GetUserStats", ctx, "u1").Return(expectedStats, nil)

	result, err := service.GetUserStats(ctx, "u1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "u1", result.UserID)
	assert.Equal(t, 10, result.TotalAssignments)

	repo.AssertExpectations(t)
}

func TestGetGlobalStats_Success(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)

	ctx := context.Background()
	expectedStats := &GlobalStats{
		TotalUsers:       50,
		ActiveUsers:      45,
		TotalTeams:       5,
		TotalPRs:         100,
		OpenPRs:          20,
		MergedPRs:        80,
		TotalAssignments: 150,
	}

	repo.On("GetGlobalStats", ctx).Return(expectedStats, nil)

	result, err := service.GetGlobalStats(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 50, result.TotalUsers)
	assert.Equal(t, 100, result.TotalPRs)

	repo.AssertExpectations(t)
}

func TestGetTopReviewers_Success(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)

	ctx := context.Background()
	expectedReviewers := []TopReviewer{
		{UserID: "u1", Username: "Alice", TeamName: "backend", ReviewCount: 20},
		{UserID: "u2", Username: "Bob", TeamName: "backend", ReviewCount: 15},
	}

	repo.On("GetTopReviewers", ctx, 10).Return(expectedReviewers, nil)

	result, err := service.GetTopReviewers(ctx, 10)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "u1", result[0].UserID)
	assert.Equal(t, 20, result[0].ReviewCount)

	repo.AssertExpectations(t)
}

func TestGetTopReviewers_LimitValidation(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)

	ctx := context.Background()

	// Limit <= 0 должен стать 10
	repo.On("GetTopReviewers", ctx, 10).Return([]TopReviewer{}, nil).Once()
	_, err := service.GetTopReviewers(ctx, 0)
	assert.NoError(t, err)

	// Limit > 200 должен стать 200
	repo.On("GetTopReviewers", ctx, 200).Return([]TopReviewer{}, nil).Once()
	_, err = service.GetTopReviewers(ctx, 300)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestStats_ToDTOs(t *testing.T) {
	us := &UserStats{UserID: "u1", Username: "A", TeamName: "backend", TotalAssignments: 3, ActiveReviews: 1, CompletedReviews: 2}
	usd := us.ToDTO()
	if usd.UserID != "u1" || usd.TotalAssignments != 3 || usd.CompletedReviews != 2 {
		t.Fatalf("bad usd: %+v", usd)
	}

	ts := &TeamStats{TeamName: "backend", TotalMembers: 5, ActiveMembers: 4, TotalPRs: 10, OpenPRs: 3, MergedPRs: 7}
	tsd := ts.ToDTO()
	if tsd.TeamName != "backend" || tsd.TotalPRs != 10 || tsd.MergedPRs != 7 {
		t.Fatalf("bad tsd: %+v", tsd)
	}

	gs := &GlobalStats{TotalUsers: 10, ActiveUsers: 8, TotalTeams: 2, TotalPRs: 15, OpenPRs: 5, MergedPRs: 10, TotalAssignments: 20}
	gsd := gs.ToDTO()
	if gsd.TotalUsers != 10 || gsd.TotalAssignments != 20 {
		t.Fatalf("bad gsd: %+v", gsd)
	}

	trs := []TopReviewer{{UserID: "u1", Username: "A", TeamName: "backend", ReviewCount: 7}}
	trds := TopReviewersToDTO(trs)
	if len(trds) != 1 || trds[0].ReviewCount != 7 {
		t.Fatalf("bad trds: %+v", trds)
	}
}

func TestGetUserStats_RepoError(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetUserStats", ctx, "u1").Return(nil, errors.New("db"))
	res, err := service.GetUserStats(ctx, "u1")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetTeamStats_RepoError(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetTeamStats", ctx, "backend").Return(nil, errors.New("db"))
	res, err := service.GetTeamStats(ctx, "backend")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetGlobalStats_RepoError(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetGlobalStats", ctx).Return(nil, errors.New("db"))
	res, err := service.GetGlobalStats(ctx)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetTopReviewers_RepoError(t *testing.T) {
	repo := new(MockStatisticsRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetTopReviewers", ctx, 10).Return(nil, errors.New("db"))
	res, err := service.GetTopReviewers(ctx, 10)
	assert.Nil(t, res)
	assert.Error(t, err)
}
