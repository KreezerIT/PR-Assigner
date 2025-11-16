package pullrequest

import (
	"net/http"

	"github.com/kreezerit/pr-assigner/internal/app/handler/common"
	"github.com/kreezerit/pr-assigner/internal/app/validator"
	prDomain "github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	prService prDomain.Service
}

func NewHandler(prService prDomain.Service) *Handler {
	return &Handler{prService: prService}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.POST("/pullRequest/create", h.CreatePR)
	e.POST("/pullRequest/merge", h.MergePR)
	e.POST("/pullRequest/reassign", h.ReassignReviewer)
	e.GET("/users/getReview", h.GetUserReviews)
}

func (h *Handler) CreatePR(c echo.Context) error {
	var req CreatePRRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn("invalid request body", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, "invalid request body"),
		)
	}

	if err := validator.ValidateCreatePRRequest(req.PullRequestID, req.PullRequestName, req.AuthorID); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	domainDTO := req.ToDomainDTO()

	prDTO, err := h.prService.CreatePR(c.Request().Context(), domainDTO)
	if err != nil {
		return common.HandleError(c, err)
	}

	response := FromDomainDTO(prDTO)
	return c.JSON(http.StatusCreated, response)
}

func (h *Handler) MergePR(c echo.Context) error {
	var req MergePRRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn("invalid request body", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, "invalid request body"),
		)
	}

	if err := validator.ValidateMergePRRequest(req.PullRequestID); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	prDTO, err := h.prService.MergePR(c.Request().Context(), req.PullRequestID)
	if err != nil {
		return common.HandleError(c, err)
	}

	response := FromDomainDTO(prDTO)
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) ReassignReviewer(c echo.Context) error {
	var req ReassignRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn("invalid request body", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, "invalid request body"),
		)
	}

	if err := validator.ValidateReassignRequest(req.PullRequestID, req.OldUserID); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	resultDTO, err := h.prService.ReassignReviewer(
		c.Request().Context(),
		req.PullRequestID,
		req.OldUserID,
	)
	if err != nil {
		return common.HandleError(c, err)
	}

	response := FromReassignResultDTO(resultDTO)
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetUserReviews(c echo.Context) error {
	userID := c.QueryParam("user_id")

	if err := validator.ValidateGetUserReviewsRequest(userID); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	prsDTO, err := h.prService.GetUserReviews(c.Request().Context(), userID)
	if err != nil {
		return common.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, UserReviewsResponse{
		UserID:       userID,
		PullRequests: FromShortDTOs(prsDTO),
	})
}
