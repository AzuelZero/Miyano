package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"miyano/internal/adapter/postgres/gen"
	"miyano/internal/domain"
	"miyano/internal/ports"
)

// RoutineRepository implements ports.RoutineRepository on PostgreSQL.
// Every query is scoped by user_id: another user's routine reads as missing.
type RoutineRepository struct {
	pool *pgxpool.Pool
}

// NewRoutineRepository builds the repository from the connection pool.
func NewRoutineRepository(pool *pgxpool.Pool) *RoutineRepository {
	return &RoutineRepository{pool: pool}
}

// Create persists the whole routine atomically. Exercise positions are
// normalized from the slice order (index+1), so the stored order always
// matches what the client sent.
func (r *RoutineRepository) Create(ctx context.Context, userID, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error) {
	if err := r.validateExerciseIDs(ctx, exercises); err != nil {
		return domain.Routine{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Routine{}, fmt.Errorf("beginning routine transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := gen.New(tx)
	row, err := q.CreateRoutine(ctx, gen.CreateRoutineParams{
		UserID:      userID,
		Name:        name,
		Description: description,
	})
	if err != nil {
		return domain.Routine{}, fmt.Errorf("creating routine: %w", err)
	}

	stored, err := insertRoutineBody(ctx, q, row.ID, exercises, tags)
	if err != nil {
		return domain.Routine{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Routine{}, fmt.Errorf("committing routine creation: %w", err)
	}

	return domain.Routine{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Time,
		Exercises:   stored,
		Tags:        tags,
	}, nil
}

// Update replaces name, description, exercises and tags in one transaction
// (delete + insert of the body — see add-routines design decision 1).
func (r *RoutineRepository) Update(ctx context.Context, userID, routineID, name string, description *string, exercises []domain.RoutineExercise, tags []string) (domain.Routine, error) {
	if err := r.validateExerciseIDs(ctx, exercises); err != nil {
		return domain.Routine{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Routine{}, fmt.Errorf("beginning routine transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := gen.New(tx)
	row, err := q.UpdateRoutine(ctx, gen.UpdateRoutineParams{
		ID:          routineID,
		Name:        name,
		Description: description,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Routine{}, ports.ErrNotFound
		}
		return domain.Routine{}, fmt.Errorf("updating routine: %w", err)
	}

	if err := q.DeleteRoutineExercises(ctx, routineID); err != nil {
		return domain.Routine{}, fmt.Errorf("clearing routine exercises: %w", err)
	}
	if err := q.DeleteRoutineTags(ctx, routineID); err != nil {
		return domain.Routine{}, fmt.Errorf("clearing routine tags: %w", err)
	}

	stored, err := insertRoutineBody(ctx, q, routineID, exercises, tags)
	if err != nil {
		return domain.Routine{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Routine{}, fmt.Errorf("committing routine update: %w", err)
	}

	return domain.Routine{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Time,
		Exercises:   stored,
		Tags:        tags,
	}, nil
}

// Get returns one routine with exercises (by position) and tags.
func (r *RoutineRepository) Get(ctx context.Context, userID, routineID string) (domain.Routine, error) {
	q := gen.New(r.pool)
	row, err := q.GetRoutine(ctx, gen.GetRoutineParams{ID: routineID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Routine{}, ports.ErrNotFound
		}
		return domain.Routine{}, fmt.Errorf("getting routine: %w", err)
	}

	exRows, err := q.ListRoutineExercises(ctx, routineID)
	if err != nil {
		return domain.Routine{}, fmt.Errorf("listing routine exercises: %w", err)
	}
	tagRows, err := q.ListRoutineTags(ctx, routineID)
	if err != nil {
		return domain.Routine{}, fmt.Errorf("listing routine tags: %w", err)
	}

	return routineFromRows(row, exRows, tagRows), nil
}

// List returns the user's routines with their bodies attached, optionally
// filtered by an exact tag.
func (r *RoutineRepository) List(ctx context.Context, userID, tag string) ([]domain.Routine, error) {
	q := gen.New(r.pool)
	rows, err := q.ListRoutines(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing routines: %w", err)
	}
	exRows, err := q.ListRoutineExercisesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing routine exercises: %w", err)
	}
	tagRows, err := q.ListRoutineTagsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing routine tags: %w", err)
	}

	exByRoutine := map[string][]gen.RoutineExercise{}
	for _, ex := range exRows {
		exByRoutine[ex.RoutineID] = append(exByRoutine[ex.RoutineID], ex)
	}
	tagsByRoutine := map[string][]string{}
	for _, t := range tagRows {
		tagsByRoutine[t.RoutineID] = append(tagsByRoutine[t.RoutineID], t.Tag)
	}

	out := make([]domain.Routine, 0, len(rows))
	for _, row := range rows {
		tags := tagsByRoutine[row.ID]
		if tag != "" && !containsTag(tags, tag) {
			continue
		}
		out = append(out, routineFromRows(row, exByRoutine[row.ID], tags))
	}
	return out, nil
}

// Delete removes the routine; its exercises and tags go by ON DELETE CASCADE.
func (r *RoutineRepository) Delete(ctx context.Context, userID, routineID string) error {
	n, err := gen.New(r.pool).DeleteRoutine(ctx, gen.DeleteRoutineParams{ID: routineID, UserID: userID})
	if err != nil {
		return fmt.Errorf("deleting routine: %w", err)
	}
	if n == 0 {
		return ports.ErrNotFound
	}
	return nil
}

// validateExerciseIDs checks the whole set against the catalog in one query,
// so a routine never references a ghost exercise (FK as safety net only).
func (r *RoutineRepository) validateExerciseIDs(ctx context.Context, exercises []domain.RoutineExercise) error {
	if len(exercises) == 0 {
		return nil
	}
	ids := make([]string, 0, len(exercises))
	for _, ex := range exercises {
		ids = append(ids, ex.ExerciseID)
	}
	found, err := gen.New(r.pool).ValidateExerciseIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("validating exercise ids: %w", err)
	}
	seen := map[string]bool{}
	for _, id := range found {
		seen[id] = true
	}
	for _, id := range ids {
		if !seen[id] {
			return fmt.Errorf("%w: unknown exercise %q", ports.ErrInvalidInput, id)
		}
	}
	return nil
}

func insertRoutineBody(ctx context.Context, q *gen.Queries, routineID string, exercises []domain.RoutineExercise, tags []string) ([]domain.RoutineExercise, error) {
	stored := make([]domain.RoutineExercise, 0, len(exercises))
	for i, ex := range exercises {
		stored = append(stored, domain.RoutineExercise{
			ExerciseID:      ex.ExerciseID,
			Position:        i + 1,
			TargetSets:      ex.TargetSets,
			TargetReps:      ex.TargetReps,
			TargetDurationS: ex.TargetDurationS,
			RestS:           ex.RestS,
		})
		if err := q.AddRoutineExercise(ctx, gen.AddRoutineExerciseParams{
			RoutineID:       routineID,
			ExerciseID:      ex.ExerciseID,
			Position:        int32(i + 1),
			TargetSets:      i32ptr(ex.TargetSets),
			TargetReps:      i32ptr(ex.TargetReps),
			TargetDurationS: i32ptr(ex.TargetDurationS),
			RestS:           i32ptr(ex.RestS),
		}); err != nil {
			return nil, fmt.Errorf("adding exercise %s to routine: %w", ex.ExerciseID, err)
		}
	}
	for _, tag := range tags {
		if err := q.AddRoutineTag(ctx, gen.AddRoutineTagParams{RoutineID: routineID, Tag: tag}); err != nil {
			return nil, fmt.Errorf("adding tag %q to routine: %w", tag, err)
		}
	}
	return stored, nil
}

func routineFromRows(row gen.Routine, exRows []gen.RoutineExercise, tags []string) domain.Routine {
	exercises := make([]domain.RoutineExercise, 0, len(exRows))
	for _, ex := range exRows {
		exercises = append(exercises, domain.RoutineExercise{
			ExerciseID:      ex.ExerciseID,
			Position:        int(ex.Position),
			TargetSets:      intptr(ex.TargetSets),
			TargetReps:      intptr(ex.TargetReps),
			TargetDurationS: intptr(ex.TargetDurationS),
			RestS:           intptr(ex.RestS),
		})
	}
	if tags == nil {
		tags = []string{}
	}
	return domain.Routine{
		ID:          row.ID,
		UserID:      row.UserID,
		Name:        row.Name,
		Description: row.Description,
		CreatedAt:   row.CreatedAt.Time,
		Exercises:   exercises,
		Tags:        tags,
	}
}

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func i32ptr(i *int) *int32 {
	if i == nil {
		return nil
	}
	v := int32(*i) // #nosec G115 -- range-validated by the service (0..math.MaxInt32)
	return &v
}

func intptr(i *int32) *int {
	if i == nil {
		return nil
	}
	v := int(*i)
	return &v
}

// FitnessLevelRepository implements ports.FitnessLevelRepository.
type FitnessLevelRepository struct {
	pool *pgxpool.Pool
}

// NewFitnessLevelRepository builds the repository from the connection pool.
func NewFitnessLevelRepository(pool *pgxpool.Pool) *FitnessLevelRepository {
	return &FitnessLevelRepository{pool: pool}
}

// GetByID returns one seeded level or ErrNotFound.
func (r *FitnessLevelRepository) GetByID(ctx context.Context, levelID string) (domain.FitnessLevel, error) {
	return levelFromRow(gen.New(r.pool).GetFitnessLevel(ctx, levelID))
}

// GetUserLevel returns the user's level or ErrNotFound when unset.
func (r *FitnessLevelRepository) GetUserLevel(ctx context.Context, userID string) (domain.FitnessLevel, error) {
	return levelFromRow(gen.New(r.pool).GetUserFitnessLevel(ctx, userID))
}

// SetUserLevel upserts the user's level; unknown levels read as ErrNotFound.
func (r *FitnessLevelRepository) SetUserLevel(ctx context.Context, userID, levelID string) error {
	q := gen.New(r.pool)
	if _, err := q.GetFitnessLevel(ctx, levelID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ErrNotFound
		}
		return fmt.Errorf("checking fitness level: %w", err)
	}
	if err := q.SetUserFitnessLevel(ctx, gen.SetUserFitnessLevelParams{UserID: userID, FitnessLevelID: levelID}); err != nil {
		return fmt.Errorf("setting fitness level: %w", err)
	}
	return nil
}

func levelFromRow(row gen.FitnessLevel, err error) (domain.FitnessLevel, error) {
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.FitnessLevel{}, ports.ErrNotFound
		}
		return domain.FitnessLevel{}, fmt.Errorf("reading fitness level: %w", err)
	}
	f, err := row.ScaleFactor.Float64Value()
	if err != nil || !f.Valid {
		return domain.FitnessLevel{}, fmt.Errorf("reading scale factor: %w", err)
	}
	return domain.FitnessLevel{ID: row.ID, Name: row.Name, ScaleFactor: f.Float64}, nil
}
