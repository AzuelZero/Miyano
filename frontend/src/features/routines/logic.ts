import type { Exercise, RoutineInput } from '../../types'

/** Pure routine-editor logic — mirrors backend/internal/domain/routine.go. */

/** Estimate for one reps-based set when no explicit duration is given (Go DefaultSetSeconds). */
export const DEFAULT_SET_SECONDS = 45

/** Mirror of the seeded fitness levels (labels are UI copy, hence Spanish). */
export const FITNESS_LEVELS = [
  { id: 'beginner', label: 'Baja forma' },
  { id: 'intermediate', label: 'En forma' },
  { id: 'advanced', label: 'Muy en forma' },
]

export interface EditorItem {
  exerciseId: string
  exerciseName: string
  sets: number
  reps?: number
  durationS?: number
  restS: number
}

/** Mirror of TotalEstimatedSeconds: sets × (duration or default) + (sets−1) × rest, per item. */
export function estimateTotalTimeS(items: EditorItem[]): number {
  let total = 0
  for (const item of items) {
    const sets = item.sets > 0 ? item.sets : 1
    const duration = item.durationS !== undefined && item.durationS > 0 ? item.durationS : DEFAULT_SET_SECONDS
    const rest = item.restS > 0 ? item.restS : 0
    total += sets * duration + (sets - 1) * rest
  }
  return total
}

/** '45 s' | '1m 30s' | '1h 5m 30s' — only trailing zero parts are omitted. */
export function formatDuration(totalS: number): string {
  if (totalS < 60) return `${totalS} s`
  const hours = Math.floor(totalS / 3600)
  const minutes = Math.floor((totalS % 3600) / 60)
  const seconds = totalS % 60
  const parts: Array<[number, string]> = [
    [hours, 'h'],
    [minutes, 'm'],
    [seconds, 's'],
  ]
  if (hours === 0) parts.shift()
  while (parts.length > 1 && parts[parts.length - 1][0] === 0) parts.pop()
  return parts.map(([value, unit]) => `${value}${unit}`).join(' ')
}

/** Splits a comma-separated tag string: trimmed, non-empty, exact-deduped (first wins). */
export function parseTags(raw: string): string[] {
  const tags = new Set<string>()
  for (const part of raw.split(',')) {
    const tag = part.trim()
    if (tag !== '') tags.add(tag)
  }
  return [...tags]
}

export function newEditorItem(exercise: Exercise): EditorItem {
  return {
    exerciseId: exercise.id,
    exerciseName: exercise.name,
    sets: 3,
    reps: 10,
    restS: 60,
  }
}

/** Builds the API payload: blank description → undefined; optional targets only when defined. */
export function toRoutineInput(name: string, description: string, items: EditorItem[], tags: string[]): RoutineInput {
  const trimmed = description.trim()
  return {
    name,
    description: trimmed === '' ? undefined : trimmed,
    exercises: items.map((item) => ({
      exercise_id: item.exerciseId,
      target_sets: item.sets,
      ...(item.reps !== undefined ? { target_reps: item.reps } : {}),
      ...(item.durationS !== undefined ? { target_duration_s: item.durationS } : {}),
      rest_s: item.restS,
    })),
    tags,
  }
}
