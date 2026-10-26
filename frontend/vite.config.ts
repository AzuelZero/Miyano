/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  server: {
    proxy: {
      // Frontend calls relative /api/v1/... — dev lands on the local Go API.
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
      // Exercise media served by the Go static route (same origin as prod).
      '/media': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
  test: {
    environment: 'jsdom',
  },
})
