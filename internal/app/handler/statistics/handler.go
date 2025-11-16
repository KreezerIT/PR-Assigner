package statistics

import (
	"net/http"
	"strconv"

	"github.com/kreezerit/pr-assigner/internal/app/handler/common"
	statsDomain "github.com/kreezerit/pr-assigner/internal/domain/statistics"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	statsService statsDomain.Service
}

func NewHandler(statsService statsDomain.Service) *Handler {
	return &Handler{statsService: statsService}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/statistics/global", h.GetGlobalStats)
	e.GET("/statistics/user/:user_id", h.GetUserStats)
	e.GET("/statistics/team/:team_name", h.GetTeamStats)
	e.GET("/statistics/top-reviewers", h.GetTopReviewers)
}

func (h *Handler) GetGlobalStats(c echo.Context) error {
	statsDTO, err := h.statsService.GetGlobalStats(c.Request().Context())
	if err != nil {
		logger.Error("failed to get global stats", zap.Error(err))
		return common.HandleError(c, err)
	}

	response := FromGlobalStatsDTO(statsDTO)
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetUserStats(c echo.Context) error {
	userID := c.Param("user_id")

	if userID == "" {
		return c.JSON(http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidRequest, "user_id is required"))
	}

	statsDTO, err := h.statsService.GetUserStats(c.Request().Context(), userID)
	if err != nil {
		logger.Error("failed to get user stats", zap.Error(err))
		return common.HandleError(c, err)
	}

	response := FromUserStatsDTO(statsDTO)
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTeamStats(c echo.Context) error {
	teamName := c.Param("team_name")

	if teamName == "" {
		return c.JSON(http.StatusBadRequest, common.NewErrorResponse(common.CodeInvalidRequest, "team_name is required"))
	}

	statsDTO, err := h.statsService.GetTeamStats(c.Request().Context(), teamName)
	if err != nil {
		logger.Error("failed to get team stats", zap.Error(err))
		return common.HandleError(c, err)
	}

	response := FromTeamStatsDTO(statsDTO)
	return c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTopReviewers(c echo.Context) error {
	limitStr := c.QueryParam("limit")
	limit := 10

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	reviewersDTOs, err := h.statsService.GetTopReviewers(c.Request().Context(), limit)
	if err != nil {
		logger.Error("failed to get top reviewers", zap.Error(err))
		return common.HandleError(c, err)
	}

	response := FromTopReviewersDTOs(reviewersDTOs)
	return c.JSON(http.StatusOK, response)
}
