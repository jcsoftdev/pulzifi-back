package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	getorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_organization"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// GetOrganizationResponseDTO represents the response DTO for Swagger
type GetOrganizationResponseDTO struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string `json:"name" example:"Acme Corp"`
	Subdomain string `json:"subdomain" example:"acme"`
	CreatedAt string `json:"created_at" example:"2025-11-11T12:58:42Z"`
}

// GetOrganizationHandler handles GET /api/organizations/:id
type GetOrganizationHandler struct {
	handler *getorgapp.GetOrganizationHandler
}

// NewGetOrganizationHandler creates a new handler
func NewGetOrganizationHandler(handler *getorgapp.GetOrganizationHandler) *GetOrganizationHandler {
	return &GetOrganizationHandler{
		handler: handler,
	}
}

// Handle processes the HTTP request
// @Summary Get Organization
// @Description Retrieve organization details by ID
// @Tags organizations
// @Accept json
// @Produce json
// @Param id path string true "Organization ID"
// @Success 200 {object} handlers.GetOrganizationResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/organizations/{id} [get]
func (h *GetOrganizationHandler) Handle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.Error("Invalid organization ID", zap.String("id", idStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	resp, err := h.handler.Handle(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// handleError handles domain errors and maps them to HTTP responses
func (h *GetOrganizationHandler) handleError(c *gin.Context, err error) {
	var notFoundError *domainerrors.OrganizationNotFoundError
	var deletedError *domainerrors.OrganizationAlreadyDeletedError

	switch {
	case errors.As(err, &notFoundError):
		logger.Warn("Organization not found", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
	case errors.As(err, &deletedError):
		logger.Warn("Organization is deleted", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
	default:
		logger.Error("Internal server error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
