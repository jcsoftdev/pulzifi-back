package adminwiring

import (
	"context"

	"github.com/google/uuid"
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// pendingUserAdapter implements adminservices.PendingUserReader by wrapping
// the auth module's UserRepository. Lives in cmd/wiring/admin so the admin
// module itself never imports from the auth module.
type pendingUserAdapter struct {
	userRepo authrepos.UserRepository
}

// NewPendingUserAdapter creates a PendingUserReader backed by auth's UserRepository.
func NewPendingUserAdapter(userRepo authrepos.UserRepository) adminservices.PendingUserReader {
	return &pendingUserAdapter{userRepo: userRepo}
}

func (a *pendingUserAdapter) GetByID(ctx context.Context, id uuid.UUID) (*adminservices.PendingUser, error) {
	user, err := a.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return &adminservices.PendingUser{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}
