package team

import (
	"context"
	"errors"
	"testing"

	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, team *Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByName(ctx context.Context, teamName string) (*Team, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Team), args.Error(1)
}

func (m *MockTeamRepository) Exists(ctx context.Context, teamName string) (bool, error) {
	args := m.Called(ctx, teamName)
	return args.Bool(0), args.Error(1)
}

func (m *MockTeamRepository) GetIDByName(ctx context.Context, teamName string) (int, error) {
	args := m.Called(ctx, teamName)
	return args.Int(0), args.Error(1)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Upsert(ctx context.Context, user *user.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*user.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

func (m *MockUserRepository) GetActiveTeamMembers(ctx context.Context, teamID int, excludeUserIDs []string) ([]user.User, error) {
	args := m.Called(ctx, teamID, excludeUserIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestCreateTeam_Success(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)

	ctx := context.Background()
	dto := &CreateTeamDTO{
		TeamName: "backend",
		Members: []MemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	teamRepo.On("Exists", ctx, "backend").Return(false, nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*team.Team")).Return(nil)
	teamRepo.On("GetIDByName", ctx, "backend").Return(1, nil)

	userRepo.On("Upsert", ctx, mock.AnythingOfType("*user.User")).Return(nil).Twice()

	teamRepo.On("GetByName", ctx, "backend").Return(&Team{
		ID:       1,
		TeamName: "backend",
		Members: []Member{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}, nil)

	result, err := service.CreateTeam(ctx, dto)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "backend", result.TeamName)
	assert.Len(t, result.Members, 2)
	assert.Equal(t, "u1", result.Members[0].UserID)
	assert.Equal(t, "Alice", result.Members[0].Username)

	teamRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)

	ctx := context.Background()
	dto := &CreateTeamDTO{
		TeamName: "backend",
		Members:  []MemberDTO{{UserID: "u1", Username: "Alice", IsActive: true}},
	}

	teamRepo.On("Exists", ctx, "backend").Return(true, nil)

	result, err := service.CreateTeam(ctx, dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTeamAlreadyExists)

	teamRepo.AssertExpectations(t)
}

func TestCreateTeam_EmptyMembers(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)

	ctx := context.Background()
	dto := &CreateTeamDTO{
		TeamName: "backend",
		Members:  []MemberDTO{},
	}

	teamRepo.On("Exists", ctx, "backend").Return(false, nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*team.Team")).Return(nil)
	teamRepo.On("GetIDByName", ctx, "backend").Return(1, nil)
	teamRepo.On("GetByName", ctx, "backend").Return(&Team{
		ID:       1,
		TeamName: "backend",
		Members:  []Member{},
	}, nil)

	result, err := service.CreateTeam(ctx, dto)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Members, 0)

	teamRepo.AssertExpectations(t)
}

func TestGetTeam_Success(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)

	ctx := context.Background()
	expectedTeam := &Team{
		ID:       1,
		TeamName: "backend",
		Members: []Member{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	teamRepo.On("GetByName", ctx, "backend").Return(expectedTeam, nil)

	result, err := service.GetTeam(ctx, "backend")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "backend", result.TeamName)
	assert.Len(t, result.Members, 1)
	assert.Equal(t, "u1", result.Members[0].UserID)

	teamRepo.AssertExpectations(t)
}

func TestGetTeam_NotFound(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)

	ctx := context.Background()

	teamRepo.On("GetByName", ctx, "nonexistent").Return(nil, nil)

	result, err := service.GetTeam(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrTeamNotFound)

	teamRepo.AssertExpectations(t)
}

func TestTeam_ToDTO_Roundtrip(t *testing.T) {
	t1 := &Team{TeamName: "backend", Members: []Member{{UserID: "u1", Username: "A", IsActive: true}}}
	dto := t1.ToDTO()
	if dto.TeamName != "backend" || len(dto.Members) != 1 || dto.Members[0].UserID != "u1" {
		t.Fatalf("unexpected dto: %+v", dto)
	}

	model := FromCreateDTO(&CreateTeamDTO{TeamName: dto.TeamName, Members: dto.Members})
	if model.TeamName != "backend" || len(model.Members) != 1 || model.Members[0].UserID != "u1" {
		t.Fatalf("unexpected model: %+v", model)
	}
}

func TestCreateTeam_ExistsError(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)
	ctx := context.Background()
	dto := &CreateTeamDTO{TeamName: "backend", Members: []MemberDTO{}}
	teamRepo.On("Exists", ctx, "backend").Return(false, errors.New("db"))
	res, err := service.CreateTeam(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreateTeam_CreateError(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)
	ctx := context.Background()
	dto := &CreateTeamDTO{TeamName: "backend", Members: []MemberDTO{{UserID: "u1", Username: "A", IsActive: true}}}
	teamRepo.On("Exists", ctx, "backend").Return(false, nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*team.Team")).Return(errors.New("db"))
	res, err := service.CreateTeam(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreateTeam_GetIDError(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)
	ctx := context.Background()
	dto := &CreateTeamDTO{TeamName: "backend", Members: []MemberDTO{}}
	teamRepo.On("Exists", ctx, "backend").Return(false, nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*team.Team")).Return(nil)
	teamRepo.On("GetIDByName", ctx, "backend").Return(0, errors.New("db"))
	res, err := service.CreateTeam(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreateTeam_UpsertError(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)
	ctx := context.Background()
	dto := &CreateTeamDTO{TeamName: "backend", Members: []MemberDTO{{UserID: "u1", Username: "A", IsActive: true}}}
	teamRepo.On("Exists", ctx, "backend").Return(false, nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*team.Team")).Return(nil)
	teamRepo.On("GetIDByName", ctx, "backend").Return(1, nil)
	userRepo.On("Upsert", ctx, mock.AnythingOfType("*user.User")).Return(errors.New("db"))
	res, err := service.CreateTeam(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreateTeam_GetByNameError(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)
	ctx := context.Background()
	dto := &CreateTeamDTO{TeamName: "backend", Members: []MemberDTO{}}
	teamRepo.On("Exists", ctx, "backend").Return(false, nil)
	teamRepo.On("Create", ctx, mock.AnythingOfType("*team.Team")).Return(nil)
	teamRepo.On("GetIDByName", ctx, "backend").Return(1, nil)
	teamRepo.On("GetByName", ctx, "backend").Return(nil, errors.New("db"))
	res, err := service.CreateTeam(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetTeam_RepoError(t *testing.T) {
	teamRepo := new(MockTeamRepository)
	userRepo := new(MockUserRepository)
	service := NewService(teamRepo, userRepo)
	ctx := context.Background()
	teamRepo.On("GetByName", ctx, "backend").Return(nil, errors.New("db"))
	res, err := service.GetTeam(ctx, "backend")
	assert.Nil(t, res)
	assert.Error(t, err)
}
