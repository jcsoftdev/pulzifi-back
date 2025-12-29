package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
}

func NewJWTService(
	secretKey string,
	accessTokenExpiry, refreshTokenExpiry time.Duration,
	roleRepo repositories.RoleRepository,
	permRepo repositories.PermissionRepository,
) *JWTService {
	return &JWTService{
		secretKey:          []byte(secretKey),
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		roleRepo:           roleRepo,
		permRepo:           permRepo,
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

func (s *JWTService) GetTokenExpiration() time.Duration {
	return s.accessTokenExpiry
}

func (s *JWTService) GetRefreshTokenExpiration() time.Time {
	return time.Now().Add(s.refreshTokenExpiry)
}
