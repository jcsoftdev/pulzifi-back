package contextkeys

// ContextKey is the type used for all context keys in this project.
// Using a named type prevents collisions with other packages.
type ContextKey string

const (
	// UserIDKey is the context key under which the authenticated user's UUID string is stored.
	// It is written by modules/auth/infrastructure/middleware and read by shared/middleware.
	UserIDKey ContextKey = "user_id"

	// UserEmailKey is the context key under which the authenticated user's email is stored.
	UserEmailKey ContextKey = "user_email"

	// UserRolesKey is the context key under which the authenticated user's roles slice is stored.
	UserRolesKey ContextKey = "user_roles"

	// UserPermsKey is the context key under which the authenticated user's permissions slice is stored.
	UserPermsKey ContextKey = "user_permissions"
)
