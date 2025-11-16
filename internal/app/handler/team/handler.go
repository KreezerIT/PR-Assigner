package team

import (
	"net/http"

	"github.com/kreezerit/pr-assigner/internal/app/handler/common"
	"github.com/kreezerit/pr-assigner/internal/app/validator"
	teamDomain "github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	teamService teamDomain.Service
}

func NewHandler(teamService teamDomain.Service) *Handler {
	return &Handler{teamService: teamService}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.POST("/team/add", h.AddTeam)
	e.GET("/team/get", h.GetTeam)
}

func (h *Handler) AddTeam(c echo.Context) error {
	var req CreateTeamRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn("invalid request body", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, "invalid request body"),
		)
	}

	members := make([]teamDomain.Member, len(req.Members))
	for i, m := range req.Members {
		members[i] = teamDomain.Member{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		}
	}

	if err := validator.ValidateCreateTeamRequest(req.TeamName, members); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	domainDTO := req.ToDomainDTO()

	teamDTO, err := h.teamService.CreateTeam(c.Request().Context(), domainDTO)
	if err != nil {
		return common.HandleError(c, err)
	}

	response := FromDomainDTO(teamDTO)
	return c.JSON(http.StatusCreated, response)
}

func (h *Handler) GetTeam(c echo.Context) error {
	teamName := c.QueryParam("team_name")

	if err := validator.ValidateGetTeamRequest(teamName); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	teamDTO, err := h.teamService.GetTeam(c.Request().Context(), teamName)
	if err != nil {
		return common.HandleError(c, err)
	}

	response := FromDomainDTO(teamDTO)
	return c.JSON(http.StatusOK, response.Team)
}
