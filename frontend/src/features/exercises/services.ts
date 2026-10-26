import { apiJson } from '../../services/api'
import type { Exercise, ExerciseDetail } from '../../types'

export async function listExercises(equipment = ''): Promise<Exercise[]> {
  const suffix = equipment ? `?equipment=${encodeURIComponent(equipment)}` : ''
  return apiJson<Exercise[]>(`/exercises${suffix}`)
}

export async function getExercise(id: string): Promise<ExerciseDetail> {
  return apiJson<ExerciseDetail>(`/exercises/${id}`)
}

/** Media base URL: same origin in dev (Vite proxy → Go static route /media). */
export const MEDIA_URL = import.meta.env.VITE_MEDIA_URL ?? '/media'

export function mediaSrc(path?: string): string | null {
  if (!path) return null
  return `${MEDIA_URL}/${path.replace(/^\//, '')}`
}
