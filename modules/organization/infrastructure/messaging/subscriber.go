package messaging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Subscriber subscribes to topics using MessageBus
type Subscriber struct {
	bus eventbus.MessageBus
	db  *sql.DB
}

// NewSubscriber creates a new subscriber
func NewSubscriber(bus eventbus.MessageBus, db *sql.DB) *Subscriber {
	return &Subscriber{
		bus: bus,
		db:  db,
	}
}

// ListenToEvents subscribes to topics
func (s *Subscriber) ListenToEvents(ctx context.Context) {
	logger.Info("Starting subscriber for organization module")

	// Topics this module subscribes to
	topics := []string{
		"user.deleted",
		"workspace.created",
	}

	for _, topic := range topics {
		// Define handler wrapper to pass context
		handler := func(t string, payload []byte) {
			logger.Info("Received event", zap.String("topic", t), zap.Int("payload_size", len(payload)))
			s.handleEvent(context.Background(), t, "", payload) // In-memory bus doesn't support key yet, passing empty
		}

		if err := s.bus.Subscribe(topic, handler); err != nil {
			logger.Error("Failed to subscribe to topic", zap.String("topic", topic), zap.Error(err))
		}
	}

	logger.Info("Organization module subscriber ready")

	// Keep alive until context cancelled
	<-ctx.Done()
	logger.Info("Stopping subscriber")
}

// handleEvent processes incoming Kafka messages
func (s *Subscriber) handleEvent(ctx context.Context, topic string, key string, payload []byte) {
	switch topic {
	case "user.deleted":
		s.handleUserDeleted(ctx, key, payload)
	case "workspace.created":
		s.handleWorkspaceCreated(ctx, key, payload)
	default:
		logger.Warn("Unknown event topic", zap.String("topic", topic))
	}
}

// userDeletedPayload represents the JSON payload for a user.deleted event
type userDeletedPayload struct {
	UserID string `json:"user_id"`
}

// handleUserDeleted handles user deletion events from other modules.
// When a user is deleted, we cascade:
//   - For each organization where the user is a member:
//     1. If the user is the sole owner, soft-delete the organization.
//     2. Remove the user from organization_members.
//     3. Remove the user from user_roles.
func (s *Subscriber) handleUserDeleted(ctx context.Context, _ string, payload []byte) {
	// Parse the event payload
	var event userDeletedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Error("Failed to parse user.deleted payload", zap.Error(err))
		return
	}

	userID, err := uuid.Parse(event.UserID)
	if err != nil {
		logger.Error("Invalid user_id in user.deleted payload",
			zap.String("raw_user_id", event.UserID),
			zap.Error(err),
		)
		return
	}

	logger.Info("Processing user.deleted event", zap.String("user_id", userID.String()))

	// 1. Find all organizations where the user is a member
	orgIDs, err := s.findUserOrganizations(ctx, userID)
	if err != nil {
		logger.Error("Failed to query user organizations",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return
	}

	if len(orgIDs) == 0 {
		logger.Info("User has no organization memberships, nothing to cascade",
			zap.String("user_id", userID.String()),
		)
	}

	for _, orgID := range orgIDs {
		logFields := []zap.Field{
			zap.String("user_id", userID.String()),
			zap.String("organization_id", orgID.String()),
		}

		// 2. Check if the user is the sole owner
		soleOwner, err := s.isSoleOwner(ctx, orgID, userID)
		if err != nil {
			logger.Error("Failed to check sole owner status", append(logFields, zap.Error(err))...)
			continue
		}

		// 3. If sole owner, soft-delete the organization
		if soleOwner {
			if err := s.softDeleteOrganization(ctx, orgID); err != nil {
				logger.Error("Failed to soft-delete organization", append(logFields, zap.Error(err))...)
				continue
			}
			logger.Info("Soft-deleted organization (user was sole owner)", logFields...)
		}

		// 4. Remove user from organization_members
		if err := s.removeOrganizationMember(ctx, orgID, userID); err != nil {
			logger.Error("Failed to remove user from organization_members", append(logFields, zap.Error(err))...)
			continue
		}
		logger.Info("Removed user from organization_members", logFields...)
	}

	// 5. Remove user from user_roles
	if err := s.removeUserRoles(ctx, userID); err != nil {
		logger.Error("Failed to remove user from user_roles",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return
	}
	logger.Info("Removed user from user_roles", zap.String("user_id", userID.String()))

	logger.Info("Finished processing user.deleted cascade",
		zap.String("user_id", userID.String()),
		zap.Int("organizations_processed", len(orgIDs)),
	)
}

// findUserOrganizations returns all organization IDs where the user is an active member.
func (s *Subscriber) findUserOrganizations(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT organization_id
		FROM public.organization_members
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query organization_members: %w", err)
	}
	defer rows.Close()

	var orgIDs []uuid.UUID
	for rows.Next() {
		var orgID uuid.UUID
		if err := rows.Scan(&orgID); err != nil {
			return nil, fmt.Errorf("scan organization_id: %w", err)
		}
		orgIDs = append(orgIDs, orgID)
	}
	return orgIDs, rows.Err()
}

// isSoleOwner checks whether the given user is the only active member with role='owner'
// in the specified organization.
func (s *Subscriber) isSoleOwner(ctx context.Context, orgID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM public.organization_members
		WHERE organization_id = $1
		  AND role = 'owner'
		  AND deleted_at IS NULL
		  AND user_id != $2
	`
	var otherOwners int
	if err := s.db.QueryRowContext(ctx, query, orgID, userID).Scan(&otherOwners); err != nil {
		return false, fmt.Errorf("count other owners: %w", err)
	}
	return otherOwners == 0, nil
}

// softDeleteOrganization sets deleted_at = NOW() on the organization row.
func (s *Subscriber) softDeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
	query := `UPDATE public.organizations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := s.db.ExecContext(ctx, query, orgID)
	if err != nil {
		return fmt.Errorf("soft-delete organization: %w", err)
	}
	return nil
}

// removeOrganizationMember soft-deletes the user's membership in the given organization.
func (s *Subscriber) removeOrganizationMember(ctx context.Context, orgID, userID uuid.UUID) error {
	query := `UPDATE public.organization_members SET deleted_at = NOW() WHERE organization_id = $1 AND user_id = $2 AND deleted_at IS NULL`
	_, err := s.db.ExecContext(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("remove organization member: %w", err)
	}
	return nil
}

// removeUserRoles deletes all role assignments for the given user.
func (s *Subscriber) removeUserRoles(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM public.user_roles WHERE user_id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("delete user_roles: %w", err)
	}
	return nil
}

// handleWorkspaceCreated handles workspace creation events from other modules
// When a workspace is created, we may need to update organization metadata
func (s *Subscriber) handleWorkspaceCreated(ctx context.Context, workspaceID string, payload []byte) {
	logger.Info("Processing workspace.created event",
		zap.String("workspace_id", workspaceID),
	)

	// This is informational - just log that we received it
	// Organization module doesn't need to do anything when workspaces are created
	// This handler exists for future extensibility
}
