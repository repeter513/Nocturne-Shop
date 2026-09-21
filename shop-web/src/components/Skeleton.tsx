// Loading skeleton placeholders for async content. / Скелетоны-заглушки для асинхронного контента.

interface Props {
  // Additional CSS class names. / Дополнительные CSS-классы.
  className?: string
}

// Generic animated skeleton block. / Универсальный анимированный блок-скелетон.
export function Skeleton({ className = '' }: Props) {
  return <div className={`skeleton ${className}`.trim()} aria-hidden="true" />
}

// Skeleton grid mimicking the product catalog layout. / Скелетон-сетка, имитирующая каталог товаров.
export function ProductGridSkeleton({ count = 6 }: { count?: number }) {
  return (
    <div className="product-grid">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className="skeleton-card">
          <Skeleton className="skeleton-image" />
          <Skeleton className="skeleton-line skeleton-line-lg" />
          <Skeleton className="skeleton-line" />
          <Skeleton className="skeleton-line skeleton-line-sm" />
        </div>
      ))}
    </div>
  )
}

// Skeleton for the product detail page. / Скелетон страницы деталей товара.
export function ProductDetailSkeleton() {
  return (
    <div className="product-detail">
      <Skeleton className="skeleton-detail-image" />
      <div>
        <Skeleton className="skeleton-line skeleton-line-xl" />
        <Skeleton className="skeleton-line skeleton-line-md" />
        <Skeleton className="skeleton-line" />
        <Skeleton className="skeleton-line" />
        <Skeleton className="skeleton-line skeleton-line-sm" />
      </div>
    </div>
  )
}

// Skeleton for the cart page item list. / Скелетон списка позиций на странице корзины.
export function CartSkeleton() {
  return (
    <div className="cart-list">
      {Array.from({ length: 3 }, (_, i) => (
        <div key={i} className="skeleton-cart-item">
          <Skeleton className="skeleton-line skeleton-line-lg" />
          <Skeleton className="skeleton-line skeleton-line-sm" />
        </div>
      ))}
    </div>
  )
}

// Skeleton for the orders list page. / Скелетон страницы списка заказов.
export function OrderListSkeleton({ count = 4 }: { count?: number }) {
  return (
    <div className="order-list">
      {Array.from({ length: count }, (_, i) => (
        <div key={i} className="skeleton-order-card">
          <Skeleton className="skeleton-line skeleton-line-md" />
          <Skeleton className="skeleton-line skeleton-line-sm" />
        </div>
      ))}
    </div>
  )
}
