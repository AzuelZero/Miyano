// Package ports defines the interfaces the domain depends on and the
// infrastructure implements (hexagonal / ports & adapters).
package ports

import (
	"context"
	"errors"

	"miyano/internal/domain"
)

// ErrNotFound is returned when an entity does not exist.
var ErrNotFound = errors.New("not found")

// ExerciseRepository is the storage contract for the exercise catalog.
type ExerciseRepository interface {
	// List returns all exercises, optionally filtered by exact equipment.
	// An empty equipment filter returns the full catalog.
	List(ctx context.Context, equipment string) ([]domain.Exercise, error)

	// GetByID returns one exercise plus its available translations.
	// Returns ErrNotFound when the id does not exist.
	GetByID(ctx context.Context, id string) (domain.Exercise, []domain.Translation, error)
}
