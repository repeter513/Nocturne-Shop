// Application entry point — mounts React root. / Точка входа приложения — монтирование React-корня.
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './index.css'

// StrictMode double-invokes effects in dev to surface side-effect bugs. / StrictMode дважды вызывает effects в dev для выявления побочных эффектов.
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
