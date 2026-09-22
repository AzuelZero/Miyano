package jwttoken

import (
	"errors"
	"testing"
	"time"

	"miyano/internal/ports"
)

const testSecret = "test-secret-that-is-long-enough-32ch"

func newTestGenerator() *Generator {
	return NewGenerator(testSecret, 15*time.Minute, 7*24*time.Hour)
}

func TestGenerateVerifyRoundTrip(t *testing.T) {
	g := newTestGenerator()

	token, expiresAt, err := g.Generate("user-123", ports.TokenTypeAccess)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !time.Now().Before(expiresAt) {
		t.Error("expiresAt is not in the future")
	}

	userID, err := g.Verify(token, ports.TokenTypeAccess)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if userID != "user-123" {
		t.Errorf("userID = %q, want user-123", userID)
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	g := NewGenerator(testSecret, -time.Minute, -time.Minute)

	token, _, err := g.Generate("user-123", ports.TokenTypeAccess)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := g.Verify(token, ports.TokenTypeAccess); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify() error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyRejectsWrongType(t *testing.T) {
	g := newTestGenerator()

	refresh, _, err := g.Generate("user-123", ports.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// A refresh token must not grant access.
	if _, err := g.Verify(refresh, ports.TokenTypeAccess); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify(refresh, access) error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	issuer := newTestGenerator()
	verifier := NewGenerator("another-secret-also-long-enough-32ch", 15*time.Minute, 7*24*time.Hour)

	token, _, err := issuer.Generate("user-123", ports.TokenTypeAccess)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := verifier.Verify(token, ports.TokenTypeAccess); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify() with different secret error = %v, want ErrInvalidToken", err)
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	g := newTestGenerator()

	if _, err := g.Verify("not-a-jwt", ports.TokenTypeAccess); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Verify(garbage) error = %v, want ErrInvalidToken", err)
	}
}

func TestTokensUniqueWithinSameSecond(t *testing.T) {
	g := newTestGenerator()

	first, _, err := g.Generate("user-123", ports.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, _, err := g.Generate("user-123", ports.TokenTypeRefresh)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if first == second {
		t.Error("two refresh tokens for the same user minted back-to-back are identical, want unique (jti)")
	}
}
