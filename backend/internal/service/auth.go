package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"miyano/internal/ports"
)

const minPasswordLength = 8

// AuthService serves registration, login and refresh-token rotation.
type AuthService struct {
	users  ports.UserAuthRepository
	hasher ports.PasswordHasher
	tokens ports.TokenGenerator
}

// NewAuthService wires the service with its ports.
func NewAuthService(users ports.UserAuthRepository, hasher ports.PasswordHasher, tokens ports.TokenGenerator) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens}
}

// Register validates the input, hashes the password, persists the user with
// its profile and issues the first token pair.
func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (ports.AuthTokens, error) {
	if len(password) < minPasswordLength {
		return ports.AuthTokens{}, ports.ErrWeakPassword
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return ports.AuthTokens{}, fmt.Errorf("hashing password: %w", err)
	}

	user, err := s.users.CreateUser(ctx, email, hash, displayName)
	if err != nil {
		return ports.AuthTokens{}, err
	}
	return s.issueTokens(ctx, user.ID)
}

// Login verifies the credentials. Every failure (unknown email, wrong
// password, unusable hash) maps to ErrInvalidCredentials so responses do
// not reveal whether the email exists.
func (s *AuthService) Login(ctx context.Context, email, password string) (ports.AuthTokens, error) {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ports.AuthTokens{}, ports.ErrInvalidCredentials
		}
		return ports.AuthTokens{}, err
	}

	ok, err := s.hasher.Verify(password, user.PasswordHash)
	if err != nil || !ok {
		return ports.AuthTokens{}, ports.ErrInvalidCredentials
	}
	return s.issueTokens(ctx, user.ID)
}

// Refresh rotates a valid refresh token: the old one is revoked and a fresh
// access+refresh pair is issued. Unknown, revoked, expired or mistyped
// tokens all map to ErrRefreshRejected.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (ports.AuthTokens, error) {
	userID, err := s.tokens.Verify(refreshToken, ports.TokenTypeRefresh)
	if err != nil {
		return ports.AuthTokens{}, ports.ErrRefreshRejected
	}

	hash := hashToken(refreshToken)
	stored, err := s.users.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return ports.AuthTokens{}, ports.ErrRefreshRejected
		}
		return ports.AuthTokens{}, err
	}
	if stored.RevokedAt != nil || !time.Now().Before(stored.ExpiresAt) {
		return ports.AuthTokens{}, ports.ErrRefreshRejected
	}

	if err := s.users.RevokeRefreshToken(ctx, hash); err != nil {
		if errors.Is(err, ports.ErrRefreshRejected) {
			return ports.AuthTokens{}, ports.ErrRefreshRejected
		}
		return ports.AuthTokens{}, err
	}
	return s.issueTokens(ctx, userID)
}

func (s *AuthService) issueTokens(ctx context.Context, userID string) (ports.AuthTokens, error) {
	access, _, err := s.tokens.Generate(userID, ports.TokenTypeAccess)
	if err != nil {
		return ports.AuthTokens{}, fmt.Errorf("issuing access token: %w", err)
	}
	refresh, refreshExpiresAt, err := s.tokens.Generate(userID, ports.TokenTypeRefresh)
	if err != nil {
		return ports.AuthTokens{}, fmt.Errorf("issuing refresh token: %w", err)
	}
	if err := s.users.CreateRefreshToken(ctx, userID, hashToken(refresh), refreshExpiresAt); err != nil {
		return ports.AuthTokens{}, err
	}

	return ports.AuthTokens{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
	}, nil
}

// hashToken renders the hex SHA-256 of a token. Refresh tokens are stored
// hashed so a database leak does not yield usable tokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
