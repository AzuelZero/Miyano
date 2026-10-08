package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"miyano/internal/domain"
	"miyano/internal/ports"
)

// fakes: in-memory repos, no database.

type fakeRoutineRepo struct {
	byUser    map[string]map[string]domain.Routine
	knownExes map[string]bool
	nextID    int
}

func newFakeRoutineRepo(knownExes ...string) *fakeRoutineRepo {
	known := map[string]bool{}
	for _, id := range knownExes {
		known[id] = true
	}
	return &fakeRoutineRepo{byUser: map[string]map[string]domain.Routine{}, knownExes: known}
}

func (f *fakeRoutineRepo) validate(exes []domain.RoutineExercise) error {
	for _, ex := range exes {
		if !f.knownExes[ex.ExerciseID] {
			return fmt.Errorf("%w: unknown exercise %q", ports.ErrInvalidInput, ex.ExerciseID)
		}
	}
	return nil
}

func (f *fakeRoutineRepo) Create(_ context.Context, userID, name string, description *string, exes []domain.RoutineExercise, tags []string) (domain.Routine, error) {
	if err := f.validate(exes); err != nil {
		return domain.Routine{}, err
	}
	f.nextID++
	id := fmt.Sprintf("r-%d", f.nextID)
	stored := make([]domain.RoutineExercise, 0, len(exes))
	for i, ex := range exes {
		ex.Position = i + 1
		stored = append(stored, ex)
	}
	r := domain.Routine{ID: id, UserID: userID, Name: name, Description: description, Exercises: stored, Tags: tags}
	if f.byUser[userID] == nil {
		f.byUser[userID] = map[string]domain.Routine{}
	}
	f.byUser[userID][id] = r
	return r, nil
}

func (f *fakeRoutineRepo) Update(_ context.Context, userID, routineID, name string, description *string, exes []domain.RoutineExercise, tags []string) (domain.Routine, error) {
	if err := f.validate(exes); err != nil {
		return domain.Routine{}, err
	}
	r, ok := f.byUser[userID][routineID]
	if !ok {
		return domain.Routine{}, ports.ErrNotFound
	}
	r.Name, r.Description, r.Tags = name, description, tags
	r.Exercises = append([]domain.RoutineExercise(nil), exes...)
	f.byUser[userID][routineID] = r
	return r, nil
}

func (f *fakeRoutineRepo) Get(_ context.Context, userID, routineID string) (domain.Routine, error) {
	r, ok := f.byUser[userID][routineID]
	if !ok {
		return domain.Routine{}, ports.ErrNotFound
	}
	return r, nil
}

func (f *fakeRoutineRepo) List(_ context.Context, userID, tag string) ([]domain.Routine, error) {
	var out []domain.Routine
	for _, r := range f.byUser[userID] {
		if tag == "" || containsTag(r.Tags, tag) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeRoutineRepo) Delete(_ context.Context, userID, routineID string) error {
	if _, ok := f.byUser[userID][routineID]; !ok {
		return ports.ErrNotFound
	}
	delete(f.byUser[userID], routineID)
	return nil
}

func containsTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

type fakeLevelRepo struct {
	levels map[string]domain.FitnessLevel
	users  map[string]string
}

func newFakeLevelRepo() *fakeLevelRepo {
	return &fakeLevelRepo{
		levels: map[string]domain.FitnessLevel{
			"beginner":     {ID: "beginner", Name: "Baja forma", ScaleFactor: 0.70},
			"intermediate": {ID: "intermediate", Name: "En forma", ScaleFactor: 1.00},
			"advanced":     {ID: "advanced", Name: "Muy en forma", ScaleFactor: 1.30},
		},
		users: map[string]string{},
	}
}

func (f *fakeLevelRepo) GetByID(_ context.Context, id string) (domain.FitnessLevel, error) {
	l, ok := f.levels[id]
	if !ok {
		return domain.FitnessLevel{}, ports.ErrNotFound
	}
	return l, nil
}

func (f *fakeLevelRepo) GetUserLevel(_ context.Context, userID string) (domain.FitnessLevel, error) {
	id, ok := f.users[userID]
	if !ok {
		return domain.FitnessLevel{}, ports.ErrNotFound
	}
	return f.levels[id], nil
}

func (f *fakeLevelRepo) SetUserLevel(_ context.Context, userID, levelID string) error {
	if _, ok := f.levels[levelID]; !ok {
		return ports.ErrNotFound
	}
	f.users[userID] = levelID
	return nil
}

func newRoutineService(knownExes ...string) (*RoutineService, *fakeRoutineRepo, *fakeLevelRepo) {
	repo := newFakeRoutineRepo(knownExes...)
	levels := newFakeLevelRepo()
	return NewRoutineService(repo, levels), repo, levels
}

func iptr(i int) *int { return &i }

func TestCreateScalesForDefaultBeginner(t *testing.T) {
	svc, _, _ := newRoutineService("0001")
	ctx := context.Background()

	detail, err := svc.Create(ctx, "user-a", RoutineInput{
		Name: "Leg day",
		Exercises: []domain.RoutineExercise{{
			ExerciseID: "0001", Position: 1,
			TargetSets: iptr(3), TargetReps: iptr(10), TargetDurationS: iptr(60), RestS: iptr(90),
		}},
		Tags: []string{"pierna"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if detail.Level.ID != DefaultFitnessLevelID {
		t.Errorf("level = %q, want default %q", detail.Level.ID, DefaultFitnessLevelID)
	}
	// 10 reps × 0.70 = 7 (spec scenario).
	if got := *detail.ScaledExercises[0].TargetReps; got != 7 {
		t.Errorf("scaled reps = %d, want 7", got)
	}
	if got := *detail.Routine.Exercises[0].TargetReps; got != 10 {
		t.Errorf("base reps = %d, want 10 (base must stay intact)", got)
	}
	// Spec scenario: 3×60 + 2×90 = 360.
	if detail.TotalTimeS != 360 {
		t.Errorf("TotalTimeS = %d, want 360", detail.TotalTimeS)
	}
	// Scaled: duration 60×0.7=42 → 3×42 + 2×90 = 306.
	if detail.TotalTimeScaledS != 306 {
		t.Errorf("TotalTimeScaledS = %d, want 306", detail.TotalTimeScaledS)
	}
}

func TestScaledRoundingAndMinimum(t *testing.T) {
	ex := domain.RoutineExercise{TargetReps: iptr(3), TargetDurationS: iptr(45)}

	down := ex.Scaled(0.70) // 3×0.7=2.1 → 2; 45×0.7 = 31.4999… in float64 → 31
	if *down.TargetReps != 2 || *down.TargetDurationS != 31 {
		t.Errorf("scaled(0.7) = reps %d, dur %d; want 2, 31", *down.TargetReps, *down.TargetDurationS)
	}

	up := ex.Scaled(1.30) // 3×1.3=3.9 → 4; 45×1.3=58.5 → 59
	if *up.TargetReps != 4 || *up.TargetDurationS != 59 {
		t.Errorf("scaled(1.3) = reps %d, dur %d; want 4, 59", *up.TargetReps, *up.TargetDurationS)
	}

	floor := domain.RoutineExercise{TargetReps: iptr(1)}.Scaled(0.70) // 0.7 → minimum 1
	if *floor.TargetReps != 1 {
		t.Errorf("scaled floor = %d, want 1", *floor.TargetReps)
	}

	// Sets and rest are never scaled.
	keeper := domain.RoutineExercise{TargetSets: iptr(5), RestS: iptr(120), TargetReps: iptr(10)}.Scaled(1.30)
	if *keeper.TargetSets != 5 || *keeper.RestS != 120 {
		t.Errorf("sets/rest changed: sets %d rest %d, want 5/120", *keeper.TargetSets, *keeper.RestS)
	}
}

func TestTimeWithoutExplicitDuration(t *testing.T) {
	// 3 sets, reps-based (no duration), 60s rest → 3×45 + 2×60 = 255.
	ex := domain.RoutineExercise{TargetSets: iptr(3), TargetReps: iptr(12), RestS: iptr(60)}
	if got := ex.EstimatedSeconds(); got != 255 {
		t.Errorf("EstimatedSeconds() = %d, want 255", got)
	}
	// Absent sets count as one set.
	single := domain.RoutineExercise{TargetDurationS: iptr(30)}
	if got := single.EstimatedSeconds(); got != 30 {
		t.Errorf("single set = %d, want 30", got)
	}
}

func TestCreateValidatesExercisesAndName(t *testing.T) {
	svc, repo, _ := newRoutineService("0001")
	ctx := context.Background()

	if _, err := svc.Create(ctx, "user-a", RoutineInput{
		Name:      "Ghost",
		Exercises: []domain.RoutineExercise{{ExerciseID: "9999"}},
	}); !errors.Is(err, ports.ErrInvalidInput) {
		t.Errorf("Create(unknown exercise) error = %v, want ErrInvalidInput", err)
	}
	if len(repo.byUser["user-a"]) != 0 {
		t.Error("routine persisted despite invalid exercise")
	}

	if _, err := svc.Create(ctx, "user-a", RoutineInput{Name: "  "}); !errors.Is(err, ports.ErrInvalidInput) {
		t.Errorf("Create(blank name) error = %v, want ErrInvalidInput", err)
	}
}

func TestUserIsolationIs404(t *testing.T) {
	svc, _, _ := newRoutineService("0001")
	ctx := context.Background()

	detail, err := svc.Create(ctx, "user-a", RoutineInput{Name: "Mine", Exercises: []domain.RoutineExercise{{ExerciseID: "0001"}}})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if _, err := svc.Get(ctx, "user-b", detail.Routine.ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("Get(other user) error = %v, want ErrNotFound", err)
	}
	list, err := svc.List(ctx, "user-b", "")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("user-b sees %d routines, want 0", len(list))
	}
}

func TestUpdateReplacesAndDeleteRemoves(t *testing.T) {
	svc, _, _ := newRoutineService("0001", "0002")
	ctx := context.Background()

	detail, err := svc.Create(ctx, "user-a", RoutineInput{
		Name:      "V1",
		Tags:      []string{"old"},
		Exercises: []domain.RoutineExercise{{ExerciseID: "0001"}},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := svc.Update(ctx, "user-a", detail.Routine.ID, RoutineInput{
		Name:      "V2",
		Tags:      []string{"new", "pierna"},
		Exercises: []domain.RoutineExercise{{ExerciseID: "0002"}, {ExerciseID: "0001"}},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Routine.Name != "V2" || len(updated.Routine.Exercises) != 2 || updated.Routine.Tags[0] != "new" {
		t.Errorf("Update() result = %+v, want replaced body", updated.Routine)
	}

	if err := svc.Delete(ctx, "user-a", detail.Routine.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := svc.Get(ctx, "user-a", detail.Routine.ID); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("Get(after delete) error = %v, want ErrNotFound", err)
	}
}

func TestListFiltersByExactTag(t *testing.T) {
	svc, _, _ := newRoutineService("0001")
	ctx := context.Background()

	if _, err := svc.Create(ctx, "user-a", RoutineInput{Name: "Pierna", Tags: []string{"pierna"}}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := svc.Create(ctx, "user-a", RoutineInput{Name: "Pecho", Tags: []string{"pecho"}}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	all, _ := svc.List(ctx, "user-a", "")
	if len(all) != 2 {
		t.Errorf("List() = %d routines, want 2", len(all))
	}
	legs, _ := svc.List(ctx, "user-a", "pierna")
	if len(legs) != 1 || legs[0].Name != "Pierna" {
		t.Errorf("List(pierna) = %+v, want only Pierna", legs)
	}
}

func TestLevelDefaultAndSet(t *testing.T) {
	svc, _, levels := newRoutineService("0001")
	ctx := context.Background()

	level, err := svc.GetLevel(ctx, "user-a")
	if err != nil {
		t.Fatalf("GetLevel() error = %v", err)
	}
	if level.ID != "beginner" || level.ScaleFactor != 0.70 {
		t.Errorf("default level = %+v, want beginner ×0.70", level)
	}

	if err := svc.SetLevel(ctx, "user-a", "advanced"); err != nil {
		t.Fatalf("SetLevel() error = %v", err)
	}
	if levels.users["user-a"] != "advanced" {
		t.Errorf("stored level = %q, want advanced", levels.users["user-a"])
	}

	level, err = svc.GetLevel(ctx, "user-a")
	if err != nil || level.ScaleFactor != 1.30 {
		t.Errorf("GetLevel(after set) = %+v err %v, want ×1.30", level, err)
	}

	if err := svc.SetLevel(ctx, "user-a", "gigachad"); !errors.Is(err, ports.ErrInvalidInput) {
		t.Errorf("SetLevel(unknown) error = %v, want ErrInvalidInput", err)
	}
}
