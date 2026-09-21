// TypeScript types mirroring backend API / proto models. / TypeScript-типы, соответствующие моделям API / proto бэкенда.

// Authenticated user profile. / Профиль авторизованного пользователя.
export interface User {
  // Numeric user ID from auth service. / Числовой ID пользователя из auth-сервиса.
  id: number
  // Login email address. / Email для входа.
  email: string
}

// JWT token pair returned on login / refresh. / Пара JWT-токенов, возвращаемая при входе / обновлении.
export interface AuthTokens {
  // Short-lived JWT sent as Authorization: Bearer header. / Краткоживущий JWT в заголовке Authorization: Bearer.
  accessToken: string
  // Long-lived token stored in localStorage for silent refresh. / Долгоживущий токен в localStorage для тихого обновления.
  refreshToken: string
  // Access token TTL in seconds. / TTL access-токена в секундах.
  expiresIn: number
  // Owner user ID. / ID владельца.
  userId: number
}

// Catalog product. / Товар каталога.
export interface Product {
  // Product ID (proto int64 serialized as string). / ID товара (proto int64 сериализуется как строка).
  id: string
  // Display name. / Отображаемое название.
  name: string
  // Full description (may contain \n\n paragraphs). / Полное описание (может содержать абзацы \n\n).
  description: string
  // Unit price in rubles (normalized from protojson string). / Цена за единицу в рублях (нормализована из protojson-строки).
  price: number
  // Category foreign key. / Внешний ключ категории.
  categoryId: string
  // Whether the product can be purchased. / Можно ли купить товар.
  active: boolean
}

// Product category. / Категория товаров.
export interface Category {
  // Category ID. / ID категории.
  id: string
  // Display name shown in filter buttons. / Название для кнопок фильтра.
  name: string
  // Parent category ID (empty for root). / ID родительской категории (пусто для корня).
  parentId: string
}

// Available stock for a product. / Доступный остаток товара.
export interface Stock {
  // Product this stock row belongs to. / Товар, к которому относится остаток.
  productId: string
  // Units available for purchase. / Единиц доступно для покупки.
  quantity: number
}

// Single line item in the shopping cart. / Одна позиция в корзине.
export interface CartItem {
  // Product reference. / Ссылка на товар.
  productId: string
  // Units in cart. / Количество в корзине.
  quantity: number
  // Denormalized product name for display. / Денормализованное название для отображения.
  name: string
  // Unit price at time of add. / Цена за единицу на момент добавления.
  price: number
}

// User shopping cart with items and total. / Корзина пользователя с позициями и итогом.
export interface Cart {
  // Owner user ID. / ID владельца корзины.
  userId: string
  // Line items. / Позиции.
  items: CartItem[]
  // Sum of line totals. / Сумма по позициям.
  totalPrice: number
}

// Order lifecycle status enum values from proto. / Значения статуса заказа из proto.
export type OrderStatus =
  | 'ORDER_STATUS_UNSPECIFIED'
  | 'ORDER_STATUS_PENDING'
  | 'ORDER_STATUS_PAID'
  | 'ORDER_STATUS_FAILED'
  | 'ORDER_STATUS_CANCELLED'

// Line item within an order. / Позиция внутри заказа.
export interface OrderItem {
  // Product reference. / Ссылка на товар.
  productId: string
  // Ordered quantity. / Заказанное количество.
  quantity: number
  // Snapshot of product name. / Снимок названия товара.
  name: string
  // Snapshot of unit price. / Снимок цены за единицу.
  price: number
}

// Placed order with items, status, and payment link. / Оформленный заказ с позициями, статусом и ссылкой на оплату.
export interface Order {
  // Unique order identifier. / Уникальный идентификатор заказа.
  orderId: string
  // Buyer user ID. / ID покупателя.
  userId: string
  // Frozen line items at checkout time. / Зафиксированные позиции на момент оформления.
  items: OrderItem[]
  // Order total in rubles. / Итог заказа в рублях.
  totalPrice: number
  // Current lifecycle status. / Текущий статус жизненного цикла.
  status: OrderStatus
  // ISO timestamp when order was created (starts payment timer). / ISO-время создания (запускает таймер оплаты).
  createdAt: string
  // Linked payment record ID. / ID связанной записи платежа.
  paymentId: string
}

// Payment lifecycle status enum values from proto. / Значения статуса платежа из proto.
export type PaymentStatus =
  | 'PAYMENT_STATUS_UNSPECIFIED'
  | 'PAYMENT_STATUS_PENDING'
  | 'PAYMENT_STATUS_SUCCESS'
  | 'PAYMENT_STATUS_FAILED'

// Payment record linked to an order. / Запись платежа, связанная с заказом.
export interface Payment {
  // Payment identifier. / Идентификатор платежа.
  paymentId: string
  // Parent order. / Родительский заказ.
  orderId: string
  // Payer user ID. / ID плательщика.
  userId: string
  // Charged amount in rubles. / Списанная сумма в рублях.
  amount: number
  // Payment processing status. / Статус обработки платежа.
  status: PaymentStatus
  // ISO timestamp of payment creation. / ISO-время создания платежа.
  createdAt: string
}

// Standard API error response shape. / Стандартный формат ответа об ошибке API.
export interface ApiError {
  // Human-readable error message from BFF/gRPC. / Читаемое сообщение об ошибке от BFF/gRPC.
  error: string
}

// Paginated product list response. / Ответ со списком товаров и пагинацией.
export interface PaginatedProducts {
  // Current page of products. / Текущая страница товаров.
  products: Product[]
  // Total matching products across all pages. / Всего подходящих товаров на всех страницах.
  totalCount: number
}

// Paginated order list response. / Ответ со списком заказов и пагинацией.
export interface PaginatedOrders {
  // Current page of orders. / Текущая страница заказов.
  orders: Order[]
  // Total orders for this user. / Всего заказов пользователя.
  totalCount: number
}

// Paginated payment list response. / Ответ со списком платежей и пагинацией.
export interface PaginatedPayments {
  // Current page of payments. / Текущая страница платежей.
  payments: Payment[]
  // Total matching payments. / Всего подходящих платежей.
  totalCount: number
}
