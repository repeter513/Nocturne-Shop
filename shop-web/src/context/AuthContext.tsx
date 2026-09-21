// Authentication context — login, register, logout, session restore. / Контекст аутентификации — вход, регистрация, выход, восстановление сессии.
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { api } from '../api/client'
import type { User } from '../api/types'

// Auth state and actions exposed to consumers. / Состояние auth и действия, доступные потребителям.
interface AuthState {
  // Current user profile or null when logged out. / Профиль текущего пользователя или null при выходе.
  user: User | null
  // True while restoring session from localStorage on mount. / True пока восстанавливается сессия из localStorage при монтировании.
  loading: boolean
  // Authenticates with email/password, persists tokens. / Аутентификация по email/паролю, сохранение токенов.
  login: (email: string, password: string) => Promise<void>
  // Creates account then auto-logs in. / Создаёт аккаунт и автоматически входит.
  register: (email: string, password: string) => Promise<void>
  // Clears tokens and user state. / Очищает токены и состояние пользователя.
  logout: () => void
}

const AuthContext = createContext<AuthState | null>(null)

// localStorage keys for JWT token persistence across page reloads. / Ключи localStorage для сохранения JWT между перезагрузками.
const TOKEN_KEY = 'shop_access_token'
const REFRESH_KEY = 'shop_refresh_token'

// Saves tokens to localStorage and sets API client token. / Сохраняет токены в localStorage и устанавливает токен API-клиента.
function persistTokens(access: string, refresh: string) {
  localStorage.setItem(TOKEN_KEY, access)
  localStorage.setItem(REFRESH_KEY, refresh)
  api.setToken(access)
}

// Removes tokens from localStorage and clears API client token. / Удаляет токены из localStorage и сбрасывает токен API-клиента.
function clearTokens() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
  api.setToken(null)
}

// AuthProvider restores session on mount and exposes auth actions. / AuthProvider восстанавливает сессию при монтировании и предоставляет auth-действия.
export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  // Restores user from stored token, refreshing if expired. / Восстанавливает пользователя из сохранённого токена, обновляя при истечении.
  const loadUser = useCallback(async () => {
    const token = localStorage.getItem(TOKEN_KEY)
    if (!token) {
      setLoading(false)
      return
    }
    api.setToken(token)
    try {
      // GET /auth/me — validates JWT and returns profile. / GET /auth/me — валидирует JWT и возвращает профиль.
      setUser(await api.me())
    } catch {
      const refresh = localStorage.getItem(REFRESH_KEY)
      if (refresh) {
        try {
          // Access token expired — exchange refresh token for new pair. / Access token истёк — обмениваем refresh token на новую пару.
          const tokens = await api.refresh(refresh)
          persistTokens(tokens.accessToken, tokens.refreshToken)
          setUser(await api.me())
          return
        } catch {
          clearTokens()
        }
      }
      clearTokens()
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  // Run session restore once on app mount. / Восстановление сессии один раз при монтировании приложения.
  useEffect(() => {
    loadUser()
  }, [loadUser])

  const login = useCallback(async (email: string, password: string) => {
    const tokens = await api.login(email, password)
    persistTokens(tokens.accessToken, tokens.refreshToken)
    setUser(await api.me())
  }, [])

  const register = useCallback(async (email: string, password: string) => {
    await api.register(email, password)
    await login(email, password)
  }, [login])

  const logout = useCallback(() => {
    clearTokens()
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ user, loading, login, register, logout }),
    [user, loading, login, register, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// Hook to access auth state; throws if used outside AuthProvider. / Хук для доступа к auth; выбрасывает ошибку вне AuthProvider.
export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth outside AuthProvider')
  return ctx
}
