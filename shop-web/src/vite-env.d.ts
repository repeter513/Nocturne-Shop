// Vite environment type declarations. / Объявления типов окружения Vite.
/// <reference types="vite/client" />

// Environment variables exposed to the client. / Переменные окружения, доступные клиенту.
interface ImportMetaEnv {
  // Override API base URL (default: /api/v1 via dev proxy). / Переопределение базового URL API (по умолчанию: /api/v1 через dev proxy).
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
