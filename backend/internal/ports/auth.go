package ports

import (
	"context"
	"errors"
	"time"

	"miyano/internal/domain"
)

// Sentinel errors for the auth flow. Handlers map them to status codes;
// messages are generic on purpose (anti-enumeration, see add-auth design).
var (
	// ErrEmailTaken signals a duplicate email on registration (409 generic).
	ErrEmailTaken = errors.New("email already registered")
	// ErrInvalidCredentials signals a failed login for ANY reason (401,
	// identical whether the email exists or the password is wrong).
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrWeakPassword signals a password below the minimum length (400).
	ErrWeakPassword = errors.New("password too short")
	// ErrRefreshRejected signals an unknown, revoked or expired refresh
	// token (401, identical message for all three cases).
	ErrRefreshRejected = errors.New("refresh token rejected")
)

// Token types carried in the "typ" claim.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// AuthTokens is the token pair returned by register, login and refresh.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	TokenType    string // e.g. "Bearer"
}

// PasswordHasher hashes passwords for storage and verifies candidates.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (bool, error)
}

// TokenVerifier verifies JWTs of a given type and returns the subject.
type TokenVerifier interface {
	Verify(token, wantType string) (userID string, err error)
}

// TokenGenerator issues JWTs of a given type along with their expiry.
type TokenGenerator interface {
	TokenVerifier
	Generate(userID, tokenType string) (token string, expiresAt time.Time, err error)
}

// UserAuthRepository persists users and their refresh tokens.
type UserAuthRepository interface {
	// CreateUser inserts the user and its profile in one transaction.
	// Returns ErrEmailTaken when the email already exists.
	CreateUser(ctx context.Context, email, passwordHash, displayName string) (domain.User, error)

	// GetUserByEmail returns the user or ErrNotFound.
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)

	// CreateRefreshToken stores a refresh token hash with its expiry.
	CreateRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error

	// GetRefreshTokenByHash returns the token row or ErrNotFound.
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)

	// RevokeRefreshToken marks a still-active token as revoked (rotation).
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
}
