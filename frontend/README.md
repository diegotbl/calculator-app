# Frontend

React + TypeScript (Vite) client for the calculator API. The app, its design decisions, and
the API contract are documented in the repo root: [`../README.md`](../README.md) and
[`../DECISIONS.md`](../DECISIONS.md).

```
npm install
npm run dev              # Vite dev server on :5173, proxies /calculate to :8080
npm test                 # Vitest run (npm run test:coverage for the coverage summary)
```

The dev server proxies `/calculate` to the Go backend on `:8080`, so start the backend first.
