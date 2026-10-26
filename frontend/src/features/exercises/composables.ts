import { useQuery } from '@tanstack/vue-query'
import { getExercise, listExercises } from './services'

/** Full catalog, cached once by TanStack Query (filtering happens client-side). */
export function useExercises() {
  return useQuery({
    queryKey: ['exercises'],
    queryFn: () => listExercises(),
  })
}

export function useExerciseDetail(id: string) {
  return useQuery({
    queryKey: ['exercise', id],
    queryFn: () => getExercise(id),
  })
}
