package invitemember

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type InviteMemberHandler struct {
	repo repositories.TeamMemberRepository
	db   *sql.DB
}

func NewInviteMemberHandler(repo repositories.TeamMemberRepository, db *sql.DB) *InviteMemberHandler {
	return &InviteMemberHandler{repo: repo, db: db}
}

func (h *InviteMemberHandler) Handle(ctx context.Context, subdomain string, inviterID uuid.UUID, req *InviteMemberRequest) (*InviteMemberResponse, error) {
	role := strings.ToUpper(req.Role)
	if role == "" {
		role = "MEMBER"
	}

	orgID, err := h.repo.GetOrganizationIDBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	// Find user by email; auto-create if not found
	user, err := h.repo.FindUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	var userID uuid.UUID
	isNewUser := false
	if user == nil {
		// Auto-create user with a random temporary password
		tmpPass, hashErr := generateTemporaryPassword()
		if hashErr != nil {
			return nil, hashErr
		}
		userID, err = h.repo.CreateUser(ctx, req.Email, "", "", tmpPass)
		if err != nil {
			return nil, err
		}
		isNewUser = true
	} else {
		userID = user.UserID
	}

	// Check if already member
	existing, err := h.repo.GetByUserAndOrg(ctx, orgID, userID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyMember
	}

	inviterIDPtr := &inviterID
	invitationStatus := "active"
	if isNewUser {
		invitationStatus = "pending"
	}
	member, err := h.repo.AddMember(ctx, orgID, userID, role, inviterIDPtr, invitationStatus)
	if err != nil {
		return nil, err
	}

	resp := &InviteMemberResponse{
		ID:        member.ID,
		UserID:    member.UserID,
		Role:      member.Role,
		FirstName: member.FirstName,
		LastName:  member.LastName,
		Email:     member.Email,
		JoinedAt:  member.JoinedAt,
		IsNewUser: isNewUser,
	}

	// Generate a password reset token for new users so they can set their own password
	if isNewUser && h.db != nil {
		token, err := GenerateResetToken(ctx, h.db, userID)
		if err != nil {
			logger.Error("Failed to generate set-password token for invited user", zap.Error(err))
		} else {
			resp.SetPasswordToken = token
		}
	}

	return resp, nil
}

// GenerateResetToken creates a password reset token and stores it in public.password_resets.
func GenerateResetToken(ctx context.Context, db *sql.DB, userID uuid.UUID) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(72 * time.Hour) // 3 days for invite tokens

	_, err := db.ExecContext(ctx,
		`INSERT INTO public.password_resets (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New(), userID, token, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

// generateTemporaryPassword creates a random password hash for new invited users.
// The user must reset their password via the password reset flow before logging in.
func generateTemporaryPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	raw := base64.StdEncoding.EncodeToString(b)
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
