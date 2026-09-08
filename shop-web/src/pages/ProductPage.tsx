import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../api/client'
import type { Product, Stock } from '../api/types'
import { useAuth } from '../context/AuthContext'
import { ProductDetailSkeleton } from '../components/Skeleton'
import { useCart } from '../context/CartContext'
import { useToast } from '../context/ToastContext'
import { getDetailParagraphs, getProductImage } from '../data/productImages'

export function ProductPage() {
  const { id } = useParams<{ id: string }>()
  const { user } = useAuth()
  const { toast } = useToast()
  const { refresh } = useCart()
  const navigate = useNavigate()
  const [product, setProduct] = useState<Product | null>(null)
  const [stock, setStock] = useState<Stock | null>(null)
  const [quantity, setQuantity] = useState(1)
  const [loading, setLoading] = useState(true)
  const [adding, setAdding] = useState(false)
  const [error, setError] = useState('')
  const [added, setAdded] = useState(false)
  const [imgError, setImgError] = useState(false)

  useEffect(() => {
    if (!id) return
    setLoading(true)
    setImgError(false)
    setAdded(false)
    Promise.all([api.getProduct(id), api.getStock(id)])
      .then(([p, s]) => {
        setProduct(p)
        setStock(s)
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [id])

  const handleAdd = async () => {
    if (!user) {
      navigate('/login')
      return
    }
    if (!product) return
    setAdding(true)
    setError('')
    setAdded(false)
    try {
      await api.addToCart(Number(product.id), quantity)
      setAdded(true)
      await refresh()
      toast('Товар добавлен в корзину')
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Ошибка'
      setError(msg)
      toast(msg, 'error')
    } finally {
      setAdding(false)
    }
  }

  if (loading) {
    return (
      <div className="page">
        <Link to="/" className="back-link">← Каталог</Link>
        <ProductDetailSkeleton />
      </div>
    )
  }
  if (error && !product) return <div className="alert alert-error">{error}</div>
  if (!product) return <div className="empty">Товар не найден</div>

  const image = getProductImage(product.id)
  const paragraphs = getDetailParagraphs(product.description)

  return (
    <div className="page">
      <Link to="/" className="back-link">← Каталог</Link>
      <div className="product-detail">
        <div className="product-detail-image">
          {image && !imgError ? (
            <img
              src={image}
              alt={product.name}
              onError={() => setImgError(true)}
            />
          ) : (
            <div className="product-image-placeholder" />
          )}
        </div>
        <div className="product-detail-info">
          <h1>{product.name}</h1>
          <div className="product-detail-meta">
            <span className="product-price-lg">{product.price.toFixed(2)} ₽</span>
            {stock && (
              <span className="stock-info">В наличии: {stock.quantity} шт.</span>
            )}
          </div>

          <div className="product-detail-desc">
            {paragraphs.map((p, i) => (
              <p key={i}>{p}</p>
            ))}
          </div>

          {error && <div className="alert alert-error">{error}</div>}
          {added && (
            <div className="add-success">
              <span className="add-success-text">Добавлено в корзину</span>
              <Link to="/cart" className="btn btn-cart-link">Перейти в корзину</Link>
            </div>
          )}

          {product.active && stock && stock.quantity > 0 && (
            <div className="add-to-cart">
              <div className="quantity-control">
                <button
                  type="button"
                  className="btn btn-ghost"
                  disabled={quantity <= 1}
                  onClick={() => setQuantity((q) => q - 1)}
                >
                  −
                </button>
                <span className="quantity-value">{quantity}</span>
                <button
                  type="button"
                  className="btn btn-ghost"
                  disabled={quantity >= stock.quantity}
                  onClick={() => setQuantity((q) => q + 1)}
                >
                  +
                </button>
              </div>
              <button
                type="button"
                className="btn btn-primary"
                disabled={adding}
                onClick={handleAdd}
              >
                {adding ? 'Добавление...' : 'В корзину'}
              </button>
            </div>
          )}

          {!product.active && (
            <div className="alert alert-muted">Товар недоступен</div>
          )}
        </div>
      </div>
    </div>
  )
}
