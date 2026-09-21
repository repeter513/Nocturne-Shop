// Catalog home page — product grid with search and category filters. / Главная страница каталога — сетка товаров с поиском и фильтрами категорий.
import { useEffect, useMemo, useState } from 'react'
import { api } from '../api/client'
import type { Category, Product } from '../api/types'
import { ProductCard } from '../components/ProductCard'
import { ProductGridSkeleton } from '../components/Skeleton'

// HomePage lists products with pagination, search, and category filter. / HomePage показывает товары с пагинацией, поиском и фильтром по категории.
export function HomePage() {
  const [products, setProducts] = useState<Product[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [categoryId, setCategoryId] = useState('')
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const pageSize = 12
  const isSearching = search.trim().length > 0

  // Load category filter buttons once on mount. / Загружаем кнопки фильтра категорий один раз при монтировании.
  useEffect(() => {
    api.listCategories()
      .then((r) => setCategories(r.categories ?? []))
      .catch(() => {})
  }, [])

  // Re-fetch products when page, category, or search changes. / Перезагружаем товары при смене page, category или search.
  useEffect(() => {
    setLoading(true)
    setError('')
    const query = search.trim()
    // Client-side search: fetch up to 100 items then filter locally. / Клиентский поиск: загружаем до 100 товаров и фильтруем локально.
    const fetchPage = query ? 1 : page
    const fetchSize = query ? 100 : pageSize

    api.listProducts(fetchPage, fetchSize, categoryId || undefined)
      .then((r) => {
        setProducts(r.products)
        setTotal(r.totalCount)
      })
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [page, categoryId, search])

  // Client-side name/description filter when search is active. / Клиентская фильтрация по названию/описанию при активном поиске.
  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase()
    if (!q) return products
    return products.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q),
    )
  }, [products, search])

  const totalPages = isSearching ? 1 : Math.ceil(total / pageSize)

  return (
    <div className="page">
      <div className="page-header">
        <h1>Каталог</h1>
        <input
          type="search"
          className="search-input"
          placeholder="Поиск по названию или описанию..."
          value={search}
          onChange={(e) => {
            setSearch(e.target.value)
            setPage(1)
          }}
          aria-label="Поиск товаров"
        />
        {categories.length > 0 && (
          <div className="filters">
            <button
              type="button"
              className={`filter-btn ${!categoryId ? 'active' : ''}`}
              onClick={() => { setCategoryId(''); setPage(1) }}
            >
              Все
            </button>
            {categories.map((c) => (
              <button
                key={c.id}
                type="button"
                className={`filter-btn ${categoryId === c.id ? 'active' : ''}`}
                onClick={() => { setCategoryId(c.id); setPage(1) }}
              >
                {c.name}
              </button>
            ))}
          </div>
        )}
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      {loading ? (
        <ProductGridSkeleton />
      ) : filtered.length === 0 ? (
        <div className="empty">
          {isSearching ? 'Ничего не найдено' : 'Товары не найдены'}
        </div>
      ) : (
        <>
          <div className="product-grid">
            {filtered.map((p) => <ProductCard key={p.id} product={p} />)}
          </div>
          {!isSearching && totalPages > 1 && (
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
