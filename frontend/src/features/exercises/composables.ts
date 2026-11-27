import { useQuery } from '@tanstack/vue-query'
import type { Exercise, ExerciseDetail } from '../../types'
import { getExercise, listExercises } from './services'
import { readCatalog, writeCatalog } from '../../lib/db'

/**
 * Full catalog, cached once by TanStack Query (filtering happens client-side).
 * Offline reads (sync D2): a successful fetch mirrors to IndexedDB; when the
 * network fails, the stale mirror answers instead of an error
 * (stale-while-revalidate by hand — fork F3).
 */
export function useExercises() {
  return useQuery({
    queryKey: ['exercises'],
    queryFn: async () => {
      try {
        const fresh = await listExercises()
        void writeCatalog(fresh)
        return fresh
      } catch (err) {
        const mirror = await readCatalog()
        if (mirror !== undefined) return mirror.data as Exercise[]
        throw err
      }
    },
  })
}

export function useExerciseDetail(id: string) {
  return useQuery({
    queryKey: ['exercise', id],
    queryFn: async () => {
      const detailKey = `detail:${id}`
      try {
        const fresh = await getExercise(id)
        void writeCatalog(fresh, detailKey)
        return fresh
      } catch (err) {
        // Full detail mirrored from a previous fetch; else the catalog's
        // base entry (name/GIF without translations) rather than nothing.
        const detail = await readCatalog(detailKey)
        if (detail !== undefined) return detail.data as ExerciseDetail
        const mirror = await readCatalog()
        const fallback = mirror?.data
        if (Array.isArray(fallback)) {
          const hit = (fallback as { id: string }[]).find((exercise) => exercise.id === id)
          if (hit !== undefined) return hit as ExerciseDetail
        }
        throw err
      }
    },
  })
}
