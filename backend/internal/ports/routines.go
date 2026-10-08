package ports

import (
	"context"
	"errors"

	"miyano/internal/domain"
)

// ErrInvalidInput signals a client mistake the service could not rescue
// (empty name, unknown exercise id, unknown level). Handlers map it to 400.
var ErrInvalidInput = errors.New("invalid input")

// RoutineRepository persists routines with their exercises and tags.
// Every method is scoped to userID: another user's routine is
// indistinguishable from a missing one (ErrNotFound, never 403).
type RoutineRepository interface {
	// Create persists the whole routine atomically. Returns ErrInvalidInput
	// when an exercise id is not in the catalog.
	Create(ctx context.Context, userID string, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error)

	// Update replaces name, description, exercises and tags atomically.
	// ErrNotFound when the routine does not belong to the user.
	Update(ctx context.Context, userID, routineID string, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error)

	// Get returns one routine (exercises ordered by position). ErrNotFound
	// when missing or owned by someone else.
	Get(ctx context.Context, userID, routineID string) (domain.Routine, error)

	// List returns the user's routines (newest first) with exercises and
	// tags attached. An empty tag returns all of them; otherwise only those
	// carrying the exact tag.
	List(ctx context.Context, userID, tag string) ([]domain.Routine, error)

	// Delete removes the routine (cascading exercises and tags).
	// ErrNotFound when missing or owned by someone else.
	Delete(ctx context.Context, userID, routineID string) error
}

// FitnessLevelRepository reads and writes the user's fitness level.
type FitnessLevelRepository interface {
	// GetByID returns one seeded level or ErrNotFound.
	GetByID(ctx context.Context, levelID string) (domain.FitnessLevel, error)

	// GetUserLevel returns the user's level or ErrNotFound when unset.
	GetUserLevel(ctx context.Context, userID string) (domain.FitnessLevel, error)

	// SetUserLevel upserts the user's level. ErrNotFound for an unknown level.
	SetUserLevel(ctx context.Context, userID, levelID string) error
}
