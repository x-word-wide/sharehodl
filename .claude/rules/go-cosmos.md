---
paths:
  - "x/**/*.go"
  - "app/**/*.go"
  - "cmd/**/*.go"
---

# Go / Cosmos SDK Rules

## Module Structure
- Follow keeper pattern: state management in keeper, business logic in msg_server
- All messages must implement `sdk.Msg` interface with `ValidateBasic()`
- Use `errorsmod.Wrap()` for error context, never raw `errors.New()`
- Emit events for all state changes for indexer consumption

## State Management
- Use typed store keys from `types/keys.go`
- Prefix all keys with module name
- Use `collections` package for complex data structures
- Never store more than necessary - derive when possible

## Error Handling
- Return errors, never panic in message handlers
- Use module-specific error types from `types/errors.go`
- Include context in error messages (addresses, amounts)
- Log errors at appropriate levels

## Security
- Validate all inputs in `ValidateBasic()` AND in keeper
- Check authorization before state changes
- Use `sdk.AccAddress` for address validation
- Never trust client-provided data

## Testing
- Table-driven tests for all keeper methods
- Mock bank keeper for token operations
- Test genesis import/export roundtrip
- Test error conditions, not just happy path

## Proto Generation
- Define messages in `proto/sharehodl/{module}/v1/`
- Run `make proto-gen` after proto changes
- Keep proto files in sync with Go types
