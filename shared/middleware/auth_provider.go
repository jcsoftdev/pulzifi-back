package middleware

import (
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
)

var AuthMiddleware *authmw.AuthMiddleware
var OrgMiddleware *OrganizationMiddleware

func SetAuthMiddleware(middleware *authmw.AuthMiddleware) {
	AuthMiddleware = middleware
}

func SetOrganizationMiddleware(middleware *OrganizationMiddleware) {
	OrgMiddleware = middleware
}
