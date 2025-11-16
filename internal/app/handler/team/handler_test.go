package team

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	teamDomain "github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTeamService struct {
	mock.Mock
}

func (m *MockTeamService) CreateTeam(ctx context.Context, dto *teamDomain.CreateTeamDTO) (*teamDomain.DTO, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*teamDomain.DTO), args.Error(1)
}

func (m *MockTeamService) GetTeam(ctx context.Context, teamName string) (*teamDomain.DTO, error) {
	args := m.Called(ctx, teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*teamDomain.DTO), args.Error(1)
}

func TestMain(m *testing.M) {
	logger.InitializeNoop()
	m.Run()
}

func TestAddTeam_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	reqBody := CreateTeamRequest{
		TeamName: "backend",
		Members: []MemberRequest{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	expectedDTO := &teamDomain.DTO{
		TeamName: "backend",
		Members: []teamDomain.MemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("*team.CreateTeamDTO")).Return(expectedDTO, nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddTeam(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response TeamResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "backend", response.Team.TeamName)
	assert.Len(t, response.Team.Members, 2)

	mockService.AssertExpectations(t)
}

func TestAddTeam_ValidationError(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	reqBody := CreateTeamRequest{
		TeamName: "",
		Members: []MemberRequest{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddTeam(c)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTeam_Success(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	expectedDTO := &teamDomain.DTO{
		TeamName: "backend",
		Members: []teamDomain.MemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}

	mockService.On("GetTeam", mock.Anything, "backend").Return(expectedDTO, nil)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTeam(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response TeamView
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "backend", response.TeamName)

	mockService.AssertExpectations(t)
}

func TestCreateTeamRequest_ToDomainDTO(t *testing.T) {
	req := &CreateTeamRequest{
		TeamName: "backend",
		Members: []MemberRequest{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	dto := req.ToDomainDTO()

	assert.Equal(t, "backend", dto.TeamName)
	assert.Len(t, dto.Members, 2)
	assert.Equal(t, "u1", dto.Members[0].UserID)
	assert.Equal(t, "Bob", dto.Members[1].Username)
	assert.False(t, dto.Members[1].IsActive)
}

func TestFromDomainDTO_Team(t *testing.T) {
	domainDTO := &teamDomain.DTO{
		TeamName: "backend",
		Members: []teamDomain.MemberDTO{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: false},
		},
	}

	resp := FromDomainDTO(domainDTO)

	assert.Equal(t, "backend", resp.Team.TeamName)
	assert.Len(t, resp.Team.Members, 2)
	assert.Equal(t, "u2", resp.Team.Members[1].UserID)
	assert.False(t, resp.Team.Members[1].IsActive)
}

func TestAddTeam_InvalidJSON(t *testing.T) {
	logger.InitializeNoop()
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader([]byte("{bad")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	mockService.AssertExpectations(t)
}

func TestAddTeam_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	reqBody := CreateTeamRequest{TeamName: "backend", Members: []MemberRequest{{UserID: "u1", Username: "A", IsActive: true}}}

	testErr := errors.New("create failed")
	mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("*team.CreateTeamDTO")).Return(nil, testErr)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTeam_ServiceError(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	testErr := errors.New("repo failed")
	mockService.On("GetTeam", mock.Anything, "backend").Return(nil, testErr)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetTeam_TeamNotFound(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	mockService.On("GetTeam", mock.Anything, "unknown").Return(nil, teamDomain.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=unknown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetTeam_ValidationError_EmptyName(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAddTeam_TeamAlreadyExists(t *testing.T) {
	e := echo.New()
	mockService := new(MockTeamService)
	handler := NewHandler(mockService)

	reqBody := CreateTeamRequest{TeamName: "backend", Members: []MemberRequest{{UserID: "u1", Username: "Alice", IsActive: true}}}
	mockService.On("CreateTeam", mock.Anything, mock.AnythingOfType("*team.CreateTeamDTO")).Return(nil, teamDomain.ErrTeamAlreadyExists)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.AddTeam(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
