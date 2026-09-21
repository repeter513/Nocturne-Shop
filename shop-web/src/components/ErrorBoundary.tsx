// React error boundary — catches render errors and shows fallback UI. / React error boundary — перехватывает ошибки рендера и показывает запасной UI.
import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  // Subtree to protect (typically all Routes). / Поддерево для защиты (обычно все Routes).
  children: ReactNode
}

interface State {
  // Captured error or null when subtree is healthy. / Перехваченная ошибка или null когда поддерево в порядке.
  error: Error | null
}

// ErrorBoundary catches unhandled React errors in the subtree. / ErrorBoundary перехватывает необработанные ошибки React в поддереве.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  // React lifecycle: derive state from thrown error (no side effects). / React lifecycle: выводим state из ошибки (без побочных эффектов).
  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  // React lifecycle: log stack for debugging (runs after render). / React lifecycle: логируем stack для отладки (после рендера).
  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('UI error:', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <div className="page error-boundary">
          <h1>Что-то пошло не так</h1>
          <p className="error-boundary-text">
            Страница не загрузилась. Попробуйте обновить или вернуться в каталог.
          </p>
          <div className="error-boundary-actions">
            {/* Hard navigation resets entire React tree. / Жёсткая навигация сбрасывает всё React-дерево. */}
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => window.location.assign('/')}
            >
              В каталог
            </button>
            {/* Clears error state and retries rendering children. / Сбрасывает error state и повторяет рендер children. */}
            <button
              type="button"
              className="btn btn-ghost"
              onClick={() => this.setState({ error: null })}
            >
              Попробовать снова
            </button>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
