import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import type { Cart } from '../api/types'
import { CartSkeleton } from '../components/Skeleton'
import { useCart } from '../context/CartContext'
import { useToast } from '../context/ToastContext'
import { useRequireAuth } from '../hooks/useRequireAuth'

export function CartPage() {
  const { user, authLoading } = useRequireAuth()
  const { toast } = useToast()
  const { refresh } = useCart()
  const navigate = useNavigate()
  const [cart, setCart] = useState<Cart | null>(null)
  const [loading, setLoading] = useState(true)
  const [ordering, setOrdering] = useState(false)
  const [clearing, setClearing] = useState(false)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setCart(await api.getCart())
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (authLoading || !user) return
    load()
  }, [user, authLoading, load])

  const syncCart = async (fn: () => Promise<Cart>) => {
    const next = await fn()
    setCart(next)
    await refresh()
    return next
  }

  const updateQty = async (productId: string, quantity: number) => {
    try {
      await syncCart(() => api.updateCartItem(productId, quantity))
      toast('Количество обновлено', 'info')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка'
      setError(msg)
      toast(msg, 'error')
    }
  }

  const remove = async (productId: string) => {
    try {
      await syncCart(() => api.removeFromCart(productId))
      toast('Товар удалён из корзины', 'info')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка'
      setError(msg)
      toast(msg, 'error')
    }
  }

  const clear = async () => {
    if (!window.confirm('Очистить корзину?')) return
    setClearing(true)
    setError('')
    try {
      await api.clearCart()
      setCart({ userId: cart?.userId ?? '', items: [], totalPrice: 0 })
      await refresh()
      toast('Корзина очищена')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка'
      setError(msg)
      toast(msg, 'error')
    } finally {
      setClearing(false)
    }
  }

  const checkout = async () => {
    setOrdering(true)
    setError('')
    try {
      const order = await api.createOrder()
      toast('Заказ оформлен — у вас 15 минут на оплату')
      navigate(`/orders/${order.orderId}`)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка оформления'
      setError(msg)
      toast(msg, 'error')
    } finally {
      setOrdering(false)
    }
  }

  if (authLoading || loading) {
    return (
      <div className="page">
        <h1>Корзина</h1>
        <CartSkeleton />
      </div>
    )
  }

  const items = cart?.items ?? []

  return (
    <div className="page">
      <h1>Корзина</h1>
      {error && <div className="alert alert-error">{error}</div>}

      {items.length === 0 ? (
        <div className="empty">
          Корзина пуста. <Link to="/">Перейти в каталог</Link>
        </div>
      ) : (
        <>
          <div className="cart-list">
            {items.map((item) => (
              <div key={item.productId} className="cart-item">
                <div className="cart-item-info">
                  <Link to={`/products/${item.productId}`} className="cart-item-name">
                    {item.name}
                  </Link>
                  <span className="cart-item-price">{item.price.toFixed(2)} ₽</span>
                </div>
                <div className="cart-item-actions">
                  <div className="quantity-control">
                    <button
                      type="button"
                      className="btn btn-ghost btn-sm"
                      disabled={item.quantity <= 1}
                      onClick={() => updateQty(item.productId, item.quantity - 1)}
                    >
                      −
                    </button>
                    <span className="quantity-value">{item.quantity}</span>
                    <button
                      type="button"
                      className="btn btn-ghost btn-sm"
                      onClick={() => updateQty(item.productId, item.quantity + 1)}
                    >
                      +
                    </button>
                  </div>
                  <span className="cart-item-total">
                    {(item.price * item.quantity).toFixed(2)} ₽
                  </span>
                  <button
                    type="button"
                    className="btn btn-ghost btn-sm"
                    onClick={() => remove(item.productId)}
                  >
                    ✕
                  </button>
                </div>
              </div>
            ))}
          </div>

          <div className="cart-summary">
            <div className="cart-total">
              Итого: <strong>{cart?.totalPrice.toFixed(2)} ₽</strong>
            </div>
            <div className="cart-actions">
              <button
                type="button"
                className="btn btn-ghost"
                disabled={clearing}
                onClick={clear}
              >
                {clearing ? 'Очистка...' : 'Очистить корзину'}
              </button>
              <button
                type="button"
                className="btn btn-primary"
                disabled={ordering}
                onClick={checkout}
              >
                {ordering ? 'Резервирование...' : 'Оформить заказ'}
              </button>
            </div>
          </div>
          <p className="cart-hint">
            После оформления товары резервируются на 15 минут — оплатить можно на странице заказа.
          </p>
        </>
      )}
    </div>
  )
}
