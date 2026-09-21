// HTTP client for the shop REST / gRPC-gateway API. / HTTP-клиент для REST / gRPC-gateway API магазина.
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

// Base URL from env or same-origin proxy (Vite dev proxy → BFF :8090). / Базовый URL из env или прокси того же origin (Vite dev proxy → BFF :8090).
const BASE = import.meta.env.VITE_API_BASE_URL || '/api/v1'

// protojson serializes int64 as strings; UI expects numbers for .toFixed etc. / protojson сериализует int64 как строки; UI ожидает числа для .toFixed и т.п.
const NUMERIC_KEYS = new Set(['price', 'totalPrice', 'amount'])

// Recursively converts protojson string numbers to JS numbers. / Рекурсивно преобразует строковые числа protojson в числа JS.
function normalizeProtoJson<T>(value: T): T {
  if (Array.isArray(value)) {
    return value.map((item) => normalizeProtoJson(item)) as T
  }
  if (value !== null && typeof value === 'object') {
    const out: Record<string, unknown> = {}
    for (const [key, val] of Object.entries(value)) {
      out[key] =
        typeof val === 'string' && NUMERIC_KEYS.has(key) ? Number(val) : normalizeProtoJson(val)
    }
    return out as T
  }
  return value
}

// ApiClient wraps fetch with auth headers and protojson normalization. / ApiClient оборачивает fetch с auth-заголовками и нормализацией protojson.
class ApiClient {
  // In-memory access token; synced with localStorage by AuthContext. / Access token в памяти; синхронизируется с localStorage через AuthContext.
  private accessToken: string | null = null

  // Sets the bearer access token for authenticated requests. / Устанавливает bearer access token для авторизованных запросов.
  setToken(token: string | null) {
    this.accessToken = token
  }

  // Returns the current access token. / Возвращает текущий access token.
  getToken() {
    return this.accessToken
  }

  // Low-level JSON request with optional auth. / Низкоуровневый JSON-запрос с опциональной авторизацией.
  private async request<T>(
    path: string,
    options: RequestInit = {},
    auth = false,
  ): Promise<T> {
    const headers = new Headers(options.headers)
    headers.set('Content-Type', 'application/json')
    // Attach JWT when auth=true — BFF requireAuth middleware checks this header. / Прикрепляем JWT при auth=true — middleware requireAuth BFF проверяет этот заголовок.
    if (auth && this.accessToken) {
      headers.set('Authorization', `Bearer ${this.accessToken}`)
    }

    const res = await fetch(`${BASE}${path}`, { ...options, headers })

    // 204 No Content — e.g. DELETE /cart. / 204 No Content — например DELETE /cart.
    if (res.status === 204) return undefined as T

    const body = await res.json().catch(() => null)

    if (!res.ok) {
      const err = body as ApiError | null
      throw new Error(err?.error ?? `HTTP ${res.status}`)
    }

    return normalizeProtoJson(body) as T
  }

  // POST /auth/register → BFF AuthService.RegisterUser / POST /auth/register → BFF AuthService.RegisterUser
  register(email: string, password: string) {
    return this.request<{ userId: number }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  }

  // POST /auth/login → BFF AuthService.LoginUser / POST /auth/login → BFF AuthService.LoginUser
  login(email: string, password: string) {
    return this.request<AuthTokens>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    })
  }

  // POST /auth/refresh → BFF AuthService.RefreshToken / POST /auth/refresh → BFF AuthService.RefreshToken
  refresh(refreshToken: string) {
    return this.request<Omit<AuthTokens, 'userId'>>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refreshToken }),
    })
  }

  // GET /auth/me → BFF ValidateToken + GetUserInfo (auth required) / GET /auth/me → BFF ValidateToken + GetUserInfo (требуется auth)
  me() {
    return this.request<User>('/auth/me', {}, true)
  }

  // GET /products → BFF CatalogService.ListProducts / GET /products → BFF CatalogService.ListProducts
  listProducts(page = 1, pageSize = 12, categoryId?: string) {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    })
    if (categoryId) params.set('categoryId', categoryId)
    return this.request<PaginatedProducts>(`/products?${params}`)
      .then((r) => ({ products: r.products ?? [], totalCount: r.totalCount ?? 0 }))
  }

  // GET /products/:id → BFF CatalogService.GetProduct / GET /products/:id → BFF CatalogService.GetProduct
  getProduct(id: string) {
    return this.request<Product>(`/products/${id}`)
  }

  // GET /products/:id/stock → BFF CatalogService.GetStock / GET /products/:id/stock → BFF CatalogService.GetStock
  getStock(id: string) {
    return this.request<Stock>(`/products/${id}/stock`)
  }

  // GET /categories → BFF CatalogService.ListCategories / GET /categories → BFF CatalogService.ListCategories
  listCategories() {
    return this.request<{ categories: Category[] }>('/categories')
      .then((r) => ({ categories: r.categories ?? [] }))
  }

  // GET /cart → BFF CartService.GetCart (auth required) / GET /cart → BFF CartService.GetCart (требуется auth)
  getCart() {
    return this.request<Cart>('/cart', {}, true)
  }

  // POST /cart/items → BFF CartService.AddToCart (auth required) / POST /cart/items → BFF CartService.AddToCart (требуется auth)
  addToCart(productId: number, quantity: number) {
    return this.request<Cart>('/cart/items', {
      method: 'POST',
      body: JSON.stringify({ productId, quantity }),
    }, true)
  }

  // PUT /cart/items/:id → BFF CartService.UpdateCartItem (auth required) / PUT /cart/items/:id → BFF CartService.UpdateCartItem (требуется auth)
  updateCartItem(productId: string, quantity: number) {
    return this.request<Cart>(`/cart/items/${productId}`, {
      method: 'PUT',
      body: JSON.stringify({ quantity }),
    }, true)
  }

  // DELETE /cart/items/:id → BFF CartService.RemoveFromCart (auth required) / DELETE /cart/items/:id → BFF CartService.RemoveFromCart (требуется auth)
  removeFromCart(productId: string) {
    return this.request<Cart>(`/cart/items/${productId}`, {
      method: 'DELETE',
    }, true)
  }

  // DELETE /cart → BFF CartService.ClearCart (auth required) / DELETE /cart → BFF CartService.ClearCart (требуется auth)
  clearCart() {
    return this.request<void>('/cart', { method: 'DELETE' }, true)
  }

  // POST /orders → BFF OrderService.CreateOrder (auth required) / POST /orders → BFF OrderService.CreateOrder (требуется auth)
  createOrder() {
    return this.request<Order>('/orders', { method: 'POST' }, true)
  }

  // POST /orders/:id/pay → BFF OrderService.PayOrder (auth required) / POST /orders/:id/pay → BFF OrderService.PayOrder (требуется auth)
  payOrder(id: string) {
    return this.request<Order>(`/orders/${id}/pay`, { method: 'POST' }, true)
  }

  // POST /orders/:id/cancel → BFF OrderService.CancelOrder (auth required) / POST /orders/:id/cancel → BFF OrderService.CancelOrder (требуется auth)
  cancelOrder(id: string) {
    return this.request<Order>(`/orders/${id}/cancel`, { method: 'POST' }, true)
  }

  // GET /orders → BFF OrderService.ListOrders (auth required) / GET /orders → BFF OrderService.ListOrders (требуется auth)
  listOrders(page = 1, pageSize = 10) {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    })
    return this.request<PaginatedOrders>(`/orders?${params}`, {}, true)
      .then((r) => ({ orders: r.orders ?? [], totalCount: r.totalCount ?? 0 }))
  }

  // GET /orders/:id → BFF OrderService.GetOrder (auth required) / GET /orders/:id → BFF OrderService.GetOrder (требуется auth)
  getOrder(id: string) {
    return this.request<Order>(`/orders/${id}`, {}, true)
  }

  // GET /payments → BFF PaymentService.ListPayments (auth required) / GET /payments → BFF PaymentService.ListPayments (требуется auth)
  listPayments(page = 1, pageSize = 10, orderId?: string) {
    const params = new URLSearchParams({
      page: String(page),
      pageSize: String(pageSize),
    })
    if (orderId) params.set('orderId', orderId)
    return this.request<PaginatedPayments>(`/payments?${params}`, {}, true)
      .then((r) => ({ payments: r.payments ?? [], totalCount: r.totalCount ?? 0 }))
  }

  // GET /payments/:id → BFF PaymentService.GetPayment (auth required) / GET /payments/:id → BFF PaymentService.GetPayment (требуется auth)
  getPayment(id: string) {
    return this.request<Payment>(`/payments/${id}`, {}, true)
  }
}

// Singleton API client instance used across the app. / Единственный экземпляр API-клиента, используемый в приложении.
export const api = new ApiClient()
