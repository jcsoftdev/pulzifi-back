package http

import (
	"github.com/gin-gonic/gin"
	createorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/create_organization"
	getorgapp "github.com/jcsoftdev/pulzifi-back/modules/organization/application/get_organization"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/infrastructure/http/handlers"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router sets up all HTTP routes for the organization module
type Router struct {
	createOrgHandler *handlers.CreateOrganizationHandler
	getOrgHandler    *handlers.GetOrganizationHandler
}

// NewRouter creates a new router instance
func NewRouter(
	createOrgHandler *createorgapp.CreateOrganizationHandler,
	getOrgHandler *getorgapp.GetOrganizationHandler,
) *Router {
	return &Router{
		createOrgHandler: handlers.NewCreateOrganizationHandler(createOrgHandler),
		getOrgHandler:    handlers.NewGetOrganizationHandler(getOrgHandler),
	}
}

// Setup registers all routes with the Gin router
func (r *Router) Setup(router *gin.Engine) {
	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// POST /api/organizations - Create new organization
	router.POST("/api/organizations", r.createOrgHandler.Handle)

	// GET /api/organizations/:id - Get organization by ID
	router.GET("/api/organizations/:id", r.getOrgHandler.Handle)
}
