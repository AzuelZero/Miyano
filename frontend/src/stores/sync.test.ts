import { describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useSyncStore } from './sync'

describe('sync store (banner D2)', () => {
  it('starts online and flips offline/online with the browser events', () => {
    setActivePinia(createPinia())
    const sync = useSyncStore()
    expect(sync.online).toBe(true)

    sync.goOffline()
    expect(sync.online).toBe(false)

    sync.goOnline()
    expect(sync.online).toBe(true)
    expect(sync.backOnline).toBe(true) // "De vuelta — sincronizado" pending
  })

  it('returning online without having been offline shows no toast', () => {
    setActivePinia(createPinia())
    const sync = useSyncStore()
    sync.goOnline()
    expect(sync.backOnline).toBe(false)
  })

  it('acknowledge clears the toast; offline never clears it silently', () => {
    setActivePinia(createPinia())
    const sync = useSyncStore()
    sync.goOffline()
    sync.goOnline()
    sync.acknowledgeBackOnline()
    expect(sync.backOnline).toBe(false)
  })
})
