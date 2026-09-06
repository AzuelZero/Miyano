package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"miyano/internal/domain"
	"miyano/internal/ports"
	"miyano/internal/service"
)

// ExerciseHandler serves the exercise catalog endpoints.
type ExerciseHandler struct {
	svc *service.ExerciseService
}

// NewExerciseHandler wires the handler with its service.
func NewExerciseHandler(svc *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{svc: svc}
}

type exerciseDTO struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	BodyPart     string   `json:"body_part"`
	Equipment    string   `json:"equipment"`
	ForceType    *string  `json:"force_type,omitempty"`
	Difficulty   *string  `json:"difficulty,omitempty"`
	IsBodyweight bool     `json:"is_bodyweight"`
	GifURL       *string  `json:"gif_url,omitempty"`
}

type translationDTO struct {
	Lang             string   `json:"lang"`
	Instructions     *string  `json:"instructions,omitempty"`
	InstructionSteps []string `json:"instruction_steps,omitempty"`
}

type exerciseDetailDTO struct {
	exerciseDTO
	Translations []translationDTO `json:"translations"`
}

func toDTO(ex domain.Exercise) exerciseDTO {
	return exerciseDTO{
		ID:           ex.ID,
		Name:         ex.Name,
		Category:     ex.Category,
		BodyPart:     ex.BodyPart,
		Equipment:    ex.Equipment,
		ForceType:    ex.ForceType,
		Difficulty:   ex.Difficulty,
		IsBodyweight: ex.IsBodyweight,
		GifURL:       ex.GifURL,
	}
}

// List handles GET /api/v1/exercises (optional ?equipment= exact filter).
// Public until the auth capability lands (see OpenSpec exercise-catalog).
func (h *ExerciseHandler) List(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.svc.List(r.Context(), r.URL.Query().Get("equipment"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	dtos := make([]exerciseDTO, 0, len(exercises))
	for _, ex := range exercises {
		dtos = append(dtos, toDTO(ex))
	}
	writeJSON(w, http.StatusOK, dtos)
}

// GetByID handles GET /api/v1/exercises/{id}.
func (h *ExerciseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	detail, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			writeError(w, http.StatusNotFound, "exercise not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	dto := exerciseDetailDTO{Translations: make([]translationDTO, 0, len(detail.Translations))}
	dto.exerciseDTO = toDTO(detail.Exercise)
	for _, tr := range detail.Translations {
		dto.Translations = append(dto.Translations, translationDTO{
			Lang:             tr.Lang,
			Instructions:     tr.Instructions,
			InstructionSteps: tr.InstructionSteps,
		})
	}
	writeJSON(w, http.StatusOK, dto)
}
