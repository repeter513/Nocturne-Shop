import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { Order } from '../api/types'
import { StatusBadge } from '../components/StatusBadge'
import { OrderListSkeleton } from '../components/Skeleton'
import { useRequireAuth } from '../hooks/useRequireAuth'

const PAGE_SIZE = 10

export function OrdersPage() {
  const { user, authLoading } = useRequireAuth()
  const [orders, setOrders] = useState<Order[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (authLoading || !user) return
    setLoading(true)
    setError('')
    api.listOrders(page, PAGE_SIZE)
      .then((r) => {
        setOrders(r.orders)
        setTotal(r.totalCount)
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [user, authLoading, page])

  if (authLoading || (loading && orders.length === 0)) {
    return (
      <div className="page">
        <h1>Заказы</h1>
        <OrderListSkeleton />
      </div>
    )
  }

  const totalPages = Math.ceil(total / PAGE_SIZE)

  return (
    <div className="page">
      <h1>Заказы</h1>
      {error && <div className="alert alert-error">{error}</div>}

      {orders.length === 0 ? (
        <div className="empty">
          Заказов пока нет. <Link to="/">Перейти в каталог</Link>
        </div>
      ) : (
        <>
          <div className="order-list">
            {orders.map((order) => (
              <Link key={order.orderId} to={`/orders/${order.orderId}`} className="order-card">
                <div className="order-card-header">
                  <span className="order-id">#{order.orderId}</span>
                  <StatusBadge status={order.status} />
                </div>
                <div className="order-card-body">
                  <span>{(order.items ?? []).length} поз.</span>
                  <span className="order-total">{order.totalPrice.toFixed(2)} ₽</span>
                </div>
                <span className="order-date">
                  {new Date(order.createdAt).toLocaleString('ru-RU')}
                </span>
              </Link>
            ))}
          </div>
          {totalPages > 1 && (
            <div className="pagination">
              <button
                type="button"
                className="btn btn-ghost"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                Назад
              </button>
              <span className="page-info">{page} / {totalPages}</span>
              <button
                type="button"
                className="btn btn-ghost"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Далее
              </button>
            </div>
          )}
        </>
      )}
    </div>
  )
}
