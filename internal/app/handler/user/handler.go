package user

import (
	"net/http"

	"github.com/kreezerit/pr-assigner/internal/app/handler/common"
	"github.com/kreezerit/pr-assigner/internal/app/validator"
	userDomain "github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Handler struct {
	userService userDomain.Service
}

func NewHandler(userService userDomain.Service) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.POST("/users/setIsActive", h.SetIsActive)
}

func (h *Handler) SetIsActive(c echo.Context) error {
	var req SetIsActiveRequest
	if err := c.Bind(&req); err != nil {
		logger.Warn("invalid request body", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, "invalid request body"),
		)
	}

	if err := validator.ValidateSetIsActiveRequest(req.UserID); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return c.JSON(
			http.StatusBadRequest,
			common.NewErrorResponse(common.CodeInvalidRequest, err.Error()),
		)
	}

	userDTO, err := h.userService.SetIsActive(c.Request().Context(), req.UserID, req.IsActive)
	if err != nil {
		return common.HandleError(c, err)
	}

	response := FromDomainDTO(userDTO)
	return c.JSON(http.StatusOK, response)
}
