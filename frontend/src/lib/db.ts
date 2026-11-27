import { openDB, type IDBPDatabase } from 'idb'

/**
 * IndexedDB mirror for offline reads (sync D2, design decision 5) and the
 * outbox store D3 will drain. localStorage stays for tokens (fork F5).
 *
 * The interface is deliberately serializable: it is the SyncStorage contract
 * Capacitor (Preferences/Filesystem) will implement later.
 */

export const DB_NAME = 'miyano'
export const DB_VERSION = 1

export const STORE_CATALOG = 'catalog'
export const STORE_ROUTINES = 'routines_mirror'
export const STORE_OUTBOX = 'outbox'

/** Catalog mirror: the full exercise list plus when it was fetched. */
export interface CatalogMirror {
  /** What a catalog fetch returned, kept as-is (read-only server data). */
  data: unknown
  savedAt: number
}

/** Routines mirror: one entry per fetched payload (list or detail). */
export interface RoutineMirror {
  /** Owner of the data (higiene multi-usuario, fork F5 — guard lands in D3). */
  userId: string
  data: unknown
  savedAt: number
}

let dbPromise: Promise<IDBPDatabase> | null = null

function getDB(): Promise<IDBPDatabase> {
  dbPromise ??= openDB(DB_NAME, DB_VERSION, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORE_CATALOG)) {
        db.createObjectStore(STORE_CATALOG)
      }
      if (!db.objectStoreNames.contains(STORE_ROUTINES)) {
        db.createObjectStore(STORE_ROUTINES)
      }
      if (!db.objectStoreNames.contains(STORE_OUTBOX)) {
        db.createObjectStore(STORE_OUTBOX, { keyPath: 'id' })
      }
    },
  })
  return dbPromise
}

export async function readCatalog(key = 'snapshot'): Promise<CatalogMirror | undefined> {
  const db = await getDB()
  return db.get(STORE_CATALOG, key)
}

export async function writeCatalog(data: unknown, key = 'snapshot'): Promise<void> {
  const db = await getDB()
  await db.put(STORE_CATALOG, { data, savedAt: Date.now() } satisfies CatalogMirror, key)
}

export async function readRoutine(key: string): Promise<RoutineMirror | undefined> {
  const db = await getDB()
  return db.get(STORE_ROUTINES, key)
}

export async function writeRoutine(key: string, userId: string, data: unknown): Promise<void> {
  const db = await getDB()
  await db.put(STORE_ROUTINES, { userId, data, savedAt: Date.now() } satisfies RoutineMirror, key)
}

/** Wipes the routine mirror — the hygiene step when ANOTHER user signs in
 * (fork F5; also resets state between tests). Catalog/outbox untouched. */
export async function clearRoutineMirrors(): Promise<void> {
  const db = await getDB()
  await db.clear(STORE_ROUTINES)
}
