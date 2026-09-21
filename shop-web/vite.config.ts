// Vite build configuration for the shop web frontend. / Конфигурация сборки Vite для веб-интерфейса магазина.
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Dev server with API proxy to the gateway. / Dev-сервер с прокси API на шлюз.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      // Forward /api/v1/* to BFF on :8090 — avoids CORS in dev without VITE_API_BASE_URL. / Проксируем /api/v1/* на BFF :8090 — без CORS в dev без VITE_API_BASE_URL.
      '/api/v1': {
        target: 'http://localhost:8090',
        changeOrigin: true,
      },
    },
  },
})
