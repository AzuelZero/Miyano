package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"miyano/internal/domain"
	"miyano/internal/ports"
)

// DefaultFitnessLevelID is the level assumed when the user never set one
// (lazy default — never persisted until the user chooses, design decision 5).
const DefaultFitnessLevelID = "beginner"

// RoutineService serves the routine use cases: CRUD, tag filtering,
// level management and the scaled/time-computed views (ADR D007).
type RoutineService struct {
	routines ports.RoutineRepository
	levels   ports.FitnessLevelRepository
}

// NewRoutineService wires the service with its ports.
func NewRoutineService(routines ports.RoutineRepository, levels ports.FitnessLevelRepository) *RoutineService {
	return &RoutineService{routines: routines, levels: levels}
}

// RoutineInput is what create and update receive from the transport.
type RoutineInput struct {
	Name        string
	Description *string
	Exercises   []domain.RoutineExercise
	Tags        []string
}

// RoutineSummary is the list item: identity plus computed time.
type RoutineSummary struct {
	ID            string
	Name          string
	Description   *string
	Tags          []string
	ExerciseCount int
	TotalTimeS    int
}

// RoutineDetail is the full view: the base routine plus the variant scaled
// to the user's level and both time estimates.
type RoutineDetail struct {
	Routine           domain.Routine
	ScaledExercises   []domain.RoutineExercise
	Level             domain.FitnessLevel
	TotalTimeS        int
	TotalTimeScaledS  int
}

// Create validates and persists a new routine, returning its full detail.
func (s *RoutineService) Create(ctx context.Context, userID string, in RoutineInput) (RoutineDetail, error) {
	if err := validateRoutineInput(in); err != nil {
		return RoutineDetail{}, err
	}
	routine, err := s.routines.Create(ctx, userID, in.Name, in.Description, in.Exercises, in.Tags)
	if err != nil {
		return RoutineDetail{}, err
	}
	return s.detail(ctx, userID, routine)
}

// Update replaces the routine completely and returns the new detail.
func (s *RoutineService) Update(ctx context.Context, userID, routineID string, in RoutineInput) (RoutineDetail, error) {
	if err := validateRoutineInput(in); err != nil {
		return RoutineDetail{}, err
	}
	routine, err := s.routines.Update(ctx, userID, routineID, in.Name, in.Description, in.Exercises, in.Tags)
	if err != nil {
		return RoutineDetail{}, err
	}
	return s.detail(ctx, userID, routine)
}

// Get returns one routine with its scaled variant and time estimates.
func (s *RoutineService) Get(ctx context.Context, userID, routineID string) (RoutineDetail, error) {
	routine, err := s.routines.Get(ctx, userID, routineID)
	if err != nil {
		return RoutineDetail{}, err
	}
	return s.detail(ctx, userID, routine)
}

// List returns the user's routines (optionally by exact tag) as summaries.
func (s *RoutineService) List(ctx context.Context, userID, tag string) ([]RoutineSummary, error) {
	routines, err := s.routines.List(ctx, userID, tag)
	if err != nil {
		return nil, err
	}
	out := make([]RoutineSummary, 0, len(routines))
	for _, r := range routines {
		out = append(out, RoutineSummary{
			ID:            r.ID,
			Name:          r.Name,
			Description:   r.Description,
			Tags:          r.Tags,
			ExerciseCount: len(r.Exercises),
			TotalTimeS:    domain.TotalEstimatedSeconds(r.Exercises),
		})
	}
	return out, nil
}

// Delete removes one routine (ErrNotFound when missing or not owned).
func (s *RoutineService) Delete(ctx context.Context, userID, routineID string) error {
	return s.routines.Delete(ctx, userID, routineID)
}

// GetLevel returns the user's level, defaulting to beginner when unset.
func (s *RoutineService) GetLevel(ctx context.Context, userID string) (domain.FitnessLevel, error) {
	level, err := s.levels.GetUserLevel(ctx, userID)
	if err == nil {
		return level, nil
	}
	if errors.Is(err, ports.ErrNotFound) {
		return s.levels.GetByID(ctx, DefaultFitnessLevelID)
	}
	return domain.FitnessLevel{}, err
}

// SetLevel fixes the user's level (400 for an unknown one).
func (s *RoutineService) SetLevel(ctx context.Context, userID, levelID string) error {
	if err := s.levels.SetUserLevel(ctx, userID, levelID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return fmt.Errorf("%w: unknown fitness level %q", ports.ErrInvalidInput, levelID)
		}
		return err
	}
	return nil
}

func (s *RoutineService) detail(ctx context.Context, userID string, routine domain.Routine) (RoutineDetail, error) {
	level, err := s.GetLevel(ctx, userID)
	if err != nil {
		return RoutineDetail{}, err
	}

	scaled := make([]domain.RoutineExercise, 0, len(routine.Exercises))
	for _, ex := range routine.Exercises {
		scaled = append(scaled, ex.Scaled(level.ScaleFactor))
	}

	return RoutineDetail{
		Routine:          routine,
		ScaledExercises:  scaled,
		Level:            level,
		TotalTimeS:       domain.TotalEstimatedSeconds(routine.Exercises),
		TotalTimeScaledS: domain.TotalEstimatedSeconds(scaled),
	}, nil
}

func validateRoutineInput(in RoutineInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("%w: name is required", ports.ErrInvalidInput)
	}
	for _, ex := range in.Exercises {
		for _, v := range []*int{ex.TargetSets, ex.TargetReps, ex.TargetDurationS, ex.RestS} {
			if v != nil && (*v < 0 || *v > math.MaxInt32) {
				return fmt.Errorf("%w: exercise %q target out of range (0..%d)", ports.ErrInvalidInput, ex.ExerciseID, math.MaxInt32)
			}
		}
	}
	return nil
}
