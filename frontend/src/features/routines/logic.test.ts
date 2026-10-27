import { describe, expect, it } from 'vitest'
import type { Exercise } from '../../types'
import type { EditorItem } from './logic'
import {
  DEFAULT_SET_SECONDS,
  estimateTotalTimeS,
  formatDuration,
  newEditorItem,
  parseTags,
  toRoutineInput,
} from './logic'

const makeExercise = (id: string, name: string): Exercise => ({
  id,
  name,
  category: 'test',
  body_part: 'chest',
  equipment: 'body weight',
  is_bodyweight: true,
})

const item = (overrides: Partial<EditorItem> = {}): EditorItem => ({
  exerciseId: '0001',
  exerciseName: 'Push-up',
  sets: 3,
  reps: 10,
  restS: 60,
  ...overrides,
})

describe('estimateTotalTimeS', () => {
  it('mirrors the backend default of 45 s per set', () => {
    expect(DEFAULT_SET_SECONDS).toBe(45)
  })

  it('3 sets × 45 s (default) + 2 rests of 60 s = 255', () => {
    expect(estimateTotalTimeS([item({ sets: 3, restS: 60, durationS: undefined, reps: undefined })])).toBe(255)
  })

  it('uses a custom duration instead of the default', () => {
    expect(estimateTotalTimeS([item({ sets: 2, durationS: 30, restS: 30 })])).toBe(2 * 30 + 1 * 30)
    expect(estimateTotalTimeS([item({ sets: 1, durationS: 60, restS: 0 })])).toBe(60)
  })

  it('single set with no rest costs only the set time', () => {
    expect(estimateTotalTimeS([item({ sets: 1, restS: 0, durationS: undefined })])).toBe(45)
  })

  it('returns 0 for an empty list', () => {
    expect(estimateTotalTimeS([])).toBe(0)
  })

  it('sums across items', () => {
    const total = estimateTotalTimeS([item({ sets: 1, restS: 0, durationS: 30 }), item({ sets: 1, restS: 0, durationS: 60 })])
    expect(total).toBe(90)
  })
})

describe('formatDuration', () => {
  it('formats seconds below one minute', () => {
    expect(formatDuration(45)).toBe('45 s')
    expect(formatDuration(0)).toBe('0 s')
  })

  it('formats minutes without trailing 0s', () => {
    expect(formatDuration(60)).toBe('1m')
    expect(formatDuration(90)).toBe('1m 30s')
    expect(formatDuration(396)).toBe('6m 36s')
    expect(formatDuration(450)).toBe('7m 30s')
  })

  it('formats hours, omitting only trailing zero parts', () => {
    expect(formatDuration(3600)).toBe('1h')
    expect(formatDuration(3661)).toBe('1h 1m 1s')
  })
})

describe('parseTags', () => {
  it('splits, trims, drops empties and dedupes (first wins)', () => {
    expect(parseTags('pierna, casa,, pierna')).toEqual(['pierna', 'casa'])
    expect(parseTags('  pierna , casa ')).toEqual(['pierna', 'casa'])
  })

  it('returns an empty array for blank input', () => {
    expect(parseTags('')).toEqual([])
    expect(parseTags(' , , ')).toEqual([])
  })
})

describe('newEditorItem', () => {
  it('uses the editor defaults', () => {
    const editorItem = newEditorItem(makeExercise('0001', 'Push-up'))
    expect(editorItem).toEqual({
      exerciseId: '0001',
      exerciseName: 'Push-up',
      sets: 3,
      reps: 10,
      restS: 60,
    })
    expect(editorItem.durationS).toBeUndefined()
  })
})

describe('toRoutineInput', () => {
  it('blank description becomes undefined', () => {
    const input = toRoutineInput('Pierna', '   ', [], ['pierna'])
    expect(input.description).toBeUndefined()
  })

  it('keeps a trimmed description', () => {
    const input = toRoutineInput('Pierna', ' rutina de piernas ', [], [])
    expect(input.description).toBe('rutina de piernas')
  })

  it('omits reps and duration keys when undefined', () => {
    const input = toRoutineInput('Pierna', '', [item({ reps: undefined, durationS: undefined })], [])
    expect(input.exercises[0]).not.toHaveProperty('target_reps')
    expect(input.exercises[0]).not.toHaveProperty('target_duration_s')
    expect(input.exercises[0]).toEqual({ exercise_id: '0001', target_sets: 3, rest_s: 60 })
  })

  it('maps sets, reps, duration and rest when defined', () => {
    const input = toRoutineInput('Pierna', 'desc', [item({ sets: 4, reps: 12, durationS: 40, restS: 90 })], ['pierna'])
    expect(input.exercises[0]).toEqual({
      exercise_id: '0001',
      target_sets: 4,
      target_reps: 12,
      target_duration_s: 40,
      rest_s: 90,
    })
    expect(input.name).toBe('Pierna')
    expect(input.tags).toEqual(['pierna'])
  })
})
