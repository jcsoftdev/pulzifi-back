package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	createorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/create_organization"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// CreateOrganizationHandler handles POST /api/organizations
type CreateOrganizationHandler struct {
	handler *createorgapp.CreateOrganizationHandler
}

// NewCreateOrganizationHandler creates a new handler
func NewCreateOrganizationHandler(handler *createorgapp.CreateOrganizationHandler) *CreateOrganizationHandler {
	return &CreateOrganizationHandler{
		handler: handler,
	}
}

// Handle processes the HTTP request
// @Summary Create Organization
// @Description Create a new organization with the provided name and subdomain
// @Tags organizations
// @Accept json
// @Produce json
// @Param request body createorgapp.CreateOrganizationRequest true "Create Organization Request"
// @Success 201 {object} createorgapp.CreateOrganizationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/organizations [post]
func (h *CreateOrganizationHandler) Handle(c *gin.Context) {
	var req createorgapp.Request

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Extract user ID from context (set by JWT middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDUUID, ok := userID.(uuid.UUID)
	if !ok {
		logger.Error("Invalid user ID type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	resp, err := h.handler.Handle(c.Request.Context(), &req, userIDUUID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// handleError handles domain errors and maps them to HTTP responses
func (h *CreateOrganizationHandler) handleError(c *gin.Context, err error) {
	var subdomainError *domainerrors.SubdomainAlreadyExistsError
	var invalidDataError *domainerrors.InvalidOrganizationDataError

	switch {
	case errors.As(err, &subdomainError):
		logger.Warn("Subdomain already exists", zap.Error(err))
		c.JSON(http.StatusConflict, gin.H{"error": "Subdomain already exists"})
	case errors.As(err, &invalidDataError):
		logger.Warn("Invalid organization data", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		logger.Error("Internal server error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
