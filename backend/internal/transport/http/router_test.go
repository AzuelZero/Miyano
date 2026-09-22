package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"miyano/internal/adapter/jwttoken"
	"miyano/internal/domain"
	"miyano/internal/ports"
	"miyano/internal/service"
	"miyano/internal/transport/http/handler"
)

const testRouterSecret = "router-test-secret-long-enough!!"

// fakeRepo is an in-memory ExerciseRepository for router tests.
type fakeRepo struct {
	exercises []domain.Exercise
}

func (f *fakeRepo) List(_ context.Context, equipment string) ([]domain.Exercise, error) {
	out := make([]domain.Exercise, 0, len(f.exercises))
	for _, ex := range f.exercises {
		if equipment == "" || ex.Equipment == equipment {
			out = append(out, ex)
		}
	}
	return out, nil
}

func (f *fakeRepo) GetByID(_ context.Context, id string) (domain.Exercise, []domain.Translation, error) {
	for _, ex := range f.exercises {
		if ex.ID == id {
			return ex, nil, nil
		}
	}
	return domain.Exercise{}, nil, ports.ErrNotFound
}

// fakeUserRepo is an in-memory UserAuthRepository (register/login/refresh).
type fakeUserRepo struct {
	users  map[string]string // email -> passwordHash
	tokens map[string]bool   // token hash -> revoked
	nextID int
}

func (f *fakeUserRepo) CreateUser(_ context.Context, email, passwordHash, _ string) (domain.User, error) {
	if _, exists := f.users[email]; exists {
		return domain.User{}, ports.ErrEmailTaken
	}
	f.nextID++
	f.users[email] = passwordHash
	return domain.User{ID: strconv.Itoa(f.nextID), Email: email, PasswordHash: passwordHash}, nil
}

func (f *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (domain.User, error) {
	hash, exists := f.users[email]
	if !exists {
		return domain.User{}, ports.ErrNotFound
	}
	return domain.User{ID: email, Email: email, PasswordHash: hash}, nil
}

func (f *fakeUserRepo) CreateRefreshToken(_ context.Context, _, tokenHash string, _ time.Time) error {
	f.tokens[tokenHash] = false
	return nil
}

func (f *fakeUserRepo) GetRefreshTokenByHash(_ context.Context, tokenHash string) (domain.RefreshToken, error) {
	if _, exists := f.tokens[tokenHash]; !exists {
		return domain.RefreshToken{}, ports.ErrNotFound
	}
	return domain.RefreshToken{UserID: "1", ExpiresAt: time.Now().Add(time.Hour)}, nil
}

func (f *fakeUserRepo) RevokeRefreshToken(_ context.Context, tokenHash string) error {
	if revoked, exists := f.tokens[tokenHash]; !exists || revoked {
		return ports.ErrRefreshRejected
	}
	f.tokens[tokenHash] = true
	return nil
}

// fakeHasher keeps router tests fast: no real argon2 derivation.
type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) { return "fake:" + password, nil }
func (fakeHasher) Verify(password, encoded string) (bool, error) {
	return encoded == "fake:"+password, nil
}

// newTestRouter wires the full route tree with fast fakes and the REAL
// token generator, so the middleware exercises genuine HS256 verification.
func newTestRouter(exercises ...domain.Exercise) http.Handler {
	tokens := jwttoken.NewGenerator(testRouterSecret, 15*time.Minute, 7*24*time.Hour)
	authSvc := service.NewAuthService(&fakeUserRepo{
		users:  map[string]string{},
		tokens: map[string]bool{},
	}, fakeHasher{}, tokens)
	exerciseSvc := service.NewExerciseService(&fakeRepo{exercises: exercises})
	return NewRouter(Handlers{
		Health:    handler.Health(),
		Exercises: handler.NewExerciseHandler(exerciseSvc),
		Auth:      handler.NewAuthHandler(authSvc),
		Verifier:  tokens,
	})
}

// mintAccessToken issues a genuine access token for the protected group.
func mintAccessToken(t *testing.T) string {
	t.Helper()
	tokens := jwttoken.NewGenerator(testRouterSecret, 15*time.Minute, 7*24*time.Hour)
	token, _, err := tokens.Generate("user-1", ports.TokenTypeAccess)
	if err != nil {
		t.Fatalf("minting access token: %v", err)
	}
	return token
}

func TestRouterHealthEndpoint(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestRouterUnknownPathReturns404(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/definitely-not-a-route")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRouterHealthRejectsPost(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/v1/health", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestRouterExercisesRequireToken(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(
		domain.Exercise{ID: "0001", Name: "Push-up", Category: "Chest", BodyPart: "Chest", Equipment: "Body Weight"},
	))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/exercises")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (catalog is protected)", resp.StatusCode, http.StatusUnauthorized)
	}
	var body map[string]string
	if err := jsonDecode(resp.Body, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] != "missing or invalid token" {
		t.Errorf("error = %q, want missing or invalid token", body["error"])
	}
}

func TestRouterExercisesWithValidToken(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(
		domain.Exercise{ID: "0001", Name: "Push-up", Category: "Chest", BodyPart: "Chest", Equipment: "Body Weight"},
	))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/exercises", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+mintAccessToken(t))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var body []map[string]any
	if err := jsonDecode(resp.Body, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body) != 1 || body[0]["id"] != "0001" {
		t.Errorf("body = %v, want one exercise 0001", body)
	}
}

func TestRouterExercisesRejectGarbageToken(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/exercises", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer not-a-real-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestRouterRegisterAndDuplicate(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	payload := func() *bytes.Reader {
		b, _ := json.Marshal(map[string]string{
			"email":        "router@example.com",
			"password":     "password123",
			"display_name": "Router",
		})
		return bytes.NewReader(b)
	}

	first, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", payload())
	if err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	defer first.Body.Close()
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first register status = %d, want %d", first.StatusCode, http.StatusCreated)
	}
	var tokens map[string]string
	if err := jsonDecode(first.Body, &tokens); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if tokens["access_token"] == "" || tokens["refresh_token"] == "" {
		t.Errorf("tokens = %v, want both access and refresh", tokens)
	}

	second, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", payload())
	if err != nil {
		t.Fatalf("second register failed: %v", err)
	}
	defer second.Body.Close()
	if second.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register status = %d, want %d", second.StatusCode, http.StatusConflict)
	}
	var body map[string]string
	if err := jsonDecode(second.Body, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] != "could not create account" {
		t.Errorf("error = %q, want generic could not create account", body["error"])
	}
}

func TestRouterRegisterWeakPassword(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	b, _ := json.Marshal(map[string]string{"email": "weak@example.com", "password": "short", "display_name": "W"})
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRouterLoginInvalidCredentials(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	b, _ := json.Marshal(map[string]string{"email": "ghost@example.com", "password": "password123"})
	resp, err := http.Post(srv.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	var body map[string]string
	if err := jsonDecode(resp.Body, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["error"] != "invalid credentials" {
		t.Errorf("error = %q, want invalid credentials", body["error"])
	}
}

func TestRouterAuthRateLimit(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	b, _ := json.Marshal(map[string]string{"email": "ghost@example.com", "password": "password123"})
	url := srv.URL + "/api/v1/auth/login"

	var last int
	for i := 0; i < 6; i++ {
		resp, err := http.Post(url, "application/json", bytes.NewReader(b))
		if err != nil {
			t.Fatalf("login %d failed: %v", i+1, err)
		}
		resp.Body.Close()
		last = resp.StatusCode
	}

	if last != http.StatusTooManyRequests {
		t.Errorf("6th login status = %d, want %d (5/min rate limit)", last, http.StatusTooManyRequests)
	}
}
