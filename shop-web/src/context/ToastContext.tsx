// Toast notification context — ephemeral success / error / info messages. / Контекст toast-уведомлений — кратковременные success / error / info сообщения.
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import '../components/Toast.css'

type ToastType = 'success' | 'error' | 'info'

interface Toast {
  // Unique ID for list key and auto-dismiss filter. / Уникальный ID для key списка и фильтра авто-скрытия.
  id: number
  // User-visible message text. / Текст сообщения для пользователя.
  message: string
  // Visual variant (controls CSS class toast-{type}). / Визуальный вариант (CSS-класс toast-{type}).
  type: ToastType
}

interface ToastContextValue {
  // Shows a toast that auto-dismisses after TOAST_DURATION_MS. / Показывает toast с авто-скрытием через TOAST_DURATION_MS.
  toast: (message: string, type?: ToastType) => void
}

const ToastContext = createContext<ToastContextValue | null>(null)

const TOAST_DURATION_MS = 4500

// ToastProvider renders a toast stack and exposes the toast() function. / ToastProvider отображает стек toast и предоставляет функцию toast().
export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const toast = useCallback((message: string, type: ToastType = 'success') => {
    const id = Date.now() + Math.floor(Math.random() * 1000)
    setToasts((prev) => [...prev, { id, message, type }])
    // Auto-remove after duration — no manual dismiss needed. / Авто-удаление через duration — ручное закрытие не нужно.
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, TOAST_DURATION_MS)
  }, [])

  const value = useMemo(() => ({ toast }), [toast])

  return (
    <ToastContext.Provider value={value}>
      {children}
      {/* aria-live="polite" announces new toasts to screen readers. / aria-live="polite" озвещает новые toast для screen reader. */}
      <div className="toast-container" aria-live="polite" aria-relevant="additions">
        {toasts.map((t) => (
          <div key={t.id} className={`toast toast-${t.type}`} role="status">
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

// Hook to show toasts; throws if used outside ToastProvider. / Хук для показа toast; выбрасывает ошибку вне ToastProvider.
export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast outside ToastProvider')
  return ctx
}
