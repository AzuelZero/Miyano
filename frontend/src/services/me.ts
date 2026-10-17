import { apiJson } from './api'
import type { FitnessLevel } from '../types'

export async function getFitnessLevel(): Promise<FitnessLevel> {
  return apiJson<FitnessLevel>('/me/fitness-level')
}
