-- name: ListExercises :many
SELECT * FROM exercises
ORDER BY id;

-- name: ListExercisesByEquipment :many
SELECT * FROM exercises
WHERE equipment = $1
ORDER BY id;

-- name: GetExercise :one
SELECT * FROM exercises
WHERE id = $1;

-- name: ListExerciseTranslations :many
SELECT * FROM exercise_translations
WHERE exercise_id = $1;
