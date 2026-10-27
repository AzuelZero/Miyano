import { apiJson } from './api'
import type { FitnessLevel } from '../types'

export async function getFitnessLevel(): Promise<FitnessLevel> {
  return apiJson<FitnessLevel>('/me/fitness-level')
}

export async function putFitnessLevel(levelId: string): Promise<FitnessLevel> {
  return apiJson<FitnessLevel>('/me/fitness-level', {
    method: 'PUT',
    body: JSON.stringify({ level_id: levelId }),
  })
}
