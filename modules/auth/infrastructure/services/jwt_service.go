package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type JWTService struct {
	secretKey          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	roleRepo           repositories.RoleRepository
	permRepo           repositories.PermissionRepository
	userRepo           repositories.UserRepository
	refreshTokenRepo   repositories.RefreshTokenRepository
}

func NewJWTService(
	secretKey string,
	accessTokenExpiry, refreshTokenExpiry time.Duration,
	roleRepo repositories.RoleRepository,
	permRepo repositories.PermissionRepository,
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
) *JWTService {
	return &JWTService{
		secretKey:          []byte(secretKey),
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		roleRepo:           roleRepo,
		permRepo:           permRepo,
		userRepo:           userRepo,
		refreshTokenRepo:   refreshTokenRepo,
	}
}

type jwtClaims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func (s *JWTService) GenerateAccessToken(ctx context.Context, userID uuid.UUID, email string) (string, error) {
	roles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user roles", zap.Error(err))
		return "", err
	}

	permissions, err := s.permRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user permissions", zap.Error(err))
		return "", err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	permNames := make([]string, len(permissions))
	for i, perm := range permissions {
		permNames[i] = perm.Name
	}

	claims := jwtClaims{
		UserID:      userID.String(),
		Email:       email,
		Roles:       roleNames,
		Permissions: permNames,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(ctx context.Context, tokenString string) (*services.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	return &services.TokenClaims{
		UserID:      userID,
		Email:       claims.Email,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}, nil
}

// GenerateTokenPairForUser issues an access+refresh token pair for the given user.
// Looks up the user's email via the UserRepository (required for the access-token claims)
// and returns the access-token lifetime in seconds. Does not persist the refresh token —
// callers are responsible for persistence (so this can also be used for cookie-only flows
// like the invitation accept handler, which mirrors the login persistence behavior).
func (s *JWTService) GenerateTokenPairForUser(ctx context.Context, userID uuid.UUID) (string, string, int64, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", 0, err
	}

	accessToken, err := s.GenerateAccessToken(ctx, userID, user.Email)
	if err != nil {
		return "", "", 0, err
	}

	refreshToken, err := s.GenerateRefreshToken(ctx, userID)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, int64(s.accessTokenExpiry.Seconds()), nil
}

// SaveRefreshToken persists a refresh token string for the given user. Provides a
// single source-of-truth so callers (the BFF session helper) don't need to wire up a
// RefreshTokenRepository directly or know how to construct the RefreshToken entity.
func (s *JWTService) SaveRefreshToken(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	expiresAt := time.Now().Add(s.refreshTokenExpiry)
	rt := entities.NewRefreshToken(userID, refreshToken, expiresAt)
	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		logger.Error("Failed to store refresh token", zap.Error(err))
		return err
	}
	return nil
}

func (s *JWTService) GetTokenExpiration() time.Duration {
	return s.accessTokenExpiry
}

func (s *JWTService) GetRefreshTokenExpiration() time.Duration {
	return s.refreshTokenExpiry
}
