-- name: CreateRoutine :one
INSERT INTO routines (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateRoutine :one
UPDATE routines SET name = $2, description = $3
WHERE id = $1 AND user_id = $4
RETURNING *;

-- name: DeleteRoutineExercises :exec
DELETE FROM routine_exercises WHERE routine_id = $1;

-- name: DeleteRoutineTags :exec
DELETE FROM routine_tags WHERE routine_id = $1;

-- name: AddRoutineExercise :exec
INSERT INTO routine_exercises (routine_id, exercise_id, position, target_sets, target_reps, target_duration_s, rest_s)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: AddRoutineTag :exec
INSERT INTO routine_tags (routine_id, tag)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetRoutine :one
SELECT * FROM routines
WHERE id = $1 AND user_id = $2;

-- name: ListRoutines :many
SELECT * FROM routines
WHERE user_id = $1
ORDER BY created_at DESC, id;

-- name: ListRoutineExercises :many
SELECT * FROM routine_exercises
WHERE routine_id = $1
ORDER BY position;

-- name: ListRoutineExercisesByUser :many
SELECT re.* FROM routine_exercises re
JOIN routines r ON r.id = re.routine_id
WHERE r.user_id = $1
ORDER BY re.routine_id, re.position;

-- name: ListRoutineTagsByUser :many
SELECT rt.routine_id, rt.tag FROM routine_tags rt
JOIN routines r ON r.id = rt.routine_id
WHERE r.user_id = $1;

-- name: ListRoutineTags :many
SELECT tag FROM routine_tags
WHERE routine_id = $1;

-- name: DeleteRoutine :execrows
DELETE FROM routines
WHERE id = $1 AND user_id = $2;

-- name: ValidateExerciseIDs :many
SELECT id FROM exercises
WHERE id = ANY($1::text[]);

-- name: GetFitnessLevel :one
SELECT * FROM fitness_levels
WHERE id = $1;

-- name: GetUserFitnessLevel :one
SELECT fl.* FROM fitness_levels fl
JOIN user_fitness_profile ufp ON ufp.fitness_level_id = fl.id
WHERE ufp.user_id = $1;

-- name: SetUserFitnessLevel :exec
INSERT INTO user_fitness_profile (user_id, fitness_level_id)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET fitness_level_id = $2, updated_at = now();
