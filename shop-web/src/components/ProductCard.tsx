// Product card for the catalog grid. / Карточка товара для сетки каталога.
import { Link } from 'react-router-dom'
import { useState } from 'react'
import type { Product } from '../api/types'
import { getProductImage, getShortDescription } from '../data/productImages'

interface Props {
  // Product data from catalog API. / Данные товара из catalog API.
  product: Product
}

// ProductCard shows image, name, short description, and price. / ProductCard показывает изображение, название, краткое описание и цену.
export function ProductCard({ product }: Props) {
  const image = getProductImage(product.id)
  // imgError: fallback to placeholder when Unsplash URL fails. / imgError: запасной placeholder при ошибке URL Unsplash.
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
