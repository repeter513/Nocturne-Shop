// App shell — header, navigation, and page outlet. / Оболочка приложения — шапка, навигация и outlet страниц.
import { Link, NavLink, Outlet } from 'react-router-dom'
import { SHOP_NAME } from '../config'
import { useAuth } from '../context/AuthContext'
import { useCart } from '../context/CartContext'
import { useToast } from '../context/ToastContext'
import './Layout.css'

// Layout renders the persistent header and nested route content. / Layout отображает постоянную шапку и вложенный контент маршрутов.
export function Layout() {
  const { user, logout, loading } = useAuth()
  const { count } = useCart()
  const { toast } = useToast()

  const handleLogout = () => {
    logout()
    toast('Вы вышли из аккаунта', 'info')
  }

  return (
    <div className="layout">
      <a href="#main-content" className="skip-link">К содержимому</a>
      <header className="header">
        <Link to="/" className="logo">{SHOP_NAME}</Link>
        <nav className="nav">
          <NavLink to="/" end>Каталог</NavLink>
          {/* Cart and Orders links only visible when authenticated. / Ссылки Корзина и Заказы видны только авторизованным. */}
          {user && (
            <>
              <NavLink to="/cart" className="nav-cart">
                Корзина
                {count > 0 && <span className="cart-badge">{count > 99 ? '99+' : count}</span>}
              </NavLink>
              <NavLink to="/orders">Заказы</NavLink>
            </>
          )}
        </nav>
        <div className="header-actions">
          {loading ? (
            <span className="header-loading" aria-hidden="true" />
          ) : user ? (
            <>
              <span className="user-email">{user.email}</span>
              <button type="button" className="btn btn-ghost" onClick={handleLogout}>
                Выйти
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="btn btn-ghost">Вход</Link>
              <Link to="/register" className="btn btn-primary">Регистрация</Link>
            </>
          )}
        </div>
      </header>
      {/* Outlet renders the matched child route (HomePage, CartPage, etc.). / Outlet рендерит совпавший дочерний маршрут (HomePage, CartPage и т.д.). */}
      <main id="main-content" className="main">
        <Outlet />
      </main>
    </div>
  )
}
