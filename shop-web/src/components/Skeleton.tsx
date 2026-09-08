interface Props {
  className?: string
}

export function Skeleton({ className = '' }: Props) {
  return <div className={`skeleton ${className}`.trim()} aria-hidden="true" />
}

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
