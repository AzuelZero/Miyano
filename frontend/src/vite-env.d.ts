/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** API base URL. Default "/api/v1" (Vite dev proxy → localhost:8080). */
  readonly VITE_API_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
