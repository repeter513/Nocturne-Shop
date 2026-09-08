import { Link } from 'react-router-dom'
import { useState } from 'react'
import type { Product } from '../api/types'
import { getProductImage, getShortDescription } from '../data/productImages'

interface Props {
  product: Product
}

export function ProductCard({ product }: Props) {
  const image = getProductImage(product.id)
  const [imgError, setImgError] = useState(false)

  return (
    <Link to={`/products/${product.id}`} className="product-card">
      <div className="product-card-image">
        {image && !imgError ? (
          <img
            src={image}
            alt={product.name}
            loading="lazy"
            onError={() => setImgError(true)}
          />
        ) : (
          <div className="product-image-placeholder" />
        )}
      </div>
      <div className="product-card-body">
        <h3 className="product-name">{product.name}</h3>
        <p className="product-desc">{getShortDescription(product.description)}</p>
        <div className="product-footer">
          <span className="product-price">{product.price.toFixed(2)} ₽</span>
          {!product.active && <span className="badge badge-muted">Недоступен</span>}
        </div>
      </div>
    </Link>
  )
}
