-- Miyano initial schema (core relational model, see Miyano [base de datos].md ERD).

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    display_name  TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE profiles (
    user_id            UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name       TEXT NOT NULL,
    bio                TEXT,
    avatar_url         TEXT,
    avatar_updated_at  TIMESTAMPTZ,
    weight_kg          NUMERIC(5,2),
    height_cm          NUMERIC(5,1),
    goal               TEXT,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE exercises (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    category        TEXT NOT NULL,
    body_part       TEXT NOT NULL,
    target          TEXT,
    muscle_group    TEXT,
    equipment       TEXT NOT NULL,
    force_type      TEXT,
    mechanic        TEXT,
    difficulty      TEXT,
    is_unilateral   BOOLEAN NOT NULL DEFAULT FALSE,
    is_bodyweight   BOOLEAN NOT NULL DEFAULT FALSE,
    met             NUMERIC(4,2),
    gif_url         TEXT,
    image_url       TEXT,
    attribution     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_exercises_equipment ON exercises(equipment);
CREATE INDEX idx_exercises_category  ON exercises(category);
CREATE INDEX idx_exercises_force     ON exercises(force_type);

CREATE TABLE exercise_translations (
    exercise_id       TEXT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    lang              TEXT NOT NULL,
    name              TEXT,
    instructions      TEXT,
    instruction_steps JSONB,
    PRIMARY KEY (exercise_id, lang)
);

CREATE TABLE user_equipment (
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    equipment TEXT NOT NULL,
    PRIMARY KEY (user_id, equipment)
);

CREATE TABLE routines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE routine_exercises (
    routine_id        UUID NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    exercise_id       TEXT NOT NULL REFERENCES exercises(id),
    position          INT NOT NULL,
    target_sets       INT,
    target_reps       INT,
    target_duration_s INT,
    rest_s            INT,
    PRIMARY KEY (routine_id, exercise_id)
);

CREATE TABLE routine_tags (
    routine_id UUID NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    tag        TEXT NOT NULL,
    PRIMARY KEY (routine_id, tag)
);
CREATE INDEX idx_routine_tags_tag ON routine_tags(tag);

CREATE TABLE fitness_levels (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    scale_factor  NUMERIC(3,2) NOT NULL,
    description   TEXT,
    display_order INT NOT NULL
);

INSERT INTO fitness_levels (id, name, scale_factor, display_order) VALUES
    ('beginner',     'Baja forma',    0.70, 1),
    ('intermediate', 'En forma',      1.00, 2),
    ('advanced',     'Muy en forma',  1.30, 3)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE user_fitness_profile (
    user_id           UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    fitness_level_id  TEXT NOT NULL REFERENCES fitness_levels(id),
    goal              TEXT,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
