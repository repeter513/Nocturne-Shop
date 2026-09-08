import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Order, Payment } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { PAYMENT_RESERVE_MS, PAYMENT_RESERVE_MINUTES } from '../config'
import { OrderListSkeleton } from '../components/Skeleton'
import { useCart } from '../context/CartContext'
import { useToast } from '../context/ToastContext'
import { useRequireAuth } from '../hooks/useRequireAuth'

function formatCountdown(ms: number): string {
  const totalSec = Math.max(0, Math.floor(ms / 1000))
  const min = Math.floor(totalSec / 60)
  const sec = totalSec % 60
  return `${min}:${sec.toString().padStart(2, '0')}`
}

export function OrderDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user, authLoading } = useRequireAuth()
  const { refresh } = useCart()
  const { toast } = useToast()
  const [order, setOrder] = useState<Order | null>(null)
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState(true)
  const [paying, setPaying] = useState(false)
  const [cancelling, setCancelling] = useState(false)
  const [error, setError] = useState('')
  const [remainingMs, setRemainingMs] = useState(0)

  const load = async () => {
    if (!id) return
    setLoading(true)
    setError('')
    try {
      const [o, p] = await Promise.all([
        api.getOrder(id),
        api.listPayments(1, 10, id),
      ])
      setOrder(o)
      setPayments(p.payments)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (authLoading || !user) return
    load()
  }, [user, authLoading, id])

  useEffect(() => {
    if (!order || order.status !== 'ORDER_STATUS_PENDING') return

    const deadline = new Date(order.createdAt).getTime() + PAYMENT_RESERVE_MS
    const tick = () => setRemainingMs(deadline - Date.now())
    tick()
    const timer = window.setInterval(tick, 1000)
    return () => window.clearInterval(timer)
  }, [order])

  const pay = async () => {
    if (!id) return
    setPaying(true)
    setError('')
    try {
      const updated = await api.payOrder(id)
      setOrder(updated)
      const p = await api.listPayments(1, 10, id)
      setPayments(p.payments)
      await refresh()
      toast('Оплата прошла успешно')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка оплаты'
      setError(msg)
      toast(msg, 'error')
    } finally {
      setPaying(false)
    }
  }

  const cancel = async () => {
    if (!id) return
    if (!window.confirm('Отменить заказ и снять резерв?')) return
    setCancelling(true)
    setError('')
    try {
      const updated = await api.cancelOrder(id)
      setOrder(updated)
      toast('Заказ отменён', 'info')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка отмены'
      setError(msg)
      toast(msg, 'error')
    } finally {
      setCancelling(false)
    }
  }

  if (authLoading || loading) {
    return (
      <div className="page">
        <Link to="/orders" className="back-link">← Заказы</Link>
        <OrderListSkeleton count={1} />
      </div>
    )
  }
  if (error && !order) return <div className="alert alert-error">{error}</div>
  if (!order) return <div className="empty">Заказ не найден</div>

  const isPending = order.status === 'ORDER_STATUS_PENDING'
  const expired = isPending && remainingMs <= 0

  return (
    <div className="page">
      <Link to="/orders" className="back-link">← Заказы</Link>
      <div className="order-detail">
        <div className="order-detail-header">
          <h1>Заказ #{order.orderId}</h1>
          <StatusBadge status={order.status} />
        </div>
        <p className="order-date">
          {new Date(order.createdAt).toLocaleString('ru-RU')}
        </p>

        {isPending && (
          <div className={`payment-timer ${expired ? 'payment-timer-expired' : ''}`}>
            {expired ? (
              <span>Время резерва истекло — отмените заказ и оформите заново</span>
            ) : (
              <span>
                Оплатите в течение {PAYMENT_RESERVE_MINUTES} мин — осталось {formatCountdown(remainingMs)}
              </span>
            )}
          </div>
        )}

        {error && <div className="alert alert-error">{error}</div>}

        <div className="order-items">
          <h2>Позиции</h2>
          {(order.items ?? []).map((item) => (
            <div key={item.productId} className="order-item-row">
              <Link to={`/products/${item.productId}`}>{item.name}</Link>
              <span>{item.quantity} × {item.price.toFixed(2)} ₽</span>
              <span>{(item.quantity * item.price).toFixed(2)} ₽</span>
            </div>
          ))}
          <div className="order-item-row order-total-row">
            <span>Итого</span>
            <span />
            <strong>{order.totalPrice.toFixed(2)} ₽</strong>
          </div>
        </div>

        {isPending && (
          <div className="order-actions">
            <button
              type="button"
              className="btn btn-primary"
              disabled={paying || expired}
              onClick={pay}
            >
              {paying ? 'Оплата...' : 'Оплатить'}
            </button>
            <button
              type="button"
              className="btn btn-ghost"
              disabled={cancelling}
              onClick={cancel}
            >
              {cancelling ? 'Отмена...' : 'Отменить заказ'}
            </button>
          </div>
        )}

        {payments.length > 0 && (
          <div className="payments-section">
            <h2>Платежи</h2>
            {payments.map((p) => (
              <div key={p.paymentId} className="payment-row">
                <span>#{p.paymentId}</span>
                <StatusBadge status={p.status} />
                <span>{p.amount.toFixed(2)} ₽</span>
                <span className="order-date">
                  {new Date(p.createdAt).toLocaleString('ru-RU')}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
