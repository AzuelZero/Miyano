package domain

import (
	"math"
	"time"
)

// RoutineExercise is one exercise slot inside a routine. Numeric targets are
// optional (*int) — an exercise may be reps-based, time-based or both.
type RoutineExercise struct {
	ExerciseID      string
	Position        int
	TargetSets      *int
	TargetReps      *int
	TargetDurationS *int
	RestS           *int
}

// Routine is a user's routine with its exercises and tags.
type Routine struct {
	ID          string
	UserID      string
	Name        string
	Description *string
	CreatedAt   time.Time
	Exercises   []RoutineExercise
	Tags        []string
}

// FitnessLevel is one of the seeded levels with its auto-scale factor.
type FitnessLevel struct {
	ID          string
	Name        string
	ScaleFactor float64
}

// DefaultSetSeconds estimates one reps-based set (no explicit duration) when
// computing routine time (see OpenSpec add-routines design decision 3).
const DefaultSetSeconds = 45

func ptr[T any](v T) *T { return &v }

// Scaled returns a copy with reps and duration multiplied by factor,
// rounded half away from zero with a floor of 1. Sets and rest are never
// scaled: fractional sets or longer rests for beginners are not a thing.
func (ex RoutineExercise) Scaled(factor float64) RoutineExercise {
	out := ex
	if ex.TargetReps != nil {
		out.TargetReps = ptr(max(1, int(math.Round(float64(*ex.TargetReps)*factor))))
	}
	if ex.TargetDurationS != nil {
		out.TargetDurationS = ptr(max(1, int(math.Round(float64(*ex.TargetDurationS)*factor))))
	}
	return out
}

// EstimatedSeconds is the heuristic time of this exercise in a routine:
// sets × (duration or DefaultSetSeconds) + (sets-1) × rest. Absent sets
// count as 1; absent rest as 0.
func (ex RoutineExercise) EstimatedSeconds() int {
	sets := 1
	if ex.TargetSets != nil && *ex.TargetSets > 0 {
		sets = *ex.TargetSets
	}
	duration := DefaultSetSeconds
	if ex.TargetDurationS != nil && *ex.TargetDurationS > 0 {
		duration = *ex.TargetDurationS
	}
	rest := 0
	if ex.RestS != nil && *ex.RestS > 0 {
		rest = *ex.RestS
	}
	return sets*duration + (sets-1)*rest
}

// TotalEstimatedSeconds sums the exercise estimates (never persisted, D007).
func TotalEstimatedSeconds(exercises []RoutineExercise) int {
	total := 0
	for _, ex := range exercises {
		total += ex.EstimatedSeconds()
	}
	return total
}
