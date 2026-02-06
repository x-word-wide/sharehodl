---
paths:
  - "**/*.go"
  - "**/*.ts"
  - "**/*.tsx"
---

# Security Rules

## CRITICAL - Never Allow
- Hardcoded secrets, keys, or passwords
- SQL injection vulnerabilities
- Command injection
- Unvalidated user input in state changes
- Exposing private keys in logs or errors
- Panics in message handlers (DoS vector)

## Blockchain Security
- Always validate collateral ratios before operations
- Check authorization before state mutations
- Use atomic operations for multi-step processes
- Emit events for audit trail
- Validate addresses are valid bech32

## Input Validation
- Validate in ValidateBasic() AND keeper methods
- Check bounds on all numeric inputs
- Sanitize strings before storage
- Verify array lengths to prevent DoS

## Authentication & Authorization
- Verify message signer matches authorized party
- Check permissions before sensitive operations
- Use capability-based security where possible
- Log authorization failures

## Financial Operations
- Use safe math (check for overflow/underflow)
- Verify sufficient balances before transfers
- Use atomic transactions for swaps
- Double-check fee calculations

## Frontend Security
- Never store secrets in client code
- Sanitize all user inputs before display
- Use HTTPS for all API calls
- Implement proper CORS policies
- No dangerouslySetInnerHTML with user data

## Secrets Management
- Use environment variables for secrets
- Never commit .env files
- Rotate keys regularly
- Use separate keys per environment
