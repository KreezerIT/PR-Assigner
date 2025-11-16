package pullrequest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPRRepository struct {
	mock.Mock
}

func (m *MockPRRepository) Create(ctx context.Context, pr *PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *MockPRRepository) GetByID(ctx context.Context, prID string) (*PullRequest, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*PullRequest), args.Error(1)
}

func (m *MockPRRepository) Exists(ctx context.Context, prID string) (bool, error) {
	args := m.Called(ctx, prID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPRRepository) Merge(ctx context.Context, prID string) error {
	args := m.Called(ctx, prID)
	return args.Error(0)
}

func (m *MockPRRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPRRepository) RemoveReviewer(ctx context.Context, prID string, userID string) error {
	args := m.Called(ctx, prID, userID)
	return args.Error(0)
}

func (m *MockPRRepository) AddReviewer(ctx context.Context, prID string, userID string) error {
	args := m.Called(ctx, prID, userID)
	return args.Error(0)
}

func (m *MockPRRepository) GetByReviewer(ctx context.Context, userID string) ([]PullRequestShort, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]PullRequestShort), args.Error(1)
}

func (m *MockPRRepository) ReplaceReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	args := m.Called(ctx, prID, oldUserID, newUserID)
	return args.Error(0)
}

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Upsert(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
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

func TestCreatePR_Success(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()
	dto := &CreatePRDTO{
		PullRequestID:   "pr-1",
		PullRequestName: "Add feature",
		AuthorID:        "u1",
	}

	author := &user.User{
		UserID:   "u1",
		Username: "Alice",
		TeamID:   1,
		IsActive: true,
	}

	candidates := []user.User{
		{UserID: "u2", Username: "Bob", TeamID: 1, IsActive: true},
		{UserID: "u3", Username: "Charlie", TeamID: 1, IsActive: true},
	}

	prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	userRepo.On("GetByID", ctx, "u1").Return(author, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u1"}).Return(candidates, nil)
	prRepo.On("Create", ctx, mock.AnythingOfType("*pullrequest.PullRequest")).Return(nil)

	result, err := service.CreatePR(ctx, dto)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "pr-1", result.PullRequestID)
	assert.Equal(t, "Add feature", result.PullRequestName)
	assert.Equal(t, "u1", result.AuthorID)
	assert.Equal(t, StatusOpen, result.Status)
	assert.LessOrEqual(t, len(result.AssignedReviewers), 2)
	assert.NotNil(t, result.CreatedAt)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreatePR_AlreadyExists(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()
	dto := &CreatePRDTO{
		PullRequestID:   "pr-1",
		PullRequestName: "Add feature",
		AuthorID:        "u1",
	}

	prRepo.On("Exists", ctx, "pr-1").Return(true, nil)

	result, err := service.CreatePR(ctx, dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrPRAlreadyExists)

	prRepo.AssertExpectations(t)
}

func TestCreatePR_AuthorNotFound(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()
	dto := &CreatePRDTO{
		PullRequestID:   "pr-1",
		PullRequestName: "Add feature",
		AuthorID:        "nonexistent",
	}

	prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	userRepo.On("GetByID", ctx, "nonexistent").Return(nil, nil)

	result, err := service.CreatePR(ctx, dto)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrAuthorNotFound)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestCreatePR_NoCandidates(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()
	dto := &CreatePRDTO{
		PullRequestID:   "pr-1",
		PullRequestName: "Add feature",
		AuthorID:        "u1",
	}

	author := &user.User{
		UserID:   "u1",
		Username: "Alice",
		TeamID:   1,
		IsActive: true,
	}

	prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	userRepo.On("GetByID", ctx, "u1").Return(author, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u1"}).Return([]user.User{}, nil)
	prRepo.On("Create", ctx, mock.AnythingOfType("*pullrequest.PullRequest")).Return(nil)

	result, err := service.CreatePR(ctx, dto)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.AssignedReviewers)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestMergePR_Success(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	pr := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	now := time.Now()
	mergedPR := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusMerged,
		AssignedReviewers: []string{"u2"},
		MergedAt:          &now,
	}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	prRepo.On("Merge", ctx, "pr-1").Return(nil)
	prRepo.On("GetByID", ctx, "pr-1").Return(mergedPR, nil).Once()

	result, err := service.MergePR(ctx, "pr-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusMerged, result.Status)
	assert.NotNil(t, result.MergedAt)

	prRepo.AssertExpectations(t)
}

func TestMergePR_Idempotent(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	now := time.Now()
	mergedPR := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusMerged,
		AssignedReviewers: []string{"u2"},
		MergedAt:          &now,
	}

	prRepo.On("GetByID", ctx, "pr-1").Return(mergedPR, nil)

	result, err := service.MergePR(ctx, "pr-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, StatusMerged, result.Status)

	prRepo.AssertExpectations(t)
	prRepo.AssertNotCalled(t, "Merge")
}

func TestMergePR_NotFound(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	prRepo.On("GetByID", ctx, "nonexistent").Return(nil, nil)

	result, err := service.MergePR(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrPRNotFound)

	prRepo.AssertExpectations(t)
}

func TestReassignReviewer_Success(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	pr := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
	}

	oldUser := &user.User{
		UserID:   "u2",
		Username: "Bob",
		TeamID:   1,
		IsActive: true,
	}

	candidates := []user.User{
		{UserID: "u4", Username: "Dave", TeamID: 1, IsActive: true},
	}

	updatedPR := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusOpen,
		AssignedReviewers: []string{"u3", "u4"},
	}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	userRepo.On("GetByID", ctx, "u2").Return(oldUser, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u2", "u3", "u1"}).Return(candidates, nil)
	prRepo.On("ReplaceReviewer", ctx, "pr-1", "u2", "u4").Return(nil)
	prRepo.On("GetByID", ctx, "pr-1").Return(updatedPR, nil).Once()

	result, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.PR)
	assert.Equal(t, "u4", result.ReplacedBy)
	assert.Contains(t, result.PR.AssignedReviewers, "u4")
	assert.NotContains(t, result.PR.AssignedReviewers, "u2")

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestReassignReviewer_PRMerged(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	pr := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusMerged,
		AssignedReviewers: []string{"u2"},
	}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

	result, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrPRMerged)

	prRepo.AssertExpectations(t)
}

func TestReassignReviewer_NotAssigned(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	pr := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusOpen,
		AssignedReviewers: []string{"u3"},
	}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)

	result, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrReviewerNotAssigned)

	prRepo.AssertExpectations(t)
}

func TestReassignReviewer_NoCandidate(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	pr := &PullRequest{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            StatusOpen,
		AssignedReviewers: []string{"u2"},
	}

	oldUser := &user.User{
		UserID:   "u2",
		Username: "Bob",
		TeamID:   1,
		IsActive: true,
	}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
	userRepo.On("GetByID", ctx, "u2").Return(oldUser, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u2", "u1"}).Return([]user.User{}, nil)

	result, err := service.ReassignReviewer(ctx, "pr-1", "u2")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNoCandidate)

	prRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestGetUserReviews_Success(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	prs := []PullRequestShort{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Add feature",
			AuthorID:        "u1",
			Status:          StatusOpen,
		},
		{
			PullRequestID:   "pr-2",
			PullRequestName: "Fix bug",
			AuthorID:        "u3",
			Status:          StatusMerged,
		},
	}

	prRepo.On("GetByReviewer", ctx, "u2").Return(prs, nil)

	result, err := service.GetUserReviews(ctx, "u2")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "pr-1", result[0].PullRequestID)
	assert.Equal(t, "pr-2", result[1].PullRequestID)

	prRepo.AssertExpectations(t)
}

func TestGetUserReviews_EmptyList(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)

	ctx := context.Background()

	prRepo.On("GetByReviewer", ctx, "u2").Return([]PullRequestShort{}, nil)

	result, err := service.GetUserReviews(ctx, "u2")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)

	prRepo.AssertExpectations(t)
}

func TestCreatePR_RepoExistsError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	dto := &CreatePRDTO{PullRequestID: "pr-1", PullRequestName: "n", AuthorID: "u1"}

	prRepo.On("Exists", ctx, "pr-1").Return(false, errors.New("db"))

	res, err := service.CreatePR(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreatePR_GetAuthorError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	dto := &CreatePRDTO{PullRequestID: "pr-1", PullRequestName: "n", AuthorID: "u1"}

	prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	userRepo.On("GetByID", ctx, "u1").Return(nil, errors.New("db"))

	res, err := service.CreatePR(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreatePR_GetMembersError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	dto := &CreatePRDTO{PullRequestID: "pr-1", PullRequestName: "n", AuthorID: "u1"}
	author := &user.User{UserID: "u1", TeamID: 1, IsActive: true}

	prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	userRepo.On("GetByID", ctx, "u1").Return(author, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u1"}).Return(nil, errors.New("db"))

	res, err := service.CreatePR(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestCreatePR_CreateError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	dto := &CreatePRDTO{PullRequestID: "pr-1", PullRequestName: "n", AuthorID: "u1"}
	author := &user.User{UserID: "u1", TeamID: 1, IsActive: true}

	prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
	userRepo.On("GetByID", ctx, "u1").Return(author, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u1"}).Return([]user.User{}, nil)
	prRepo.On("Create", ctx, mock.AnythingOfType("*pullrequest.PullRequest")).Return(errors.New("db"))

	res, err := service.CreatePR(ctx, dto)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestMergePR_GetByIDError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	prRepo.On("GetByID", ctx, "pr-1").Return(nil, errors.New("db"))

	res, err := service.MergePR(ctx, "pr-1")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestMergePR_MergeError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen}
	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	prRepo.On("Merge", ctx, "pr-1").Return(errors.New("db"))

	res, err := service.MergePR(ctx, "pr-1")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestMergePR_FinalGetError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen}
	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	prRepo.On("Merge", ctx, "pr-1").Return(nil)
	prRepo.On("GetByID", ctx, "pr-1").Return(nil, errors.New("db")).Once()

	res, err := service.MergePR(ctx, "pr-1")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestReassignReviewer_GetPR_Error(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	prRepo.On("GetByID", ctx, "pr-1").Return(nil, errors.New("db"))

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestReassignReviewer_PRNotFound(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	prRepo.On("GetByID", ctx, "pr-1").Return(nil, nil)

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrPRNotFound)
}

func TestReassignReviewer_GetOldUserError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen, AssignedReviewers: []string{"u2"}}
	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
	userRepo.On("GetByID", ctx, "u2").Return(nil, errors.New("db"))

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestReassignReviewer_OldUserNotFound(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen, AssignedReviewers: []string{"u2"}}
	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil)
	userRepo.On("GetByID", ctx, "u2").Return(nil, nil)

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestReassignReviewer_GetCandidatesError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}
	oldUser := &user.User{UserID: "u2", TeamID: 1}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	userRepo.On("GetByID", ctx, "u2").Return(oldUser, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u2", "u1"}).Return(nil, errors.New("db"))

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestReassignReviewer_ReplaceError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}
	oldUser := &user.User{UserID: "u2", TeamID: 1}
	candidates := []user.User{{UserID: "u4", TeamID: 1}}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	userRepo.On("GetByID", ctx, "u2").Return(oldUser, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u2", "u1"}).Return(candidates, nil)
	prRepo.On("ReplaceReviewer", ctx, "pr-1", "u2", mock.AnythingOfType("string")).Return(errors.New("db"))

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestReassignReviewer_FinalGetError(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	pr := &PullRequest{PullRequestID: "pr-1", Status: StatusOpen, AuthorID: "u1", AssignedReviewers: []string{"u2"}}
	oldUser := &user.User{UserID: "u2", TeamID: 1}
	candidates := []user.User{{UserID: "u4", TeamID: 1}}

	prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
	userRepo.On("GetByID", ctx, "u2").Return(oldUser, nil)
	userRepo.On("GetActiveTeamMembers", ctx, 1, []string{"u2", "u1"}).Return(candidates, nil)
	prRepo.On("ReplaceReviewer", ctx, "pr-1", "u2", mock.AnythingOfType("string")).Return(nil)
	prRepo.On("GetByID", ctx, "pr-1").Return(nil, errors.New("db")).Once()

	res, err := service.ReassignReviewer(ctx, "pr-1", "u2")
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetUserReviews_Error(t *testing.T) {
	prRepo := new(MockPRRepository)
	userRepo := new(MockUserRepository)
	service := NewService(prRepo, userRepo)
	ctx := context.Background()

	prRepo.On("GetByReviewer", ctx, "u9").Return(nil, errors.New("db"))

	res, err := service.GetUserReviews(ctx, "u9")
	assert.Nil(t, res)
	assert.Error(t, err)
}
