import { describe, expect, it } from 'vitest'
import type { Exercise } from '../../types'
import { equipmentFacets, filterExercises, normalize, PAGE_SIZE } from './filtering'

const ex = (id: string, name: string, equipment: string, bodyPart = 'chest'): Exercise => ({
  id,
  name,
  category: 'test',
  body_part: bodyPart,
  equipment,
  is_bodyweight: equipment === 'body weight',
})

const catalog = [
  ex('0001', 'Push-up', 'body weight'),
  ex('0002', 'Jalón al pecho', 'machine', 'back'),
  ex('0003', 'Bench press', 'barbell'),
  ex('0004', 'Push-up decline', 'body weight'),
]

describe('normalize', () => {
  it('folds accents and case', () => {
    expect(normalize('Jalón AL Pecho')).toBe('jalon al pecho')
    expect(normalize('  Push-up ')).toBe('push-up')
  })
})

describe('filterExercises', () => {
  it('returns everything with empty filter', () => {
    expect(filterExercises(catalog, { query: '', equipment: '' })).toHaveLength(4)
  })

  it('searches name and body_part, accent/case-insensitive', () => {
    expect(filterExercises(catalog, { query: 'jalon', equipment: '' })).toHaveLength(1)
    expect(filterExercises(catalog, { query: 'PUSH', equipment: '' })).toHaveLength(2)
    expect(filterExercises(catalog, { query: 'back', equipment: '' })).toHaveLength(1)
  })

  it('filters by exact equipment', () => {
    const out = filterExercises(catalog, { query: '', equipment: 'body weight' })
    expect(out.map((e) => e.id)).toEqual(['0001', '0004'])
  })

  it('combines query and equipment', () => {
    expect(filterExercises(catalog, { query: 'decline', equipment: 'body weight' })).toHaveLength(1)
  })

  it('no matches returns empty array', () => {
    expect(filterExercises(catalog, { query: 'zzz', equipment: '' })).toEqual([])
  })
})

describe('equipmentFacets', () => {
  it('derives equipment with counts, sorted by count desc', () => {
    const facets = equipmentFacets(catalog)
    expect(facets[0]).toEqual({ equipment: 'body weight', count: 2 })
    expect(facets).toHaveLength(3)
  })
})

describe('PAGE_SIZE', () => {
  it('is a sane batch (~48 cards per load)', () => {
    expect(PAGE_SIZE).toBeGreaterThan(24)
    expect(PAGE_SIZE).toBeLessThanOrEqual(96)
  })
})
