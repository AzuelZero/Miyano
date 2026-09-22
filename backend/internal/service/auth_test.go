package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"miyano/internal/domain"
	"miyano/internal/ports"
)

// fakes: no database, no real crypto — the service is pure orchestration.

type fakeUserRepo struct {
	users  map[string]domain.User // email -> user
	tokens map[string]domain.RefreshToken
	nextID int
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[string]domain.User{}, tokens: map[string]domain.RefreshToken{}}
}

func (f *fakeUserRepo) CreateUser(_ context.Context, email, passwordHash, displayName string) (domain.User, error) {
	if _, exists := f.users[email]; exists {
		return domain.User{}, ports.ErrEmailTaken
	}
	f.nextID++
	u := domain.User{
		ID:           fmt.Sprintf("user-%d", f.nextID),
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
	}
	f.users[email] = u
	return u, nil
}

func (f *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (domain.User, error) {
	u, exists := f.users[email]
	if !exists {
		return domain.User{}, ports.ErrNotFound
	}
	return u, nil
}

func (f *fakeUserRepo) CreateRefreshToken(_ context.Context, userID, tokenHash string, expiresAt time.Time) error {
	f.tokens[tokenHash] = domain.RefreshToken{UserID: userID, ExpiresAt: expiresAt}
	return nil
}

func (f *fakeUserRepo) GetRefreshTokenByHash(_ context.Context, tokenHash string) (domain.RefreshToken, error) {
	t, exists := f.tokens[tokenHash]
	if !exists {
		return domain.RefreshToken{}, ports.ErrNotFound
	}
	return t, nil
}

func (f *fakeUserRepo) RevokeRefreshToken(_ context.Context, tokenHash string) error {
	t, exists := f.tokens[tokenHash]
	if !exists || t.RevokedAt != nil {
		return ports.ErrRefreshRejected
	}
	now := time.Now()
	t.RevokedAt = &now
	f.tokens[tokenHash] = t
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) { return "fake:" + password, nil }

func (fakeHasher) Verify(password, encoded string) (bool, error) {
	return encoded == "fake:"+password, nil
}

type fakeTokens struct{ seq int }

func (f *fakeTokens) Generate(userID, tokenType string) (string, time.Time, error) {
	f.seq++
	return fmt.Sprintf("%s-%s-%d", tokenType, userID, f.seq), time.Now().Add(time.Hour), nil
}

func (f *fakeTokens) Verify(token, wantType string) (string, error) {
	parts := strings.SplitN(token, "-", 3)
	if len(parts) != 3 || parts[0] != wantType {
		return "", errors.New("wrong token type")
	}
	return parts[1], nil
}

func newAuthService() (*AuthService, *fakeUserRepo, *fakeTokens) {
	repo := newFakeUserRepo()
	tokens := &fakeTokens{}
	return NewAuthService(repo, fakeHasher{}, tokens), repo, tokens
}

func TestRegisterOK(t *testing.T) {
	svc, repo, _ := newAuthService()

	tokens, err := svc.Register(context.Background(), "azuel@example.com", "password123", "Azuel")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("tokens = %+v, want both access and refresh", tokens)
	}
	user, err := repo.GetUserByEmail(context.Background(), "azuel@example.com")
	if err != nil {
		t.Fatalf("user not persisted: %v", err)
	}
	if user.PasswordHash != "fake:password123" {
		t.Errorf("stored hash = %q, want hashed password", user.PasswordHash)
	}
	if _, err := repo.GetRefreshTokenByHash(context.Background(), hashToken(tokens.RefreshToken)); err != nil {
		t.Errorf("refresh token not persisted: %v", err)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc, _, _ := newAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "dup@example.com", "password123", "One"); err != nil {
		t.Fatalf("first Register() error = %v", err)
	}
	if _, err := svc.Register(ctx, "dup@example.com", "password456", "Two"); !errors.Is(err, ports.ErrEmailTaken) {
		t.Errorf("second Register() error = %v, want ErrEmailTaken", err)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	svc, _, _ := newAuthService()

	if _, err := svc.Register(context.Background(), "weak@example.com", "short", "Weak"); !errors.Is(err, ports.ErrWeakPassword) {
		t.Errorf("Register(short) error = %v, want ErrWeakPassword", err)
	}
}

func TestLoginOK(t *testing.T) {
	svc, _, _ := newAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "login@example.com", "password123", "Azuel"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	tokens, err := svc.Login(ctx, "login@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("tokens = %+v, want both access and refresh", tokens)
	}
}

func TestLoginWrongPasswordAndUnknownEmailSameError(t *testing.T) {
	svc, _, _ := newAuthService()
	ctx := context.Background()

	if _, err := svc.Register(ctx, "known@example.com", "password123", "Azuel"); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	wrongPasswordErr := loginErr(t, svc, ctx, "known@example.com", "wrong-password")
	unknownEmailErr := loginErr(t, svc, ctx, "ghost@example.com", "password123")

	// Anti-enumeration: both failures must surface the SAME error.
	if wrongPasswordErr != unknownEmailErr {
		t.Errorf("wrong password error (%v) differs from unknown email error (%v)", wrongPasswordErr, unknownEmailErr)
	}
}

func loginErr(t *testing.T, svc *AuthService, ctx context.Context, email, password string) error {
	t.Helper()
	_, err := svc.Login(ctx, email, password)
	if !errors.Is(err, ports.ErrInvalidCredentials) {
		t.Errorf("Login(%s) error = %v, want ErrInvalidCredentials", email, err)
	}
	return err
}

func TestRefreshValidRotates(t *testing.T) {
	svc, repo, _ := newAuthService()
	ctx := context.Background()

	first, err := svc.Register(ctx, "refresh@example.com", "password123", "Azuel")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	second, err := svc.Refresh(ctx, first.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if second.AccessToken == "" || second.RefreshToken == "" {
		t.Errorf("tokens = %+v, want both access and refresh", second)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Error("refresh token was not rotated")
	}

	stored, err := repo.GetRefreshTokenByHash(ctx, hashToken(first.RefreshToken))
	if err != nil {
		t.Fatalf("old refresh token not found: %v", err)
	}
	if stored.RevokedAt == nil {
		t.Error("old refresh token still active after rotation, want revoked")
	}

	if _, err := repo.GetRefreshTokenByHash(ctx, hashToken(second.RefreshToken)); err != nil {
		t.Errorf("new refresh token not persisted: %v", err)
	}
}

func TestRefreshRevokedRejected(t *testing.T) {
	svc, _, _ := newAuthService()
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "revoked@example.com", "password123", "Azuel")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if _, err := svc.Refresh(ctx, tokens.RefreshToken); err != nil {
		t.Fatalf("first Refresh() error = %v", err)
	}

	// Re-using the now-revoked token must fail.
	if _, err := svc.Refresh(ctx, tokens.RefreshToken); !errors.Is(err, ports.ErrRefreshRejected) {
		t.Errorf("Refresh(revoked) error = %v, want ErrRefreshRejected", err)
	}
}

func TestRefreshExpiredRejected(t *testing.T) {
	svc, repo, tokens := newAuthService()
	ctx := context.Background()

	pair, err := svc.Register(ctx, "expired@example.com", "password123", "Azuel")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Force the stored row into the past.
	hash := hashToken(pair.RefreshToken)
	row := repo.tokens[hash]
	past := time.Now().Add(-time.Minute)
	row.ExpiresAt = past
	repo.tokens[hash] = row
	_ = tokens

	if _, err := svc.Refresh(ctx, pair.RefreshToken); !errors.Is(err, ports.ErrRefreshRejected) {
		t.Errorf("Refresh(expired) error = %v, want ErrRefreshRejected", err)
	}
}

func TestRefreshAccessTokenRejected(t *testing.T) {
	svc, _, _ := newAuthService()
	ctx := context.Background()

	pair, err := svc.Register(ctx, "types@example.com", "password123", "Azuel")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// An access token must not work as a refresh token.
	if _, err := svc.Refresh(ctx, pair.AccessToken); !errors.Is(err, ports.ErrRefreshRejected) {
		t.Errorf("Refresh(access token) error = %v, want ErrRefreshRejected", err)
	}
}
