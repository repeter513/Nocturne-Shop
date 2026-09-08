import type {
  ApiError,
  AuthTokens,
  Cart,
  Category,
  Order,
  PaginatedOrders,
  PaginatedPayments,
  PaginatedProducts,
  Payment,
  Product,
  Stock,
  User,
} from './types'

const BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1'

class ApiClient {
  private accessToken: string | null = null

  setToken(token: string | null) {
    this.accessToken = token
  }

  getToken() {
    return this.accessToken
  }

  private async request<T>(
    path: string,
    options: RequestInit = {},
    auth = false,
  ): Promise<T> {
    const headers = new Headers(options.headers)
    headers.set('Content-Type', 'application/json')
    if (auth && this.accessToken) {
      headers.set('Authorization', `Bearer ${this.accessToken}`)
    }

    const res = await fetch(`${BASE}${path}`, { ...options, headers })

    if (res.status === 204) return undefined as T

    const body = await res.json().catch(() => null)

    if (!res.ok) {
      const err = body as ApiError | null
      throw new Error(err?.error ?? `HTTP ${res.status}`)
    }

    return body as T
  }

  register(email: string, password: string) {
    return this.request<{ userId: number }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  }

  login(email: string, password: string) {
    return this.request<AuthTokens>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  }

  refresh(refreshToken: string) {
    return this.request<Omit<AuthTokens, 'userId'>>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    })
  }

  me() {
    return this.request<User>('/auth/me', {}, true)
  }

  listProducts(page = 1, pageSize = 12, categoryId?: string) {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    })
    if (categoryId) params.set('categoryId', categoryId)
    return this.request<PaginatedProducts>(`/products?${params}`)
      .then((r) => ({ products: r.products ?? [], totalCount: r.totalCount ?? 0 }))
  }

  getProduct(id: string) {
    return this.request<Product>(`/products/${id}`)
  }

  getStock(id: string) {
    return this.request<Stock>(`/products/${id}/stock`)
  }

  listCategories() {
    return this.request<{ categories: Category[] }>('/categories')
      .then((r) => ({ categories: r.categories ?? [] }))
  }

  getCart() {
    return this.request<Cart>('/cart', {}, true)
  }

  addToCart(productId: number, quantity: number) {
    return this.request<Cart>('/cart/items', {
      method: 'POST',
      body: JSON.stringify({ productId, quantity }),
    }, true)
  }

  updateCartItem(productId: string, quantity: number) {
    return this.request<Cart>(`/cart/items/${productId}`, {
      method: 'PUT',
      body: JSON.stringify({ quantity }),
    }, true)
  }

  removeFromCart(productId: string) {
    return this.request<Cart>(`/cart/items/${productId}`, {
      method: 'DELETE',
    }, true)
  }

  clearCart() {
    return this.request<void>('/cart', { method: 'DELETE' }, true)
  }

  createOrder() {
    return this.request<Order>('/orders', { method: 'POST' }, true)
  }

  payOrder(id: string) {
    return this.request<Order>(`/orders/${id}/pay`, { method: 'POST' }, true)
  }

  cancelOrder(id: string) {
    return this.request<Order>(`/orders/${id}/cancel`, { method: 'POST' }, true)
  }

  listOrders(page = 1, pageSize = 10) {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    })
    return this.request<PaginatedOrders>(`/orders?${params}`, {}, true)
      .then((r) => ({ orders: r.orders ?? [], totalCount: r.totalCount ?? 0 }))
  }

  getOrder(id: string) {
    return this.request<Order>(`/orders/${id}`, {}, true)
  }

  listPayments(page = 1, pageSize = 10, orderId?: string) {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    })
    if (orderId) params.set('orderId', orderId)
    return this.request<PaginatedPayments>(`/payments?${params}`, {}, true)
      .then((r) => ({ payments: r.payments ?? [], totalCount: r.totalCount ?? 0 }))
  }

  getPayment(id: string) {
    return this.request<Payment>(`/payments/${id}`, {}, true)
  }
}

export const api = new ApiClient()
