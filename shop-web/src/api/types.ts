export interface User {
  id: number
  email: string
}

export interface AuthTokens {
  accessToken: string
  refreshToken: string
  expiresIn: number
  userId: number
}

export interface Product {
  id: string
  name: string
  description: string
  price: number
  categoryId: string
  active: boolean
}

export interface Category {
  id: string
  name: string
  parentId: string
}

export interface Stock {
  productId: string
  quantity: number
}

export interface CartItem {
  productId: string
  quantity: number
  name: string
  price: number
}

export interface Cart {
  userId: string
  items: CartItem[]
  totalPrice: number
}

export type OrderStatus =
  | 'ORDER_STATUS_UNSPECIFIED'
  | 'ORDER_STATUS_PENDING'
  | 'ORDER_STATUS_PAID'
  | 'ORDER_STATUS_FAILED'
  | 'ORDER_STATUS_CANCELLED'

export interface OrderItem {
  productId: string
  quantity: number
  name: string
  price: number
}

export interface Order {
  orderId: string
  userId: string
  items: OrderItem[]
  totalPrice: number
  status: OrderStatus
  createdAt: string
  paymentId: string
}

export type PaymentStatus =
  | 'PAYMENT_STATUS_UNSPECIFIED'
  | 'PAYMENT_STATUS_PENDING'
  | 'PAYMENT_STATUS_SUCCESS'
  | 'PAYMENT_STATUS_FAILED'

export interface Payment {
  paymentId: string
  orderId: string
  userId: string
  amount: number
  status: PaymentStatus
  createdAt: string
}

export interface ApiError {
  error: string
}

export interface PaginatedProducts {
  products: Product[]
  totalCount: number
}

export interface PaginatedOrders {
  orders: Order[]
  totalCount: number
}

export interface PaginatedPayments {
  payments: Payment[]
  totalCount: number
}
