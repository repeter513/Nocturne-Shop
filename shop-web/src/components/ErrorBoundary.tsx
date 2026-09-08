import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

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
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => window.location.assign('/')}
            >
              В каталог
            </button>
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
