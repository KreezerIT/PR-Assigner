package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	prHandler "github.com/kreezerit/pr-assigner/internal/app/handler/pullrequest"
	statsHandler "github.com/kreezerit/pr-assigner/internal/app/handler/statistics"
	teamHandler "github.com/kreezerit/pr-assigner/internal/app/handler/team"
	userHandler "github.com/kreezerit/pr-assigner/internal/app/handler/user"
	"github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
	"github.com/kreezerit/pr-assigner/internal/domain/statistics"
	"github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/persistence/postgres"
)

type IntegrationTestSuite struct {
	suite.Suite
	db             *sql.DB
	e              *echo.Echo
	dsn            string
	migrationsPath string
}

func (s *IntegrationTestSuite) SetupSuite() {
	err := logger.Initialize("debug")
	s.Require().NoError(err)

	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5433")
	user := getEnv("TEST_DB_USER", "prservice_test")
	password := getEnv("TEST_DB_PASSWORD", "test_pass")
	dbname := getEnv("TEST_DB_NAME", "pr_reviewer_test")

	s.dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	s.db, err = sql.Open("postgres", s.dsn)
	s.Require().NoError(err)

	err = s.db.Ping()
	s.Require().NoError(err)

	// Arrange
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	s.migrationsPath = filepath.ToSlash(filepath.Join(projectRoot, "migrations"))

	// Act
	err = runMigrations(s.db, s.migrationsPath)
	// Assert
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.db != nil {
		err := s.db.Close()
		s.Require().NoError(err)
	}
}

func (s *IntegrationTestSuite) SetupTest() {
	// Arrange
	err := cleanupDB(s.db)
	// Assert
	s.Require().NoError(err)

	// Arrange
	s.e = echo.New()

	teamRepo := postgres.NewTeamRepository(s.db)
	userRepo := postgres.NewUserRepository(s.db)
	prRepo := postgres.NewPullRequestRepository(s.db)
	statsRepo := postgres.NewStatisticsRepository(s.db)

	teamService := team.NewService(teamRepo, userRepo)
	userService := user.NewService(userRepo)
	prService := pullrequest.NewService(prRepo, userRepo)
	statsService := statistics.NewService(statsRepo)

	teamH := teamHandler.NewHandler(teamService)
	userH := userHandler.NewHandler(userService)
	prH := prHandler.NewHandler(prService)
	statsH := statsHandler.NewHandler(statsService)

	// Act: регистрация маршрутов
	teamH.RegisterRoutes(s.e)
	userH.RegisterRoutes(s.e)
	prH.RegisterRoutes(s.e)
	statsH.RegisterRoutes(s.e)
}

func (s *IntegrationTestSuite) TestFullWorkflow() {
	// Arrange
	teamReq := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	var teamResp teamHandler.TeamResponse
	err := json.Unmarshal(rec.Body.Bytes(), &teamResp)
	s.Require().NoError(err)
	assert.Equal(s.T(), "backend", teamResp.Team.TeamName)
	assert.Len(s.T(), teamResp.Team.Members, 3)

	// Arrange
	req = httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	var teamView teamHandler.TeamView
	err = json.Unmarshal(rec.Body.Bytes(), &teamView)
	s.Require().NoError(err)
	assert.Equal(s.T(), "backend", teamView.TeamName)

	// Arrange
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "u1",
	}

	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	var prResp prHandler.PRResponse
	err = json.Unmarshal(rec.Body.Bytes(), &prResp)
	s.Require().NoError(err)
	assert.Equal(s.T(), "pr-1", prResp.PR.PullRequestID)
	assert.Equal(s.T(), "OPEN", prResp.PR.Status)
	assert.LessOrEqual(s.T(), len(prResp.PR.AssignedReviewers), 2)
	assert.NotNil(s.T(), prResp.PR.CreatedAt)

	if len(prResp.PR.AssignedReviewers) > 0 {
		// Arrange
		reviewerID := prResp.PR.AssignedReviewers[0]

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/getReview?user_id=%s", reviewerID), nil)
		rec = httptest.NewRecorder()
		// Act
		s.e.ServeHTTP(rec, req)
		// Assert
		assert.Equal(s.T(), http.StatusOK, rec.Code)

		var reviewsResp prHandler.UserReviewsResponse
		json.Unmarshal(rec.Body.Bytes(), &reviewsResp)
		assert.Equal(s.T(), reviewerID, reviewsResp.UserID)
		assert.Len(s.T(), reviewsResp.PullRequests, 1)
		assert.Equal(s.T(), "pr-1", reviewsResp.PullRequests[0].PullRequestID)
	}

	// Arrange
	setActiveReq := map[string]interface{}{
		"user_id":   "u3",
		"is_active": false,
	}

	body, _ = json.Marshal(setActiveReq)
	req = httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	var userResp userHandler.UserResponse
	json.Unmarshal(rec.Body.Bytes(), &userResp)
	assert.False(s.T(), userResp.User.IsActive)

	// Arrange
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-1",
	}

	body, _ = json.Marshal(mergeReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	json.Unmarshal(rec.Body.Bytes(), &prResp)
	assert.Equal(s.T(), "MERGED", prResp.PR.Status)
	assert.NotNil(s.T(), prResp.PR.MergedAt)

	// Arrange
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	json.Unmarshal(rec.Body.Bytes(), &prResp)
	assert.Equal(s.T(), "MERGED", prResp.PR.Status)
}

func (s *IntegrationTestSuite) TestReassignReviewer() {
	// Arrange
	teamReq := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
			{"user_id": "u4", "username": "Dave", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	// Arrange
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-2",
		"pull_request_name": "Fix bug",
		"author_id":         "u1",
	}

	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	var prResp prHandler.PRResponse
	json.Unmarshal(rec.Body.Bytes(), &prResp)

	if len(prResp.PR.AssignedReviewers) > 0 {
		oldReviewerID := prResp.PR.AssignedReviewers[0]

		// Arrange
		reassignReq := map[string]interface{}{
			"pull_request_id": "pr-2",
			"old_user_id":     oldReviewerID,
		}

		body, _ = json.Marshal(reassignReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		// Act
		s.e.ServeHTTP(rec, req)
		// Assert
		assert.Equal(s.T(), http.StatusOK, rec.Code)

		var reassignResp prHandler.ReassignResponse
		json.Unmarshal(rec.Body.Bytes(), &reassignResp)
		assert.NotEmpty(s.T(), reassignResp.ReplacedBy)
		assert.NotContains(s.T(), reassignResp.PR.AssignedReviewers, oldReviewerID)
		assert.Contains(s.T(), reassignResp.PR.AssignedReviewers, reassignResp.ReplacedBy)
	}
}

func (s *IntegrationTestSuite) TestCannotReassignAfterMerge() {
	// Arrange
	teamReq := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-3",
		"pull_request_name": "Update docs",
		"author_id":         "u1",
	}

	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	var prResp prHandler.PRResponse
	json.Unmarshal(rec.Body.Bytes(), &prResp)

	// Arrange
	mergeReq := map[string]interface{}{
		"pull_request_id": "pr-3",
	}

	body, _ = json.Marshal(mergeReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusOK, rec.Code)

	if len(prResp.PR.AssignedReviewers) > 0 {
		// Arrange
		reassignReq := map[string]interface{}{
			"pull_request_id": "pr-3",
			"old_user_id":     prResp.PR.AssignedReviewers[0],
		}

		body, _ = json.Marshal(reassignReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		// Act
		s.e.ServeHTTP(rec, req)
		// Assert
		assert.Equal(s.T(), http.StatusConflict, rec.Code)
	}
}

func (s *IntegrationTestSuite) TestPRWithNoReviewers() {
	// Arrange
	teamReq := map[string]interface{}{
		"team_name": "solo",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-4",
		"pull_request_name": "Solo PR",
		"author_id":         "u1",
	}

	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	var prResp prHandler.PRResponse
	json.Unmarshal(rec.Body.Bytes(), &prResp)
	// Assert
	assert.Empty(s.T(), prResp.PR.AssignedReviewers)
}

func (s *IntegrationTestSuite) TestInactiveUsersNotAssigned() {
	// Arrange
	teamReq := map[string]interface{}{
		"team_name": "backend",
		"members": []map[string]interface{}{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": false},
			{"user_id": "u3", "username": "Charlie", "is_active": false},
		},
	}

	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	prReq := map[string]interface{}{
		"pull_request_id":   "pr-5",
		"pull_request_name": "Test inactive",
		"author_id":         "u1",
	}

	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	assert.Equal(s.T(), http.StatusCreated, rec.Code)

	var prResp prHandler.PRResponse
	json.Unmarshal(rec.Body.Bytes(), &prResp)

	// Assert
	for _, reviewerID := range prResp.PR.AssignedReviewers {
		assert.NotEqual(s.T(), "u2", reviewerID)
		assert.NotEqual(s.T(), "u3", reviewerID)
	}
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests in short mode")
	}

	suite.Run(t, new(IntegrationTestSuite))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (s *IntegrationTestSuite) TestTeamGet_NotFound() {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=unknown", nil)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusNotFound {
		s.T().Fatalf("expected 404, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestReassign_ValidationError() {
	// Arrange
	body, _ := json.Marshal(map[string]any{"pull_request_id": "", "old_user_id": ""})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for empty fields, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestUserSetActive_ValidationError() {
	// Arrange
	body, _ := json.Marshal(map[string]any{"user_id": "", "is_active": false})
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for empty user ID, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestReassignReviewer_PRNotFound() {
	// Arrange
	body, _ := json.Marshal(map[string]any{"pull_request_id": "nonexistent", "old_user_id": "u1"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusNotFound {
		s.T().Fatalf("expected 404 for nonexistent PR, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestGetUserReviews_EmptyList() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members":   []map[string]any{{"user_id": "u1", "username": "A", "is_active": true}},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusOK {
		s.T().Fatalf("expected 200, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestTeamGet_ValidationError() {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=", nil)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for empty team name, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestTeamAdd_InvalidJSON() {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader([]byte("{bad json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestUserSetActive_InvalidJSON() {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestReassign_InvalidJSON() {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestMerge_InvalidJSON() {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for invalid JSON, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestPR_Merge_NotFound() {
	// Arrange
	body, _ := json.Marshal(map[string]any{"pull_request_id": "nope"})
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusNotFound {
		s.T().Fatalf("expected 404, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestPR_GetUserReviews_ValidationError() {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=", nil)
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestStatistics_GlobalStats() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodGet, "/statistics/global", nil)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusOK {
		s.T().Fatalf("expected 200, got %d", rec.Code)
	}

	var statsResp statsHandler.GlobalStatsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &statsResp)
	if statsResp.TotalUsers < 2 {
		s.T().Fatalf("expected at least 2 users in stats")
	}
}

func (s *IntegrationTestSuite) TestStatistics_UserStats() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodGet, "/statistics/user/u1", nil)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusOK {
		s.T().Fatalf("expected 200, got %d", rec.Code)
	}

	var userStats statsHandler.UserStatsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &userStats)
	if userStats.UserID != "u1" {
		s.T().Fatalf("expected u1, got %s", userStats.UserID)
	}
}

func (s *IntegrationTestSuite) TestStatistics_TeamStats() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": false},
		},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodGet, "/statistics/team/backend", nil)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusOK {
		s.T().Fatalf("expected 200, got %d", rec.Code)
	}

	var teamStats statsHandler.TeamStatsResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &teamStats)
	if teamStats.TeamName != "backend" {
		s.T().Fatalf("expected backend, got %s", teamStats.TeamName)
	}
	if teamStats.TotalMembers != 2 {
		s.T().Fatalf("expected 2 members, got %d", teamStats.TotalMembers)
	}
	if teamStats.ActiveMembers != 1 {
		s.T().Fatalf("expected 1 active member, got %d", teamStats.ActiveMembers)
	}
}

func (s *IntegrationTestSuite) TestStatistics_TopReviewers() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Charlie", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	prReq := map[string]any{"pull_request_id": "pr-1", "pull_request_name": "Test", "author_id": "u1"}
	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodGet, "/statistics/top-reviewers?limit=5", nil)
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusOK {
		s.T().Fatalf("expected 200, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestMigrations_ErrorPaths() {
	// Act
	err := runMigrations(s.db, "Z:/nonexistent/path")
	// Assert
	if err == nil {
		s.T().Fatalf("expected error for invalid migrations path")
	}
}

func (s *IntegrationTestSuite) TestCreatePR_AuthorNotFound() {
	// Arrange
	prReq := map[string]any{"pull_request_id": "pr-x", "pull_request_name": "Test", "author_id": "nonexistent"}
	body, _ := json.Marshal(prReq)
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusNotFound {
		s.T().Fatalf("expected 404 for nonexistent author, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestTeamAlreadyExists() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members":   []map[string]any{{"user_id": "u1", "username": "A", "is_active": true}},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusBadRequest {
		s.T().Fatalf("expected 400 for duplicate team, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestPR_AlreadyExists() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members":   []map[string]any{{"user_id": "u1", "username": "A", "is_active": true}},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	prReq := map[string]any{"pull_request_id": "pr-1", "pull_request_name": "Test", "author_id": "u1"}
	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusConflict {
		s.T().Fatalf("expected 409 for duplicate PR, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestReassignReviewer_NotAssigned() {
	// Arrange
	teamReq := map[string]any{
		"team_name": "backend",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
		},
	}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	prReq := map[string]any{"pull_request_id": "pr-1", "pull_request_name": "Test", "author_id": "u1"}
	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)

	// Arrange
	reassignReq := map[string]any{"pull_request_id": "pr-1", "old_user_id": "u1"}
	body, _ = json.Marshal(reassignReq)
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusConflict {
		s.T().Fatalf("expected 409 for not assigned reviewer, got %d", rec.Code)
	}
}

func (s *IntegrationTestSuite) TestUserSetActive_NotFound() {
	// Arrange
	setActiveReq := map[string]any{"user_id": "nonexistent", "is_active": false}
	body, _ := json.Marshal(setActiveReq)
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	// Act
	s.e.ServeHTTP(rec, req)
	// Assert
	if rec.Code != http.StatusNotFound {
		s.T().Fatalf("expected 404 for nonexistent user, got %d", rec.Code)
	}
}
