package ports

import (
	"context"
	"errors"

	"miyano/internal/domain"
)

// ErrInvalidInput signals a client mistake the service could not rescue
// (empty name, unknown exercise id, unknown level). Handlers map it to 400.
var ErrInvalidInput = errors.New("invalid input")

// ErrConflict signals a conditioned write (If-Match) whose revision no
// longer matches. Handlers map it to 409.
var ErrConflict = errors.New("revision conflict")

// ConflictError carries the entity's current state so the transport can
// attach it to the 409 body — the client merges without an extra request
// (sync contract D1, design decision 2).
type ConflictError struct {
	Current domain.Routine
}

func (e *ConflictError) Error() string { return "routine was modified concurrently" }

// Is makes errors.Is(err, ErrConflict) work through the wrapper.
func (e *ConflictError) Is(target error) bool { return target == ErrConflict }

// RoutineRepository persists routines with their exercises and tags.
// Every method is scoped to userID: another user's routine is
// indistinguishable from a missing one (ErrNotFound, never 403).
type RoutineRepository interface {
	// Create persists the whole routine atomically. Returns ErrInvalidInput
	// when an exercise id is not in the catalog.
	Create(ctx context.Context, userID string, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error)

	// Update replaces name, description, exercises and tags atomically.
	// ErrNotFound when the routine does not belong to the user. A non-nil
	// expectedRevision conditions the write (stale revision → *ConflictError
	// with the current row attached, never a bare 404); nil writes
	// unconditionally (today's behavior).
	Update(ctx context.Context, userID string, routineID string, name string, description *string, exercises []domain.RoutineExercise, tags []string, expectedRevision *int64) (domain.Routine, error)

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

	// ProfileRev returns the user's monotonic change counter (0 when no
	// write has happened yet) — the cursor for cheap sync polling.
	ProfileRev(ctx context.Context, userID string) (int64, error)
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
