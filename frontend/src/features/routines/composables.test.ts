import 'fake-indexeddb/auto'
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { clearRoutineMirrors, readRoutine, writeRoutine } from '../../lib/db'
import { useAuthStore } from '../../stores/auth'
import { routinesListQueryFn, routineDetailQueryFn } from './composables'
import type { RoutineDetail, RoutineSummary } from '../../types'

// Offline-mirror contract (sync D2, fork F3): fresh fetches mirror to
// IndexedDB; network failures answer from the mirror.

const summary = (id: string, name: string, tags: string[]): RoutineSummary => ({
  id,
  name,
  description: null,
  tags,
  revision: 1,
  exercise_count: 2,
  total_time_s: 450,
})

const detail = (id: string): RoutineDetail => ({
  id,
  name: 'Offline Pierna',
  description: null,
  tags: ['pierna'],
  revision: 3,
  exercises: [{ exercise_id: '0001', position: 1 }],
  level: { id: 'beginner', name: 'Baja forma', scale_factor: 0.7 },
  scaled_exercises: [{ exercise_id: '0001', position: 1 }],
  total_time_s: 450,
  total_time_scaled_s: 315,
})

/** JWT-shaped access token whose sub claim is the mirror's owner. */
function loginAs(userId: string): void {
  const payload = btoa(JSON.stringify({ sub: userId })).replace(/\+/g, '-').replace(/\//g, '_')
  const auth = useAuthStore()
  auth.$patch({ accessToken: `x.${payload}.y` })
}

/** Stub fetch: routes through the given handler or fails like an offline tab.
 * The api client uses relative URLs (`/api/v1/...`) — match that shape. */
function stubFetch(handler?: (path: string) => unknown): void {
  vi.stubGlobal('fetch', vi.fn(async (_url: string | URL) => {
    if (!handler) throw new TypeError('Failed to fetch')
    const raw = String(_url)
    const path = raw.includes('/api/v1') ? raw.slice(raw.indexOf('/api/v1') + '/api/v1'.length) : raw
    const body = handler(path)
    return new Response(JSON.stringify(body), { status: 200 })
  }))
}

beforeEach(async () => {
  setActivePinia(createPinia())
  localStorage.clear()
  await clearRoutineMirrors()
  loginAs('user-1')
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('routinesListQueryFn', () => {
  it('online: fetches fresh and mirrors the unfiltered list', async () => {
    stubFetch((path) => (path === '/routines' ? [summary('r1', 'Pierna', ['pierna'])] : []))
    const fresh = await routinesListQueryFn('')
    expect(fresh).toHaveLength(1)
    const mirror = await readRoutine('list')
    expect(mirror?.userId).toBe('user-1')
    expect((mirror?.data as RoutineSummary[])[0]!.id).toBe('r1')
  })

  it('offline: answers from the mirror, filtering by tag client-side', async () => {
    await writeRoutine('list', 'user-1', [
      summary('r1', 'Pierna', ['pierna']),
      summary('r2', 'Brazo', ['brazo']),
    ])
    stubFetch() // everything fails: offline

    expect(await routinesListQueryFn('')).toHaveLength(2)
    const filtered = await routinesListQueryFn('pierna')
    expect(filtered.map((r) => r.id)).toEqual(['r1'])
  })

  it('offline without mirror: throws (honest error, no fake empty)', async () => {
    stubFetch()
    await expect(routinesListQueryFn('')).rejects.toThrow()
  })

  it('another user’s mirror is not served (owner guard, fork F5 baseline)', async () => {
    await writeRoutine('list', 'user-2', [summary('r9', 'Ajena', ['pierna'])])
    stubFetch()
    await expect(routinesListQueryFn('')).rejects.toThrow()
  })
})

describe('routineDetailQueryFn', () => {
  it('offline: answers the mirrored detail of a routine seen before', async () => {
    await writeRoutine('detail:user-1:rt-1', 'user-1', detail('rt-1'))
    stubFetch()
    const got = await routineDetailQueryFn('rt-1')
    expect(got.id).toBe('rt-1')
    expect(got.level.scale_factor).toBe(0.7)
    expect(got.revision).toBe(3)
  })

  it('offline unseen detail: throws honestly', async () => {
    stubFetch()
    await expect(routineDetailQueryFn('ghost')).rejects.toThrow()
  })
})
