package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"miyano/internal/adapter/postgres/gen"
	"miyano/internal/domain"
	"miyano/internal/ports"
)

// ExerciseRepository implements ports.ExerciseRepository on PostgreSQL
// using the sqlc-generated queries.
type ExerciseRepository struct {
	q *gen.Queries
}

// NewExerciseRepository builds the repository from a pool (or tx).
func NewExerciseRepository(pool gen.DBTX) *ExerciseRepository {
	return &ExerciseRepository{q: gen.New(pool)}
}

// List returns all exercises, optionally filtered by exact equipment.
func (r *ExerciseRepository) List(ctx context.Context, equipment string) ([]domain.Exercise, error) {
	var rows []gen.Exercise
	var err error

	if equipment == "" {
		rows, err = r.q.ListExercises(ctx)
	} else {
		rows, err = r.q.ListExercisesByEquipment(ctx, equipment)
	}
	if err != nil {
		return nil, fmt.Errorf("listing exercises: %w", err)
	}

	out := make([]domain.Exercise, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

// GetByID returns one exercise plus its translations.
func (r *ExerciseRepository) GetByID(ctx context.Context, id string) (domain.Exercise, []domain.Translation, error) {
	row, err := r.q.GetExercise(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Exercise{}, nil, ports.ErrNotFound
		}
		return domain.Exercise{}, nil, fmt.Errorf("getting exercise %s: %w", id, err)
	}

	trRows, err := r.q.ListExerciseTranslations(ctx, id)
	if err != nil {
		return domain.Exercise{}, nil, fmt.Errorf("listing translations for %s: %w", id, err)
	}

	translations := make([]domain.Translation, 0, len(trRows))
	for _, tr := range trRows {
		steps := decodeSteps(tr.InstructionSteps)
		translations = append(translations, domain.Translation{
			ExerciseID:       tr.ExerciseID,
			Lang:             tr.Lang,
			Name:             tr.Name,
			Instructions:     tr.Instructions,
			InstructionSteps: steps,
		})
	}
	return toDomain(row), translations, nil
}

func toDomain(row gen.Exercise) domain.Exercise {
	return domain.Exercise{
		ID:           row.ID,
		Name:         row.Name,
		Category:     row.Category,
		BodyPart:     row.BodyPart,
		Target:       row.Target,
		MuscleGroup:  row.MuscleGroup,
		Equipment:    row.Equipment,
		ForceType:    row.ForceType,
		Mechanic:     row.Mechanic,
		Difficulty:   row.Difficulty,
		IsUnilateral: row.IsUnilateral,
		IsBodyweight: row.IsBodyweight,
		Met:          numericToPtr(row.Met),
		GifURL:       row.GifUrl,
		ImageURL:     row.ImageUrl,
		Attribution:  row.Attribution,
	}
}

func numericToPtr(n pgtype.Numeric) *float64 {
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	v := f.Float64
	return &v
}

func decodeSteps(raw []byte) []string {
	if len(raw) == 0 {
		return nil
	}
	var steps []string
	if err := json.Unmarshal(raw, &steps); err != nil {
		return nil
	}
	return steps
}
