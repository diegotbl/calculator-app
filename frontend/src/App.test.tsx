import { render, screen } from '@testing-library/react'
import App from './App'

// Smoke test: confirms Vitest + jsdom + React Testing Library + jest-dom
// matchers are all wired up. Replace with real Calculator tests later.
describe('App', () => {
  it('renders the Vite + React heading', () => {
    render(<App />)
    expect(screen.getByText('Vite + React')).toBeInTheDocument()
  })
})
