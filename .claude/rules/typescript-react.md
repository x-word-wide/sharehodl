---
paths:
  - "sharehodl-frontend/**/*.ts"
  - "sharehodl-frontend/**/*.tsx"
---

# TypeScript / React Rules

## TypeScript
- Use strict mode always
- Define interfaces for all props and API responses
- Avoid `any` - use `unknown` and type guards instead
- Use discriminated unions for state machines

## React
- Use functional components with hooks
- Prefer composition over inheritance
- Keep components small and focused
- Use custom hooks for reusable logic

## Next.js
- Use App Router conventions
- Server components by default, 'use client' only when needed
- Use route handlers for API endpoints
- Leverage ISR for semi-static content

## State Management
- Local state with useState for component state
- React Query/SWR for server state
- Context for cross-cutting concerns (theme, auth)
- Avoid prop drilling - use composition

## API Integration
- Type all API responses
- Handle loading, error, and empty states
- Use optimistic updates where appropriate
- Implement proper error boundaries

## Performance
- Memoize expensive computations with useMemo
- Memoize callbacks with useCallback when passed as props
- Use React.lazy for code splitting
- Avoid unnecessary re-renders

## Testing
- Test user behavior, not implementation
- Use React Testing Library
- Mock API calls, not internal modules
- Write E2E tests for critical flows
