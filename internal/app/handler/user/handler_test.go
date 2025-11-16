package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userDomain "github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*userDomain.DTO, error) {
	args := m.Called(ctx, userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.DTO), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*userDomain.DTO, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.DTO), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestSetIsActive_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	reqBody := SetIsActiveRequest{
		UserID:   "u1",
		IsActive: false,
	}

	expectedDTO := &userDomain.DTO{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: false,
	}

	mockService.On("SetIsActive", mock.Anything, "u1", false).Return(expectedDTO, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "u1", response.User.UserID)
	assert.False(t, response.User.IsActive)

	mockService.AssertExpectations(t)
}

func TestSetIsActive_UserNotFound(t *testing.T) {
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	reqBody := SetIsActiveRequest{
		UserID:   "nonexistent",
		IsActive: false,
	}

	mockService.On("SetIsActive", mock.Anything, "nonexistent", false).Return(nil, userDomain.ErrUserNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	mockService.AssertExpectations(t)
}

func TestSetIsActive_ValidationError_EmptyUserID(t *testing.T) {
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	reqBody := SetIsActiveRequest{
		UserID:   "",
		IsActive: false,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetIsActive_InvalidJSON(t *testing.T) {
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetIsActive_ActivateUser(t *testing.T) {
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	reqBody := SetIsActiveRequest{
		UserID:   "u1",
		IsActive: true,
	}

	expectedDTO := &userDomain.DTO{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	mockService.On("SetIsActive", mock.Anything, "u1", true).Return(expectedDTO, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.User.IsActive)

	mockService.AssertExpectations(t)
}

func TestSetIsActiveRequest_ToDomainDTO(t *testing.T) {
	req := &SetIsActiveRequest{UserID: "u1", IsActive: true}
	dto := req.ToDomainDTO()
	assert.Equal(t, "u1", dto.UserID)
	assert.True(t, dto.IsActive)
}

func TestFromDomainDTO_User(t *testing.T) {
	d := &userDomain.DTO{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: false}
	resp := FromDomainDTO(d)
	assert.Equal(t, "u1", resp.User.UserID)
	assert.Equal(t, "backend", resp.User.TeamName)
	assert.False(t, resp.User.IsActive)
}

func TestSetIsActive_ServiceError(t *testing.T) {
	logger.InitializeNoop()
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	reqBody := SetIsActiveRequest{UserID: "u1", IsActive: true}

	testErr := errors.New("db error")
	mockService.On("SetIsActive", mock.Anything, "u1", true).Return(nil, testErr)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestSetIsActive_UserNotFound_Mapped(t *testing.T) {
	e := echo.New()
	mockService := new(MockUserService)
	handler := NewHandler(mockService)

	reqBody := SetIsActiveRequest{UserID: "missing", IsActive: false}

	mockService.On("SetIsActive", mock.Anything, "missing", false).Return(nil, userDomain.ErrUserNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SetIsActive(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
