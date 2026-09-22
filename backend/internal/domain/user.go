package domain

import "time"

// User is a registered account with its stored password hash.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
}

// RefreshToken is a persisted refresh token (stored hashed) with its state.
type RefreshToken struct {
	UserID    string
	ExpiresAt time.Time
	RevokedAt *time.Time
}
