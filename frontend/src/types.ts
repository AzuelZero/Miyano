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
