package pullrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	prDomain "github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPRService struct {
	mock.Mock
}

func (m *MockPRService) CreatePR(ctx context.Context, dto *prDomain.CreatePRDTO) (*prDomain.DTO, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*prDomain.DTO), args.Error(1)
}

func (m *MockPRService) MergePR(ctx context.Context, prID string) (*prDomain.DTO, error) {
	args := m.Called(ctx, prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*prDomain.DTO), args.Error(1)
}

func (m *MockPRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*prDomain.ReassignResultDTO, error) {
	args := m.Called(ctx, prID, oldUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*prDomain.ReassignResultDTO), args.Error(1)
}

func (m *MockPRService) GetUserReviews(ctx context.Context, userID string) ([]prDomain.ShortDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]prDomain.ShortDTO), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestCreatePR_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add feature",
		AuthorID:        "u1",
	}

	now := time.Now()
	expectedDTO := &prDomain.DTO{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            prDomain.StatusOpen,
		AssignedReviewers: []string{"u2", "u3"},
		CreatedAt:         &now,
	}

	mockService.On("CreatePR", mock.Anything, mock.AnythingOfType("*pullrequest.CreatePRDTO")).Return(expectedDTO, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreatePR(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response PRResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "pr-1", response.PR.PullRequestID)
	assert.Equal(t, prDomain.StatusOpen, response.PR.Status)
	assert.Len(t, response.PR.AssignedReviewers, 2)

	mockService.AssertExpectations(t)
}

func TestCreatePR_AlreadyExists(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add feature",
		AuthorID:        "u1",
	}

	mockService.On("CreatePR", mock.Anything, mock.AnythingOfType("*pullrequest.CreatePRDTO")).Return(nil, prDomain.ErrPRAlreadyExists)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreatePR(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestCreatePR_ValidationError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := CreatePRRequest{
		PullRequestID:   "",
		PullRequestName: "",
		AuthorID:        "",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreatePR(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMergePR_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := MergePRRequest{
		PullRequestID: "pr-1",
	}

	now := time.Now()
	expectedDTO := &prDomain.DTO{
		PullRequestID:     "pr-1",
		PullRequestName:   "Add feature",
		AuthorID:          "u1",
		Status:            prDomain.StatusMerged,
		AssignedReviewers: []string{"u2"},
		MergedAt:          &now,
	}

	mockService.On("MergePR", mock.Anything, "pr-1").Return(expectedDTO, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.MergePR(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response PRResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, prDomain.StatusMerged, response.PR.Status)
	assert.NotNil(t, response.PR.MergedAt)

	mockService.AssertExpectations(t)
}

func TestMergePR_NotFound(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := MergePRRequest{
		PullRequestID: "nonexistent",
	}

	mockService.On("MergePR", mock.Anything, "nonexistent").Return(nil, prDomain.ErrPRNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.MergePR(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockService.AssertExpectations(t)
}

func TestReassignReviewer_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := ReassignRequest{
		PullRequestID: "pr-1",
		OldUserID:     "u2",
	}

	expectedResult := &prDomain.ReassignResultDTO{
		PR: &prDomain.DTO{
			PullRequestID:     "pr-1",
			PullRequestName:   "Add feature",
			AuthorID:          "u1",
			Status:            prDomain.StatusOpen,
			AssignedReviewers: []string{"u3", "u4"},
		},
		ReplacedBy: "u4",
	}

	mockService.On("ReassignReviewer", mock.Anything, "pr-1", "u2").Return(expectedResult, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response ReassignResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "u4", response.ReplacedBy)
	assert.Contains(t, response.PR.AssignedReviewers, "u4")
	assert.NotContains(t, response.PR.AssignedReviewers, "u2")

	mockService.AssertExpectations(t)
}

func TestReassignReviewer_PRMerged(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := ReassignRequest{
		PullRequestID: "pr-1",
		OldUserID:     "u2",
	}

	mockService.On("ReassignReviewer", mock.Anything, "pr-1", "u2").Return(nil, prDomain.ErrPRMerged)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestReassignReviewer_NotAssigned(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := ReassignRequest{
		PullRequestID: "pr-1",
		OldUserID:     "u2",
	}

	mockService.On("ReassignReviewer", mock.Anything, "pr-1", "u2").Return(nil, prDomain.ErrReviewerNotAssigned)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestReassignReviewer_NoCandidate(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := ReassignRequest{
		PullRequestID: "pr-1",
		OldUserID:     "u2",
	}

	mockService.On("ReassignReviewer", mock.Anything, "pr-1", "u2").Return(nil, prDomain.ErrNoCandidate)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusConflict, rec.Code)

	mockService.AssertExpectations(t)
}

func TestGetUserReviews_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	expectedDTOs := []prDomain.ShortDTO{
		{
			PullRequestID:   "pr-1",
			PullRequestName: "Add feature",
			AuthorID:        "u1",
			Status:          prDomain.StatusOpen,
		},
		{
			PullRequestID:   "pr-2",
			PullRequestName: "Fix bug",
			AuthorID:        "u3",
			Status:          prDomain.StatusMerged,
		},
	}

	mockService.On("GetUserReviews", mock.Anything, "u2").Return(expectedDTOs, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetUserReviews(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserReviewsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "u2", response.UserID)
	assert.Len(t, response.PullRequests, 2)
	assert.Equal(t, "pr-1", response.PullRequests[0].PullRequestID)

	mockService.AssertExpectations(t)
}

func TestGetUserReviews_EmptyList(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	mockService.On("GetUserReviews", mock.Anything, "u2").Return([]prDomain.ShortDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetUserReviews(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserReviewsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "u2", response.UserID)
	assert.Empty(t, response.PullRequests)

	mockService.AssertExpectations(t)
}

func TestGetUserReviews_ValidationError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetUserReviews(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreatePR_InvalidJSON(t *testing.T) {
	logger.InitializeNoop()
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader([]byte("{invalid")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreatePR(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockService.AssertExpectations(t)
}

func TestCreatePR_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := CreatePRRequest{PullRequestID: "pr-1", PullRequestName: "n", AuthorID: "u1"}

	testErr := errors.New("db down")
	mockService.On("CreatePR", mock.Anything, mock.AnythingOfType("*pullrequest.CreatePRDTO")).Return(nil, testErr)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreatePR(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMergePR_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader([]byte("not json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.MergePR(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMergePR_ValidationError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := MergePRRequest{PullRequestID: ""}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.MergePR(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestMergePR_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := MergePRRequest{PullRequestID: "pr-1"}

	testErr := errors.New("merge failed")
	mockService.On("MergePR", mock.Anything, "pr-1").Return(nil, testErr)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.MergePR(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestReassignReviewer_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader([]byte("oops")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReassignReviewer_ValidationError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := ReassignRequest{PullRequestID: "", OldUserID: ""}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReassignReviewer_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	reqBody := ReassignRequest{PullRequestID: "pr-1", OldUserID: "u2"}

	testErr := errors.New("replace failed")
	mockService.On("ReassignReviewer", mock.Anything, "pr-1", "u2").Return(nil, testErr)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ReassignReviewer(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetUserReviews_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockPRService)
	handler := NewHandler(mockService)

	testErr := errors.New("repo failed")
	mockService.On("GetUserReviews", mock.Anything, "u2").Return(nil, testErr)

	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetUserReviews(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestCreatePRRequest_ToDomainDTO(t *testing.T) {
	req := &CreatePRRequest{PullRequestID: "pr-1", PullRequestName: "feat", AuthorID: "u1"}
	dto := req.ToDomainDTO()
	assert.Equal(t, "pr-1", dto.PullRequestID)
	assert.Equal(t, "feat", dto.PullRequestName)
	assert.Equal(t, "u1", dto.AuthorID)
}

func TestFromDomainDTO_PR(t *testing.T) {
	now := time.Now()
	merged := now
	d := &prDomain.DTO{
		PullRequestID:     "pr-1",
		PullRequestName:   "feat",
		AuthorID:          "u1",
		Status:            prDomain.StatusMerged,
		AssignedReviewers: []string{"u2"},
		CreatedAt:         &now,
		MergedAt:          &merged,
	}

	resp := FromDomainDTO(d)
	assert.Equal(t, "pr-1", resp.PR.PullRequestID)
	assert.Equal(t, prDomain.StatusMerged, resp.PR.Status)
	if resp.PR.CreatedAt == nil || resp.PR.MergedAt == nil {
		t.Fatalf("expected timestamps to be set")
	}
}

func TestFromReassignResultDTO(t *testing.T) {
	d := &prDomain.ReassignResultDTO{PR: &prDomain.DTO{PullRequestID: "pr-1"}, ReplacedBy: "u9"}
	resp := FromReassignResultDTO(d)
	assert.Equal(t, "u9", resp.ReplacedBy)
	assert.Equal(t, "pr-1", resp.PR.PullRequestID)
}

func TestFromShortDTOs(t *testing.T) {
	shorts := []prDomain.ShortDTO{{PullRequestID: "pr-1", PullRequestName: "n", AuthorID: "a", Status: prDomain.StatusOpen}}
	views := FromShortDTOs(shorts)
	assert.Len(t, views, 1)
	assert.Equal(t, "pr-1", views[0].PullRequestID)
}
