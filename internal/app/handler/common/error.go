package common

import (
	"errors"
	"net/http"

	"github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
	"github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/labstack/echo/v4"
)

type ErrorCode string

// Набор кодов ошибок, возвращаемых в API
const (
	// Доменные
	CodeTeamExists  ErrorCode = "TEAM_EXISTS"
	CodePRExists    ErrorCode = "PR_EXISTS"
	CodePRMerged    ErrorCode = "PR_MERGED"
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"

	// Общие
	CodeNotFound       ErrorCode = "NOT_FOUND"
	CodeInvalidRequest ErrorCode = "INVALID_REQUEST"
	CodeInternalError  ErrorCode = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func NewErrorResponse(code ErrorCode, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}

// HandleError сопоставляет доменную ошибку с HTTP‑ответом
func HandleError(c echo.Context, err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, team.ErrTeamAlreadyExists):
		return c.JSON(http.StatusBadRequest,
			NewErrorResponse(CodeTeamExists, "team_name already exists"))

	case errors.Is(err, team.ErrTeamNotFound):
		return c.JSON(http.StatusNotFound,
			NewErrorResponse(CodeNotFound, "team not found"))

	case errors.Is(err, user.ErrUserNotFound):
		return c.JSON(http.StatusNotFound,
			NewErrorResponse(CodeNotFound, "user not found"))

	case errors.Is(err, pullrequest.ErrPRAlreadyExists):
		return c.JSON(http.StatusConflict,
			NewErrorResponse(CodePRExists, "PR id already exists"))

	case errors.Is(err, pullrequest.ErrPRNotFound):
		return c.JSON(http.StatusNotFound,
			NewErrorResponse(CodeNotFound, "pull request not found"))

	case errors.Is(err, pullrequest.ErrPRMerged):
		return c.JSON(http.StatusConflict,
			NewErrorResponse(CodePRMerged, "cannot reassign on merged PR"))

	case errors.Is(err, pullrequest.ErrReviewerNotAssigned):
		return c.JSON(http.StatusConflict,
			NewErrorResponse(CodeNotAssigned, "reviewer is not assigned to this PR"))

	case errors.Is(err, pullrequest.ErrNoCandidate):
		return c.JSON(http.StatusConflict,
			NewErrorResponse(CodeNoCandidate, "no active replacement candidate in team"))

	case errors.Is(err, pullrequest.ErrAuthorNotFound):
		return c.JSON(http.StatusNotFound,
			NewErrorResponse(CodeNotFound, "author not found"))

	default:
		return c.JSON(http.StatusInternalServerError,
			NewErrorResponse(CodeInternalError, "internal server error"))
	}
}
