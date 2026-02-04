import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'

// <StrictMode> etiketi bilerek kaldırıldı.
// Böylece WebSocket bağlantısı iki kere açılıp kapanmaz.
createRoot(document.getElementById('root')).render(
  <App />
)