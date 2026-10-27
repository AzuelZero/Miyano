/** Shared API types — mirror of the Go DTOs (ADR D009: names in English). */

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
}

export interface FitnessLevel {
  id: string
  name: string
  scale_factor: number
}

export interface Exercise {
  id: string
  name: string
  category: string
  body_part: string
  equipment: string
  force_type?: string
  difficulty?: string
  is_bodyweight: boolean
  image_url?: string
  gif_url?: string
}

export interface Translation {
  lang: string
  instructions?: string
  instruction_steps?: string[]
}

export interface ExerciseDetail extends Exercise {
  translations: Translation[]
}

export interface RoutineExercise {
  exercise_id: string
  position: number
  target_sets?: number
  target_reps?: number
  target_duration_s?: number
  rest_s?: number
}

export interface RoutineSummary {
  id: string
  name: string
  description?: string | null
  tags: string[]
  exercise_count: number
  total_time_s: number
}

export interface RoutineDetail {
  id: string
  name: string
  description?: string | null
  tags: string[]
  exercises: RoutineExercise[]
  level: FitnessLevel
  scaled_exercises: RoutineExercise[]
  total_time_s: number
  total_time_scaled_s: number
}

export interface RoutineExerciseInput {
  exercise_id: string
  target_sets?: number
  target_reps?: number
  target_duration_s?: number
  rest_s?: number
}

export interface RoutineInput {
  name: string
  description?: string
  exercises: RoutineExerciseInput[]
  tags: string[]
}
