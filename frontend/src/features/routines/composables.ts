import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { RoutineInput } from '../../types'
import { createRoutine, deleteRoutine, getRoutine, listRoutines, updateRoutine } from './services'

/** Routine list, optionally filtered by one exact tag. */
export function useRoutines(tag = '') {
  return useQuery({
    // '' normalizes to null so the unfiltered call shares one cache entry
    // with views that key ['routines', null] directly.
    queryKey: ['routines', tag === '' ? null : tag],
    queryFn: () => listRoutines(tag === '' ? undefined : tag),
  })
}

export function useRoutine(id: string) {
  return useQuery({
    queryKey: ['routine', id],
    queryFn: () => getRoutine(id),
  })
}

export function useCreateRoutine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: RoutineInput) => createRoutine(input),
    onSuccess: (detail) => {
      queryClient.invalidateQueries({ queryKey: ['routines'] })
      queryClient.invalidateQueries({ queryKey: ['routine', detail.id] })
    },
  })
}

export function useUpdateRoutine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (vars: { id: string; input: RoutineInput }) => updateRoutine(vars.id, vars.input),
    onSuccess: (_detail, vars) => {
      queryClient.invalidateQueries({ queryKey: ['routines'] })
      queryClient.invalidateQueries({ queryKey: ['routine', vars.id] })
    },
  })
}

export function useDeleteRoutine() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => deleteRoutine(id),
    onSuccess: (_result, id) => {
      queryClient.invalidateQueries({ queryKey: ['routines'] })
      queryClient.invalidateQueries({ queryKey: ['routine', id] })
    },
  })
}
