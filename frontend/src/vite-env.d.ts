/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** API base URL. Default "/api/v1" (Vite dev proxy → localhost:8080). */
  readonly VITE_API_URL?: string
  /** Media base URL. Default "/media" (same origin → Go static route). */
  readonly VITE_MEDIA_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
