/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  test: {
    // Simulate a browser DOM in Node so component tests can render React.
    environment: 'jsdom',
    // Allow describe/it/expect without importing them in every test file.
    globals: true,
    // Runs once before the suite: registers jest-dom's custom matchers.
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
