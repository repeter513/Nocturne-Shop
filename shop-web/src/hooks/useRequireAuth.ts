// Hook that redirects unauthenticated users to login. / Хук, перенаправляющий неавторизованных пользователей на страницу входа.
import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'

// Returns user and loading state; navigates to /login when unauthenticated. / Возвращает user и loading; перенаправляет на /login при отсутствии авторизации.
export function useRequireAuth() {
  const { user, loading } = useAuth()
  const navigate = useNavigate()

  // Routing guard: wait for session restore, then redirect if still no user. / Guard маршрута: ждём восстановление сессии, затем редирект если user всё ещё null.
  useEffect(() => {
    if (!loading && !user) navigate('/login')
  }, [user, loading, navigate])

  return { user, authLoading: loading }
}
