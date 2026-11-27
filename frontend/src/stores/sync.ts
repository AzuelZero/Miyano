import { onBeforeUnmount, ref } from 'vue'
import { defineStore } from 'pinia'

/**
 * Connectivity state for the offline banner (sync D2, design decision 8's
 * basic tier). D3 replaces this with the full five-state sync machine.
 */

export const useSyncStore = defineStore('sync', () => {
  const online = ref(navigator.onLine)
  // One-shot "De vuelta — sincronizado" toast: shown when the tab regains
  // connectivity, then cleared (AppShell renders it with a timeout).
  const backOnline = ref(false)

  function goOffline(): void {
    online.value = false
  }

  function goOnline(): void {
    const wasOffline = !online.value
    online.value = true
    if (wasOffline) backOnline.value = true
  }

  function acknowledgeBackOnline(): void {
    backOnline.value = false
  }

  /** Wire browser connectivity events once per app instance. */
  function listen(): void {
    window.addEventListener('offline', goOffline)
    window.addEventListener('online', goOnline)
    onBeforeUnmount(() => {
      window.removeEventListener('offline', goOffline)
      window.removeEventListener('online', goOnline)
    })
  }

  return { online, backOnline, goOffline, goOnline, acknowledgeBackOnline, listen }
})
