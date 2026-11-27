/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    // Offline shell (sync D2, design decision 4): the full build is precached
    // and each deploy ships a fresh SW — cache is named per build, so a deploy
    // swaps workers instead of fighting stale entries.
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.svg'],
      manifest: {
        name: 'Miyano',
        short_name: 'Miyano',
        description: 'Tu entrenamiento, tu ritmo',
        lang: 'es',
        display: 'standalone',
        theme_color: '#080616',
        background_color: '#080616',
        icons: [{ src: 'icons/app-icon-placeholder.png', sizes: '512x512', type: 'image/png' }],
      },
    }),
  ],
  server: {
    proxy: {
      // Frontend calls relative /api/v1/... — dev lands on the local Go API.
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      // Exercise media served by the Go static route (same origin as prod).
      '/media': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
  // Same proxy for `pnpm preview` (the offline E2E runs against the built
  // shell + REAL service worker, design decision 4 — `pnpm dev` never does).
  preview: {
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      '/media': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
  test: {
    environment: 'jsdom',
  },
})
