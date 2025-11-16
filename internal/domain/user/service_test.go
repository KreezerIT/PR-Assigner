package user

import (
	"context"
	"errors"
	"testing"

	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Upsert(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

func (m *MockUserRepository) GetActiveTeamMembers(ctx context.Context, teamID int, excludeUserIDs []string) ([]User, error) {
	args := m.Called(ctx, teamID, excludeUserIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestSetIsActive_Success(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)

	ctx := context.Background()
	userID := "u1"

	existingUser := &User{
		UserID:   "u1",
		Username: "Alice",
		TeamID:   1,
		TeamName: "backend",
		IsActive: true,
	}

	updatedUser := &User{
		UserID:   "u1",
		Username: "Alice",
		TeamID:   1,
		TeamName: "backend",
		IsActive: false,
	}

	repo.On("GetByID", ctx, userID).Return(existingUser, nil).Once()
	repo.On("SetIsActive", ctx, userID, false).Return(nil)
	repo.On("GetByID", ctx, userID).Return(updatedUser, nil).Once()

	result, err := service.SetIsActive(ctx, userID, false)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsActive)
	assert.Equal(t, "u1", result.UserID)
	assert.Equal(t, "Alice", result.Username)
	assert.Equal(t, "backend", result.TeamName)

	repo.AssertExpectations(t)
}

func TestSetIsActive_UserNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)

	ctx := context.Background()
	userID := "nonexistent"

	repo.On("GetByID", ctx, userID).Return(nil, nil)

	result, err := service.SetIsActive(ctx, userID, false)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrUserNotFound)

	repo.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)

	ctx := context.Background()
	expectedUser := &User{
		UserID:   "u1",
		Username: "Alice",
		TeamID:   1,
		TeamName: "backend",
		IsActive: true,
	}

	repo.On("GetByID", ctx, "u1").Return(expectedUser, nil)

	result, err := service.GetUser(ctx, "u1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "u1", result.UserID)
	assert.Equal(t, "Alice", result.Username)
	assert.Equal(t, "backend", result.TeamName)
	assert.True(t, result.IsActive)

	repo.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)

	ctx := context.Background()

	repo.On("GetByID", ctx, "nonexistent").Return(nil, nil)

	result, err := service.GetUser(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrUserNotFound)

	repo.AssertExpectations(t)
}

func TestUser_ToDTO(t *testing.T) {
	u := &User{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true}
	dto := u.ToDTO()
	if dto.UserID != "u1" || dto.Username != "Alice" || dto.TeamName != "backend" || !dto.IsActive {
		t.Fatalf("unexpected dto: %+v", dto)
	}
}

func TestSetIsActive_GetByIDError(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetByID", ctx, "u1").Return(nil, errors.New("db")).Once()
	res, err := service.SetIsActive(ctx, "u1", true)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestSetIsActive_SetError(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetByID", ctx, "u1").Return(&User{UserID: "u1"}, nil).Once()
	repo.On("SetIsActive", ctx, "u1", false).Return(errors.New("db"))
	res, err := service.SetIsActive(ctx, "u1", false)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestSetIsActive_FinalGetError(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetByID", ctx, "u1").Return(&User{UserID: "u1"}, nil).Once()
	repo.On("SetIsActive", ctx, "u1", true).Return(nil)
	repo.On("GetByID", ctx, "u1").Return(nil, errors.New("db")).Once()
	res, err := service.SetIsActive(ctx, "u1", true)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetUser_RepoError(t *testing.T) {
	repo := new(MockUserRepository)
	service := NewService(repo)
	ctx := context.Background()
	repo.On("GetByID", ctx, "u1").Return(nil, errors.New("db"))
	res, err := service.GetUser(ctx, "u1")
	assert.Nil(t, res)
	assert.Error(t, err)
}
