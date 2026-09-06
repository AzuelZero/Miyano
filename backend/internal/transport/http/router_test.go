package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"miyano/internal/domain"
	"miyano/internal/ports"
	"miyano/internal/service"
	"miyano/internal/transport/http/handler"
)

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

func newTestRouter(exercises ...domain.Exercise) http.Handler {
	svc := service.NewExerciseService(&fakeRepo{exercises: exercises})
	return NewRouter(Handlers{
		Health:    handler.Health(),
		Exercises: handler.NewExerciseHandler(svc),
	})
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

func TestRouterExercisesList(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(
		domain.Exercise{ID: "0001", Name: "Push-up", Category: "Chest", BodyPart: "Chest", Equipment: "Body Weight"},
	))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/exercises")
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

func TestRouterExercisesFilter(t *testing.T) {
	srv := httptest.NewServer(newTestRouter(
		domain.Exercise{ID: "0001", Name: "Push-up", Category: "Chest", BodyPart: "Chest", Equipment: "Body Weight"},
		domain.Exercise{ID: "0002", Name: "Bench", Category: "Chest", BodyPart: "Chest", Equipment: "Barbell"},
	))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/exercises?equipment=Barbell")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var body []map[string]any
	if err := jsonDecode(resp.Body, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body) != 1 || body[0]["equipment"] != "Barbell" {
		t.Errorf("body = %v, want only Barbell", body)
	}
}

func TestRouterExerciseDetail404(t *testing.T) {
	srv := httptest.NewServer(newTestRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/exercises/9999")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
