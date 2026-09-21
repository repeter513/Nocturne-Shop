// Cart badge context — item count synced with server cart. / Контекст бейджа корзины — количество позиций, синхронизированное с сервером.
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
import { useAuth } from './AuthContext'

interface CartContextValue {
  // Total item quantity across all cart lines (for header badge). / Общее количество позиций во всех строках корзины (для бейджа в шапке).
  count: number
  // Re-fetches cart from server and updates count. / Повторно загружает корзину с сервера и обновляет счётчик.
  refresh: () => Promise<void>
}

const CartContext = createContext<CartContextValue | null>(null)

// CartProvider tracks total item count for the header badge. / CartProvider отслеживает общее количество позиций для бейджа в шапке.
export function CartProvider({ children }: { children: ReactNode }) {
  const { user, loading: authLoading } = useAuth()
  const [count, setCount] = useState(0)

  // Fetches cart from API and updates item count. / Загружает корзину из API и обновляет счётчик позиций.
  const refresh = useCallback(async () => {
    if (!user) {
      setCount(0)
      return
    }
    try {
      // GET /cart — requires JWT; sums item.quantity for badge. / GET /cart — требует JWT; суммируем item.quantity для бейджа.
      const cart = await api.getCart()
      setCount(cart.items.reduce((sum, item) => sum + item.quantity, 0))
    } catch {
      setCount(0)
    }
  }, [user])

  // Refresh cart count after auth state settles (login/logout/session restore). / Обновляем счётчик после стабилизации auth (вход/выход/восстановление).
  useEffect(() => {
    if (!authLoading) refresh()
  }, [authLoading, refresh])

  const value = useMemo(() => ({ count, refresh }), [count, refresh])

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

// Hook to access cart count and refresh; throws outside CartProvider. / Хук для счётчика корзины и refresh; выбрасывает ошибку вне CartProvider.
export function useCart() {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error('useCart outside CartProvider')
  return ctx
}
