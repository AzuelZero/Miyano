import type { Exercise } from '../../types'

/** Client-side catalog filtering — pure functions, fully unit-tested. */

export const PAGE_SIZE = 48

/** lowercase + accent-free comparison basis (es-EN search that forgives "jalon" vs "jalón"). */
export function normalize(text: string): string {
  return text
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .trim()
}

export interface CatalogFilter {
  query: string
  equipment: string // '' = all
}

export function filterExercises(list: Exercise[], filter: CatalogFilter): Exercise[] {
  const q = normalize(filter.query)
  if (q === '' && filter.equipment === '') return list

  return list.filter((ex) => {
    if (filter.equipment !== '' && ex.equipment !== filter.equipment) return false
    if (q === '') return true
    return normalize(ex.name).includes(q) || normalize(ex.body_part).includes(q)
  })
}

export interface EquipmentFacet {
  equipment: string
  count: number
}

/** Derives the equipment list (with counts) from the catalog itself — no backend needed. */
export function equipmentFacets(list: Exercise[]): EquipmentFacet[] {
  const counts = new Map<string, number>()
  for (const ex of list) {
    counts.set(ex.equipment, (counts.get(ex.equipment) ?? 0) + 1)
  }
  return [...counts.entries()]
    .map(([equipment, count]) => ({ equipment, count }))
    .sort((a, b) => b.count - a.count || a.equipment.localeCompare(b.equipment))
}
