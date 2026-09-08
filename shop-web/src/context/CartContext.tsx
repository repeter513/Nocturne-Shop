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
  count: number
  refresh: () => Promise<void>
}

const CartContext = createContext<CartContextValue | null>(null)

export function CartProvider({ children }: { children: ReactNode }) {
  const { user, loading: authLoading } = useAuth()
  const [count, setCount] = useState(0)

  const refresh = useCallback(async () => {
    if (!user) {
      setCount(0)
      return
    }
    try {
      const cart = await api.getCart()
      setCount(cart.items.reduce((sum, item) => sum + item.quantity, 0))
    } catch {
      setCount(0)
    }
  }, [user])

  useEffect(() => {
    if (!authLoading) refresh()
  }, [authLoading, refresh])

  const value = useMemo(() => ({ count, refresh }), [count, refresh])

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

export function useCart() {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error('useCart outside CartProvider')
  return ctx
}
