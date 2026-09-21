// Root component — providers, router, and route definitions. / Корневой компонент — провайдеры, роутер и маршруты.
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { ErrorBoundary } from './components/ErrorBoundary'
import { Layout } from './components/Layout'
import { AuthProvider } from './context/AuthContext'
import { CartProvider } from './context/CartContext'
import { ToastProvider } from './context/ToastContext'
import { CartPage } from './pages/CartPage'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { OrderDetailPage } from './pages/OrderDetailPage'
import { OrdersPage } from './pages/OrdersPage'
import { ProductPage } from './pages/ProductPage'
import { RegisterPage } from './pages/RegisterPage'

// App wraps the shop UI in auth, cart, toast providers and React Router. / App оборачивает UI магазина в провайдеры auth, cart, toast и React Router.
export default function App() {
  return (
    // ToastProvider outermost — toasts visible on auth pages too. / ToastProvider снаружи — toast видны и на auth-страницах.
    <ToastProvider>
      <AuthProvider>
        {/* CartProvider depends on AuthProvider (needs user for cart fetch). / CartProvider зависит от AuthProvider (нужен user для загрузки корзины). */}
        <CartProvider>
          <BrowserRouter>
            {/* ErrorBoundary catches render crashes in any route. / ErrorBoundary перехватывает падения рендера на любом маршруте. */}
            <ErrorBoundary>
              <Routes>
                {/* Layout provides header/nav; child routes render in <Outlet>. / Layout даёт шапку/nav; дочерние маршруты рендерятся в <Outlet>. */}
                <Route element={<Layout />}>
                  <Route index element={<HomePage />} />
                  <Route path="products/:id" element={<ProductPage />} />
                  {/* Protected routes use useRequireAuth hook internally. / Защищённые маршруты используют хук useRequireAuth внутри. */}
                  <Route path="cart" element={<CartPage />} />
                  <Route path="orders" element={<OrdersPage />} />
                  <Route path="orders/:id" element={<OrderDetailPage />} />
                  <Route path="login" element={<LoginPage />} />
                  <Route path="register" element={<RegisterPage />} />
                </Route>
              </Routes>
            </ErrorBoundary>
          </BrowserRouter>
        </CartProvider>
      </AuthProvider>
    </ToastProvider>
  )
}
