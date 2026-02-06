---
paths:
  - "**/*_test.go"
  - "**/*.test.ts"
  - "**/*.test.tsx"
  - "**/*.spec.ts"
  - "tests/**/*"
---

# Testing Rules

## Test Coverage Targets
- Keeper methods: >90%
- Message handlers: >85%
- React components: >80%
- Critical paths: 100%

## Go Testing
- Use table-driven tests
- Test both success and error cases
- Mock external dependencies (bank keeper, etc.)
- Use testify/require for assertions
- Name tests: Test{Function}_{Scenario}

## React Testing
- Use React Testing Library
- Test user interactions, not implementation
- Mock API responses with MSW
- Test accessibility with jest-axe
- Write integration tests for complex flows

## Test Patterns

### Happy Path
Always test the expected successful case first.

### Edge Cases
- Empty inputs
- Maximum values
- Minimum values
- Boundary conditions

### Error Cases
- Invalid inputs
- Missing permissions
- Insufficient funds
- Network failures

### Regression Tests
Every bug fix MUST include a test that would have caught it.

## Test Data
- Use realistic test data
- Don't hardcode addresses - generate or use fixtures
- Keep test data in test files or fixtures folder
- Clean up state between tests

## Performance Tests
- Benchmark critical paths
- Test with realistic data volumes
- Monitor for memory leaks
- Check for N+1 query patterns
