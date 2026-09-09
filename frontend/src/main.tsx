import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import Calculator from './Calculator.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Calculator />
  </StrictMode>,
)
