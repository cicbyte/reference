import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {fileURLToPath, URL} from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },

  // ---- Wails dev-mode requirements ----
  server: {
    // Vite must listen on all interfaces so the Wails webview process
    // (which loads assets over http) can reach the dev server.
    host: '0.0.0.0',
    // HMR over websocket; the port must match what Wails auto-detects.
    hmr: {
      port: 5173,
    },
    // Windows / network drives sometimes miss native fs events — polling
    // guarantees file changes are picked up and HMR fires reliably.
    watch: {
      usePolling: true,
      interval: 100,
    },
  },
})
