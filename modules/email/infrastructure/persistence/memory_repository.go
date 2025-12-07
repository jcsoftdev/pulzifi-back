package persistence

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/email/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/email/domain/errors"
)

// MemoryEmailRepository implements EmailRepository using in-memory storage
type MemoryEmailRepository struct {
	emails map[uuid.UUID]*entities.Email
	mu     sync.RWMutex
}

// NewMemoryEmailRepository creates a new memory email repository
func NewMemoryEmailRepository() *MemoryEmailRepository {
	return &MemoryEmailRepository{
		emails: make(map[uuid.UUID]*entities.Email),
	}
}

// Save saves an email
func (r *MemoryEmailRepository) Save(ctx context.Context, email *entities.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if email.ID == uuid.Nil {
		return &domainerrors.InvalidEmailError{Message: "email id cannot be nil"}
	}

	r.emails[email.ID] = email
	return nil
}

// GetByID retrieves an email by ID
func (r *MemoryEmailRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	email, ok := r.emails[id]
	if !ok {
		return nil, &domainerrors.EmailNotFoundError{EmailID: id.String()}
	}

	return email, nil
}

// GetByTo retrieves emails by recipient
func (r *MemoryEmailRepository) GetByTo(ctx context.Context, to string, limit int) ([]*entities.Email, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var emails []*entities.Email
	for _, email := range r.emails {
		if email.To == to {
			emails = append(emails, email)
			if len(emails) >= limit {
				break
			}
		}
	}

	return emails, nil
}

// Update updates an email
func (r *MemoryEmailRepository) Update(ctx context.Context, email *entities.Email) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.emails[email.ID]; !ok {
		return &domainerrors.EmailNotFoundError{EmailID: email.ID.String()}
	}

	r.emails[email.ID] = email
	return nil
}
