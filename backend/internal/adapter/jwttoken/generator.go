// Package jwttoken implements the ports.TokenGenerator port with golang-jwt
// v5: HS256 with a single secret, access and refresh tokens distinguished by
// the custom "typ" claim.
package jwttoken

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"miyano/internal/ports"
)

// ErrInvalidToken is returned when a token fails verification.
var ErrInvalidToken = errors.New("invalid token")

// claims carries the registered claims plus the custom typ so one secret can
// serve both token kinds without them being interchangeable. The jti (ID)
// keeps tokens unique even when minted within the same second.
type claims struct {
	Typ string `json:"typ"`
	jwt.RegisteredClaims
}

// Generator issues and verifies JWTs (HS256).
type Generator struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewGenerator builds the generator. Secrets must come from config
// (JWT_SECRET, >= 32 chars); TTLs are decided by the caller (wiring).
func NewGenerator(secret string, accessTTL, refreshTTL time.Duration) *Generator {
	return &Generator{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// Generate returns a signed token of the given type along with its expiry.
func (g *Generator) Generate(userID, tokenType string) (string, time.Time, error) {
	ttl := g.ttl(tokenType)

	id, err := randomTokenID()
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now()
	expiresAt := now.Add(ttl)
	c := claims{
		Typ: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        id,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(g.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing %s token: %w", tokenType, err)
	}
	return signed, expiresAt, nil
}

// Verify validates the signature, algorithm, expiry and token type, then
// returns the subject (user id).
func (g *Generator) Verify(token, wantType string) (string, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&claims{},
		func(t *jwt.Token) (any, error) { return g.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", ErrInvalidToken
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid || c.Subject == "" || c.Typ != wantType {
		return "", ErrInvalidToken
	}
	return c.Subject, nil
}

func (g *Generator) ttl(tokenType string) time.Duration {
	switch tokenType {
	case ports.TokenTypeAccess:
		return g.accessTTL
	case ports.TokenTypeRefresh:
		return g.refreshTTL
	default:
		panic(fmt.Sprintf("jwttoken: unknown token type %q", tokenType))
	}
}

// randomTokenID mints the jti claim: without it, two tokens for the same
// user minted within the same second would be identical strings.
func randomTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
