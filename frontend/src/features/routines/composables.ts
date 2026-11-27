import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query'
import type { RoutineDetail, RoutineInput, RoutineSummary } from '../../types'
import { createRoutine, deleteRoutine, getRoutine, listRoutines, updateRoutine } from './services'
import { useAuthStore } from '../../stores/auth'
import { readRoutine, writeRoutine } from '../../lib/db'

function currentUserId(): string {
  return useAuthStore().userId ?? 'unknown'
}

/** QueryFn of the routine list: fresh fetch mirrored on success, stale
 * mirror on network failure (replicating the server's exact-tag filter
 * client-side — the only semantic the endpoint has). A mirror owned by
 * ANOTHER user reads as absent (fork F5 baseline; the warned cleanup of D3
 * sits on top of this). */
export async function routinesListQueryFn(tag = ''): Promise<RoutineSummary[]> {
  try {
    const fresh = await listRoutines(tag === '' ? undefined : tag)
    if (tag === '') void writeRoutine('list', currentUserId(), fresh)
    return fresh
  } catch (err) {
    const mirror = await readRoutine('list')
    if (mirror === undefined || mirror.userId !== currentUserId()) throw err
    const data = mirror.data
    if (!Array.isArray(data)) throw err
    const summaries = data as RoutineSummary[]
    return tag === '' ? summaries : summaries.filter((r) => r.tags.includes(tag))
  }
}

/** QueryFn of one routine's detail: same mirror contract as the list. */
export async function routineDetailQueryFn(id: string): Promise<RoutineDetail> {
  const detailKey = `detail:${currentUserId()}:${id}`
  try {
    const fresh = await getRoutine(id)
    void writeRoutine(detailKey, currentUserId(), fresh)
    return fresh
  } catch (err) {
    const mirror = await readRoutine(detailKey)
    if (mirror === undefined || mirror.userId !== currentUserId()) throw err
    return mirror.data as RoutineDetail
  }
}

/**
 * Routine list, optionally filtered by one exact tag. Offline reads (sync
 * D2): the unfiltered fetch mirrors to IndexedDB; when the network fails the
 * mirror answers (stale-while-revalidate by hand — fork F3).
 */
export function useRoutines(tag = '') {
  return useQuery({
    // '' normalizes to null so the unfiltered call shares one cache entry
    // with views that key ['routines', null] directly.
    queryKey: ['routines', tag === '' ? null : tag],
    queryFn: () => routinesListQueryFn(tag === '' ? undefined : tag),
  })
}

export function useRoutine(id: string) {
  return useQuery({
    queryKey: ['routine', id],
    queryFn: () => routineDetailQueryFn(id),
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
