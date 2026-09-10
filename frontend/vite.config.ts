/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      // Dev-only: forward the API call to the backend so the browser sees a
      // same-origin request and the backend needs no CORS handling.
      '/calculate': 'http://localhost:8080',
    },
  },
  test: {
    // A browser-like DOM in Node so component tests can render React.
    environment: 'jsdom',
    // describe/it/expect without importing them in every file.
    globals: true,
    setupFiles: './src/test/setup.ts',
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/main.tsx', 'src/vite-env.d.ts', 'src/**/*.test.{ts,tsx}'],
    },
  },
})
