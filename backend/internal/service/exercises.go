// Package service holds the use cases orchestrating the domain.
package service

import (
	"context"

	"miyano/internal/domain"
	"miyano/internal/ports"
)

// ExerciseService serves the exercise catalog use cases.
type ExerciseService struct {
	repo ports.ExerciseRepository
}

// NewExerciseService wires the service with its repository port.
func NewExerciseService(repo ports.ExerciseRepository) *ExerciseService {
	return &ExerciseService{repo: repo}
}

// List returns the catalog, optionally filtered by equipment.
func (s *ExerciseService) List(ctx context.Context, equipment string) ([]domain.Exercise, error) {
	return s.repo.List(ctx, equipment)
}

// ExerciseDetail is an exercise with its translations.
type ExerciseDetail struct {
	Exercise     domain.Exercise
	Translations []domain.Translation
}

// GetByID returns one exercise with its translations.
func (s *ExerciseService) GetByID(ctx context.Context, id string) (ExerciseDetail, error) {
	ex, translations, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ExerciseDetail{}, err
	}
	return ExerciseDetail{Exercise: ex, Translations: translations}, nil
}
