package statistics

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	statsDomain "github.com/kreezerit/pr-assigner/internal/domain/statistics"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStatisticsService struct {
	mock.Mock
}

func (m *MockStatisticsService) GetUserStats(ctx context.Context, userID string) (*statsDomain.UserStatsDTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*statsDomain.UserStatsDTO), args.Error(1)
}

func (m *MockStatisticsService) GetTeamStats(ctx context.Context, teamName string) (*statsDomain.TeamStatsDTO, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*statsDomain.TeamStatsDTO), args.Error(1)
}

func (m *MockStatisticsService) GetGlobalStats(ctx context.Context) (*statsDomain.GlobalStatsDTO, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*statsDomain.GlobalStatsDTO), args.Error(1)
}

func (m *MockStatisticsService) GetTopReviewers(ctx context.Context, limit int) ([]statsDomain.TopReviewerDTO, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]statsDomain.TopReviewerDTO), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestGetGlobalStats_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	expectedDTO := &statsDomain.GlobalStatsDTO{
		TotalUsers:       50,
		ActiveUsers:      45,
		TotalTeams:       5,
		TotalPRs:         100,
		OpenPRs:          20,
		MergedPRs:        80,
		TotalAssignments: 150,
	}

	mockService.On("GetGlobalStats", mock.Anything).Return(expectedDTO, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/global", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetGlobalStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response GlobalStatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 50, response.TotalUsers)
	assert.Equal(t, 100, response.TotalPRs)
	assert.Equal(t, 150, response.TotalAssignments)

	mockService.AssertExpectations(t)
}

func TestGetUserStats_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	expectedDTO := &statsDomain.UserStatsDTO{
		UserID:           "u1",
		Username:         "Alice",
		TeamName:         "backend",
		TotalAssignments: 10,
		ActiveReviews:    3,
		CompletedReviews: 7,
	}

	mockService.On("GetUserStats", mock.Anything, "u1").Return(expectedDTO, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/user/u1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("user_id")
	c.SetParamValues("u1")

	err := handler.GetUserStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserStatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "u1", response.UserID)
	assert.Equal(t, "Alice", response.Username)
	assert.Equal(t, 10, response.TotalAssignments)
	assert.Equal(t, 3, response.ActiveReviews)
	assert.Equal(t, 7, response.CompletedReviews)

	mockService.AssertExpectations(t)
}

func TestGetUserStats_EmptyUserID(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/statistics/user/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("user_id")
	c.SetParamValues("")

	err := handler.GetUserStats(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTeamStats_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	expectedDTO := &statsDomain.TeamStatsDTO{
		TeamName:      "backend",
		TotalMembers:  10,
		ActiveMembers: 8,
		TotalPRs:      50,
		OpenPRs:       10,
		MergedPRs:     40,
	}

	mockService.On("GetTeamStats", mock.Anything, "backend").Return(expectedDTO, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/team/backend", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("team_name")
	c.SetParamValues("backend")

	err := handler.GetTeamStats(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response TeamStatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "backend", response.TeamName)
	assert.Equal(t, 10, response.TotalMembers)
	assert.Equal(t, 50, response.TotalPRs)

	mockService.AssertExpectations(t)
}

func TestGetTeamStats_EmptyTeamName(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/statistics/team/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("team_name")
	c.SetParamValues("")

	err := handler.GetTeamStats(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTopReviewers_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	expectedDTOs := []statsDomain.TopReviewerDTO{
		{UserID: "u1", Username: "Alice", TeamName: "backend", ReviewCount: 20},
		{UserID: "u2", Username: "Bob", TeamName: "backend", ReviewCount: 15},
		{UserID: "u3", Username: "Charlie", TeamName: "frontend", ReviewCount: 10},
	}

	mockService.On("GetTopReviewers", mock.Anything, 10).Return(expectedDTOs, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers?limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTopReviewers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response TopReviewersResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.TopReviewers, 3)
	assert.Equal(t, "u1", response.TopReviewers[0].UserID)
	assert.Equal(t, 20, response.TopReviewers[0].ReviewCount)

	mockService.AssertExpectations(t)
}

func TestGetTopReviewers_DefaultLimit(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	mockService.On("GetTopReviewers", mock.Anything, 10).Return([]statsDomain.TopReviewerDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTopReviewers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

func TestGetTopReviewers_CustomLimit(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	mockService.On("GetTopReviewers", mock.Anything, 5).Return([]statsDomain.TopReviewerDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers?limit=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTopReviewers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

func TestGetTopReviewers_InvalidLimit(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	// Невалидный лимит должен использовать дефолтное значение
	mockService.On("GetTopReviewers", mock.Anything, 10).Return([]statsDomain.TopReviewerDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers?limit=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTopReviewers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

func TestGetTopReviewers_EmptyResult(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	mockService.On("GetTopReviewers", mock.Anything, 10).Return([]statsDomain.TopReviewerDTO{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTopReviewers(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response TopReviewersResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response.TopReviewers)

	mockService.AssertExpectations(t)
}

func TestGetGlobalStats_ServiceError(t *testing.T) {
	logger.InitializeNoop()
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	testErr := errors.New("repo failed")
	mockService.On("GetGlobalStats", mock.Anything).Return(nil, testErr)

	req := httptest.NewRequest(http.MethodGet, "/statistics/global", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetGlobalStats(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetUserStats_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	testErr := errors.New("repo failed")
	mockService.On("GetUserStats", mock.Anything, "u1").Return(nil, testErr)

	req := httptest.NewRequest(http.MethodGet, "/statistics/user/u1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("user_id")
	c.SetParamValues("u1")

	err := handler.GetUserStats(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTeamStats_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	testErr := errors.New("repo failed")
	mockService.On("GetTeamStats", mock.Anything, "backend").Return(nil, testErr)

	req := httptest.NewRequest(http.MethodGet, "/statistics/team/backend", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("team_name")
	c.SetParamValues("backend")

	err := handler.GetTeamStats(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTopReviewers_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockStatisticsService)
	handler := NewHandler(mockService)

	testErr := errors.New("repo failed")
	mockService.On("GetTopReviewers", mock.Anything, 10).Return(nil, testErr)

	req := httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTopReviewers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
