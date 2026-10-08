package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"miyano/internal/domain"
	"miyano/internal/ports"
	"miyano/internal/service"
	"miyano/internal/transport/http/middleware"
)

// RoutineHandler serves the routine and fitness-level endpoints.
type RoutineHandler struct {
	svc *service.RoutineService
}

// NewRoutineHandler wires the handler with its service.
func NewRoutineHandler(svc *service.RoutineService) *RoutineHandler {
	return &RoutineHandler{svc: svc}
}

type routineExerciseRequest struct {
	ExerciseID      string `json:"exercise_id"`
	TargetSets      *int   `json:"target_sets,omitempty"`
	TargetReps      *int   `json:"target_reps,omitempty"`
	TargetDurationS *int   `json:"target_duration_s,omitempty"`
	RestS           *int   `json:"rest_s,omitempty"`
}

type routineRequest struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	Exercises   []routineExerciseRequest `json:"exercises"`
	Tags        []string                 `json:"tags"`
}

type routineExerciseDTO struct {
	ExerciseID      string `json:"exercise_id"`
	Position        int    `json:"position"`
	TargetSets      *int   `json:"target_sets,omitempty"`
	TargetReps      *int   `json:"target_reps,omitempty"`
	TargetDurationS *int   `json:"target_duration_s,omitempty"`
	RestS           *int   `json:"rest_s,omitempty"`
}

type fitnessLevelDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ScaleFactor float64 `json:"scale_factor"`
}

type routineDetailDTO struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	Description      *string              `json:"description,omitempty"`
	Tags             []string             `json:"tags"`
	Exercises        []routineExerciseDTO `json:"exercises"`
	Level            fitnessLevelDTO      `json:"level"`
	ScaledExercises  []routineExerciseDTO `json:"scaled_exercises"`
	TotalTimeS       int                  `json:"total_time_s"`
	TotalTimeScaledS int                  `json:"total_time_scaled_s"`
}

type routineSummaryDTO struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   *string  `json:"description,omitempty"`
	Tags          []string `json:"tags"`
	ExerciseCount int      `json:"exercise_count"`
	TotalTimeS    int      `json:"total_time_s"`
}

func toRoutineInput(req routineRequest) service.RoutineInput {
	exercises := make([]domain.RoutineExercise, 0, len(req.Exercises))
	for _, ex := range req.Exercises {
		exercises = append(exercises, domain.RoutineExercise{
			ExerciseID:      ex.ExerciseID,
			TargetSets:      ex.TargetSets,
			TargetReps:      ex.TargetReps,
			TargetDurationS: ex.TargetDurationS,
			RestS:           ex.RestS,
		})
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	return service.RoutineInput{Name: req.Name, Description: req.Description, Exercises: exercises, Tags: tags}
}

func toExerciseDTOs(exercises []domain.RoutineExercise) []routineExerciseDTO {
	out := make([]routineExerciseDTO, 0, len(exercises))
	for _, ex := range exercises {
		out = append(out, routineExerciseDTO{
			ExerciseID:      ex.ExerciseID,
			Position:        ex.Position,
			TargetSets:      ex.TargetSets,
			TargetReps:      ex.TargetReps,
			TargetDurationS: ex.TargetDurationS,
			RestS:           ex.RestS,
		})
	}
	return out
}

func toRoutineDetailDTO(d service.RoutineDetail) routineDetailDTO {
	return routineDetailDTO{
		ID:               d.Routine.ID,
		Name:             d.Routine.Name,
		Description:      d.Routine.Description,
		Tags:             d.Routine.Tags,
		Exercises:        toExerciseDTOs(d.Routine.Exercises),
		Level:            fitnessLevelDTO{ID: d.Level.ID, Name: d.Level.Name, ScaleFactor: d.Level.ScaleFactor},
		ScaledExercises:  toExerciseDTOs(d.ScaledExercises),
		TotalTimeS:       d.TotalTimeS,
		TotalTimeScaledS: d.TotalTimeScaledS,
	}
}

// Create handles POST /api/v1/routines.
func (h *RoutineHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeRoutineRequest(w, r)
	if !ok {
		return
	}
	detail, err := h.svc.Create(r.Context(), middleware.UserID(r), toRoutineInput(req))
	if !writeRoutineResult(w, err, "creating routine") {
		return
	}
	writeJSON(w, http.StatusCreated, toRoutineDetailDTO(detail))
}

// List handles GET /api/v1/routines (optional ?tag= exact filter).
func (h *RoutineHandler) List(w http.ResponseWriter, r *http.Request) {
	summaries, err := h.svc.List(r.Context(), middleware.UserID(r), r.URL.Query().Get("tag"))
	if err != nil {
		log.Printf("listing routines: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]routineSummaryDTO, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, routineSummaryDTO{
			ID:            s.ID,
			Name:          s.Name,
			Description:   s.Description,
			Tags:          s.Tags,
			ExerciseCount: s.ExerciseCount,
			TotalTimeS:    s.TotalTimeS,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// Get handles GET /api/v1/routines/{id} (base + scaled variant).
func (h *RoutineHandler) Get(w http.ResponseWriter, r *http.Request) {
	detail, err := h.svc.Get(r.Context(), middleware.UserID(r), chi.URLParam(r, "id"))
	if !writeRoutineResult(w, err, "getting routine") {
		return
	}
	writeJSON(w, http.StatusOK, toRoutineDetailDTO(detail))
}

// Update handles PUT /api/v1/routines/{id} (full replacement).
func (h *RoutineHandler) Update(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeRoutineRequest(w, r)
	if !ok {
		return
	}
	detail, err := h.svc.Update(r.Context(), middleware.UserID(r), chi.URLParam(r, "id"), toRoutineInput(req))
	if !writeRoutineResult(w, err, "updating routine") {
		return
	}
	writeJSON(w, http.StatusOK, toRoutineDetailDTO(detail))
}

// Delete handles DELETE /api/v1/routines/{id}.
func (h *RoutineHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), middleware.UserID(r), chi.URLParam(r, "id")); !writeRoutineResult(w, err, "deleting routine") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetLevel handles GET /api/v1/me/fitness-level.
func (h *RoutineHandler) GetLevel(w http.ResponseWriter, r *http.Request) {
	level, err := h.svc.GetLevel(r.Context(), middleware.UserID(r))
	if err != nil {
		log.Printf("getting fitness level: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, fitnessLevelDTO{ID: level.ID, Name: level.Name, ScaleFactor: level.ScaleFactor})
}

// SetLevel handles PUT /api/v1/me/fitness-level.
func (h *RoutineHandler) SetLevel(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LevelID string `json:"level_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	userID := middleware.UserID(r)
	if err := h.svc.SetLevel(r.Context(), userID, req.LevelID); err != nil {
		if errors.Is(err, ports.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Printf("setting fitness level: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	level, err := h.svc.GetLevel(r.Context(), userID)
	if err != nil {
		log.Printf("reading fitness level after set: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, fitnessLevelDTO{ID: level.ID, Name: level.Name, ScaleFactor: level.ScaleFactor})
}

func decodeRoutineRequest(w http.ResponseWriter, r *http.Request) (routineRequest, bool) {
	var req routineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return routineRequest{}, false
	}
	return req, true
}

// writeRoutineResult maps service errors to statuses; returns false when the
// response has been written.
func writeRoutineResult(w http.ResponseWriter, err error, what string) bool {
	if err == nil {
		return true
	}
	switch {
	case errors.Is(err, ports.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ports.ErrNotFound):
		writeError(w, http.StatusNotFound, "routine not found")
	default:
		log.Printf("%s: %v", what, err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
	return false
}
