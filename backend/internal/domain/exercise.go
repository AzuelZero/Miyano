// Package domain holds the pure business entities. No I/O, no frameworks.
package domain

// Exercise is a catalog exercise, imported from the dataset.
type Exercise struct {
	ID           string
	Name         string
	Category     string
	BodyPart     string
	Target       *string
	MuscleGroup  *string
	Equipment    string
	ForceType    *string
	Mechanic     *string
	Difficulty   *string
	IsUnilateral bool
	IsBodyweight bool
	Met          *float64
	GifURL       *string
	ImageURL     *string
	Attribution  *string
}

// Translation is a localized version of an exercise's instructions.
type Translation struct {
	ExerciseID       string
	Lang             string
	Name             *string
	Instructions     *string
	InstructionSteps []string
}
