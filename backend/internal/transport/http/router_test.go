package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

// fakeRoutineRepo is an in-memory RoutineRepository (routines capability).
type fakeRoutineRepo struct {
	byUser    map[string]map[string]domain.Routine
	knownExes map[string]bool
	nextID    int
}

func (f *fakeRoutineRepo) Create(_ context.Context, userID, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error) {
	for _, ex := range exercises {
		if !f.knownExes[ex.ExerciseID] {
			return domain.Routine{}, fmt.Errorf("%w: unknown exercise %q", ports.ErrInvalidInput, ex.ExerciseID)
		}
	}
	f.nextID++
	id := fmt.Sprintf("rt-%d", f.nextID)
	stored := make([]domain.RoutineExercise, 0, len(exercises))
	for i, ex := range exercises {
		ex.Position = i + 1
		stored = append(stored, ex)
	}
	r := domain.Routine{ID: id, UserID: userID, Name: name, Description: description, Exercises: stored, Tags: tags}
	if f.byUser[userID] == nil {
		f.byUser[userID] = map[string]domain.Routine{}
	}
	f.byUser[userID][id] = r
	return r, nil
}

func (f *fakeRoutineRepo) Update(_ context.Context, userID, routineID, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error) {
	r, ok := f.byUser[userID][routineID]
	if !ok {
		return domain.Routine{}, ports.ErrNotFound
	}
	r.Name, r.Description, r.Tags = name, description, tags
	r.Exercises = append([]domain.RoutineExercise(nil), exercises...)
	f.byUser[userID][routineID] = r
	return r, nil
}

func (f *fakeRoutineRepo) Get(_ context.Context, userID, routineID string) (domain.Routine, error) {
	r, ok := f.byUser[userID][routineID]
	if !ok {
		return domain.Routine{}, ports.ErrNotFound
	}
	return r, nil
}

func (f *fakeRoutineRepo) List(_ context.Context, userID, tag string) ([]domain.Routine, error) {
	var out []domain.Routine
	for _, r := range f.byUser[userID] {
		has := false
		for _, t := range r.Tags {
			if t == tag {
				has = true
			}
		}
		if tag == "" || has {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRoutineRepo) Delete(_ context.Context, userID, routineID string) error {
	if _, ok := f.byUser[userID][routineID]; !ok {
		return ports.ErrNotFound
	}
	delete(f.byUser[userID], routineID)
	return nil
}

// fakeLevelRepo is an in-memory FitnessLevelRepository.
type fakeLevelRepo struct{ users map[string]string }

func (f *fakeLevelRepo) GetByID(_ context.Context, id string) (domain.FitnessLevel, error) {
	levels := map[string]float64{"beginner": 0.70, "intermediate": 1.00, "advanced": 1.30}
	factor, ok := levels[id]
	if !ok {
		return domain.FitnessLevel{}, ports.ErrNotFound
	}
	return domain.FitnessLevel{ID: id, ScaleFactor: factor}, nil
}

func (f *fakeLevelRepo) GetUserLevel(_ context.Context, userID string) (domain.FitnessLevel, error) {
	id, ok := f.users[userID]
	if !ok {
		return domain.FitnessLevel{}, ports.ErrNotFound
	}
	return f.GetByID(context.Background(), id)
}

func (f *fakeLevelRepo) SetUserLevel(_ context.Context, userID, levelID string) error {
	if _, err := f.GetByID(context.Background(), levelID); err != nil {
		return err
	}
	f.users[userID] = levelID
	return nil
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
	routineSvc := service.NewRoutineService(
		&fakeRoutineRepo{byUser: map[string]map[string]domain.Routine{}, knownExes: map[string]bool{"0001": true, "0002": true}},
		&fakeLevelRepo{users: map[string]string{}},
	)
	return NewRouter(Handlers{
		Health:    handler.Health(),
		Exercises: handler.NewExerciseHandler(exerciseSvc),
		Auth:      handler.NewAuthHandler(authSvc),
		Routines:  handler.NewRoutineHandler(routineSvc),
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

// --- routines capability (add-routines) ---

// registerUser creates a user through the API and returns its access token.
func registerUser(t *testing.T, srvURL, email string) string {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"email": email, "password": "password123", "display_name": "T"})
	resp, err := http.Post(srvURL+"/api/v1/auth/register", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: status %d", email, resp.StatusCode)
	}
	var tokens map[string]string
	if err := jsonDecode(resp.Body, &tokens); err != nil {
		t.Fatalf("register %s: %v", email, err)
	}
	return tokens["access_token"]
}

func authedJSON(t *testing.T, method, url, token string, body any) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func TestRouterRoutinesRequireToken(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/routines")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestRouterRoutineCreateAndGetScaled(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()
	token := registerUser(t, srv.URL, "rt-owner@example.com")

	create := authedJSON(t, "POST", srv.URL+"/api/v1/routines", token, map[string]any{
		"name": "Leg day",
		"exercises": []map[string]any{
			{"exercise_id": "0001", "target_sets": 3, "target_reps": 10, "target_duration_s": 60, "rest_s": 90},
			{"exercise_id": "0002", "target_sets": 2, "target_reps": 8},
		},
		"tags": []string{"pierna", "casa"},
	})
	defer create.Body.Close()
	if create.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", create.StatusCode)
	}
	var detail map[string]any
	if err := jsonDecode(create.Body, &detail); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if detail["name"] != "Leg day" {
		t.Errorf("name = %v", detail["name"])
	}
	exercises := detail["exercises"].([]any)
	if len(exercises) != 2 {
		t.Fatalf("exercises = %d, want 2 (ordered by position)", len(exercises))
	}
	first := exercises[0].(map[string]any)
	if first["position"].(float64) != 1 {
		t.Errorf("first position = %v, want 1 (normalized from slice order)", first["position"])
	}
	// Default level is beginner x0.70: 10 reps -> 7.
	scaled := detail["scaled_exercises"].([]any)[0].(map[string]any)
	if scaled["target_reps"].(float64) != 7 {
		t.Errorf("scaled reps = %v, want 7", scaled["target_reps"])
	}
	if detail["total_time_s"].(float64) != 450 { // ex1: 3x60+2x90=360 · ex2: 2x45+1x0=90
		t.Errorf("total_time_s = %v, want 450", detail["total_time_s"])
	}

	// GET by id returns the same shape.
	id := detail["id"].(string)
	get := authedJSON(t, "GET", srv.URL+"/api/v1/routines/"+id, token, nil)
	defer get.Body.Close()
	if get.StatusCode != http.StatusOK {
		t.Fatalf("get status = %d, want 200", get.StatusCode)
	}
}

func TestRouterRoutineCreateUnknownExercise(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()
	token := registerUser(t, srv.URL, "rt-ghost@example.com")

	resp := authedJSON(t, "POST", srv.URL+"/api/v1/routines", token, map[string]any{
		"name":      "Bad",
		"exercises": []map[string]any{{"exercise_id": "9999"}},
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
	var body map[string]string
	_ = jsonDecode(resp.Body, &body)
	if !strings.Contains(body["error"], "9999") {
		t.Errorf("error = %q, want it to name the invalid id", body["error"])
	}
}

func TestRouterRoutineIsolation(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()
	owner := registerUser(t, srv.URL, "rt-a@example.com")
	intruder := registerUser(t, srv.URL, "rt-b@example.com")

	create := authedJSON(t, "POST", srv.URL+"/api/v1/routines", owner, map[string]any{
		"name": "Mine", "exercises": []map[string]any{{"exercise_id": "0001"}},
	})
	defer create.Body.Close()
	var detail map[string]any
	_ = jsonDecode(create.Body, &detail)
	id := detail["id"].(string)

	resp := authedJSON(t, "GET", srv.URL+"/api/v1/routines/"+id, intruder, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("foreign routine status = %d, want 404 (not 403)", resp.StatusCode)
	}
}

func TestRouterRoutineUpdateDeleteAndList(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()
	token := registerUser(t, srv.URL, "rt-crud@example.com")

	create := authedJSON(t, "POST", srv.URL+"/api/v1/routines", token, map[string]any{
		"name": "V1", "tags": []string{"old"}, "exercises": []map[string]any{{"exercise_id": "0001"}},
	})
	var detail map[string]any
	_ = jsonDecode(create.Body, &detail)
	create.Body.Close()
	id := detail["id"].(string)

	update := authedJSON(t, "PUT", srv.URL+"/api/v1/routines/"+id, token, map[string]any{
		"name": "V2", "tags": []string{"pierna"}, "exercises": []map[string]any{{"exercise_id": "0002"}},
	})
	var updated map[string]any
	_ = jsonDecode(update.Body, &updated)
	update.Body.Close()
	if updated["name"] != "V2" {
		t.Errorf("after PUT name = %v, want V2", updated["name"])
	}

	list := authedJSON(t, "GET", srv.URL+"/api/v1/routines?tag=pierna", token, nil)
	var items []map[string]any
	_ = jsonDecode(list.Body, &items)
	list.Body.Close()
	if len(items) != 1 || items[0]["name"] != "V2" {
		t.Errorf("list?tag=pierna = %v, want only V2", items)
	}

	del := authedJSON(t, "DELETE", srv.URL+"/api/v1/routines/"+id, token, nil)
	del.Body.Close()
	if del.StatusCode != http.StatusNoContent {
		t.Errorf("delete status = %d, want 204", del.StatusCode)
	}
	gone := authedJSON(t, "GET", srv.URL+"/api/v1/routines/"+id, token, nil)
	gone.Body.Close()
	if gone.StatusCode != http.StatusNotFound {
		t.Errorf("after delete status = %d, want 404", gone.StatusCode)
	}
}

func TestRouterFitnessLevelDefaultAndSet(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()
	token := registerUser(t, srv.URL, "rt-level@example.com")

	first := authedJSON(t, "GET", srv.URL+"/api/v1/me/fitness-level", token, nil)
	var level map[string]any
	_ = jsonDecode(first.Body, &level)
	first.Body.Close()
	if level["id"] != "beginner" || level["scale_factor"].(float64) != 0.70 {
		t.Errorf("default level = %v, want beginner x0.70", level)
	}

	set := authedJSON(t, "PUT", srv.URL+"/api/v1/me/fitness-level", token, map[string]string{"level_id": "advanced"})
	var after map[string]any
	_ = jsonDecode(set.Body, &after)
	set.Body.Close()
	if after["scale_factor"].(float64) != 1.30 {
		t.Errorf("after set = %v, want x1.30", after)
	}

	bad := authedJSON(t, "PUT", srv.URL+"/api/v1/me/fitness-level", token, map[string]string{"level_id": "gigachad"})
	bad.Body.Close()
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("unknown level status = %d, want 400", bad.StatusCode)
	}
}
