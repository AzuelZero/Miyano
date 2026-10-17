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
