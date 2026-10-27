import { apiJson } from '../../services/api'
import type { RoutineDetail, RoutineInput, RoutineSummary } from '../../types'

export async function listRoutines(tag = ''): Promise<RoutineSummary[]> {
  const suffix = tag ? `?tag=${encodeURIComponent(tag)}` : ''
  return apiJson<RoutineSummary[]>(`/routines${suffix}`)
}

export async function getRoutine(id: string): Promise<RoutineDetail> {
  return apiJson<RoutineDetail>(`/routines/${id}`)
}

export async function createRoutine(input: RoutineInput): Promise<RoutineDetail> {
  return apiJson<RoutineDetail>('/routines', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function updateRoutine(id: string, input: RoutineInput): Promise<RoutineDetail> {
  return apiJson<RoutineDetail>(`/routines/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  })
}

export async function deleteRoutine(id: string): Promise<void> {
  return apiJson<void>(`/routines/${id}`, { method: 'DELETE' })
}
