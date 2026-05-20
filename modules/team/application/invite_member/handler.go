package invitemember

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type InviteMemberHandler struct {
	repo           repositories.TeamMemberRepository
	tokenGenerator services.TokenGenerator
}

func NewInviteMemberHandler(repo repositories.TeamMemberRepository, tokenGenerator services.TokenGenerator) *InviteMemberHandler {
	return &InviteMemberHandler{repo: repo, tokenGenerator: tokenGenerator}
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
	if isNewUser && h.tokenGenerator != nil {
		token, err := h.tokenGenerator.Generate(ctx, userID)
		if err != nil {
			logger.Error("Failed to generate set-password token for invited user", zap.Error(err))
		} else {
			resp.SetPasswordToken = token
		}
	}

	return resp, nil
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
