package entities

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Permission struct {
	ID          uuid.UUID
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RolePermission struct {
	RoleID       uuid.UUID
	PermissionID uuid.UUID
	CreatedAt    time.Time
}

type UserRole struct {
	UserID    uuid.UUID
	RoleID    uuid.UUID
	CreatedAt time.Time
}
