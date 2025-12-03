# Contributing to ShareHODL

Thank you for your interest in contributing to ShareHODL! This document provides guidelines for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Process](#development-process)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Security](#security)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to conduct@sharehodl.io.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/your-username/sharehodl-blockchain.git
   cd sharehodl-blockchain
   ```
3. **Set up development environment**:
   ```bash
   make install-deps
   make build
   ```

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Make
- Git

## Development Process

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test improvements

### Commit Messages

Follow conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

Examples:
- `feat(dex): add order book matching engine`
- `fix(hodl): resolve stablecoin minting bug`
- `docs(api): update governance endpoint documentation`

## Code Style

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` and `goimports`
- Run `golint` and `go vet`
- Maximum line length: 120 characters
- Use meaningful variable and function names

### Documentation

- Document all public functions and types
- Include examples for complex functionality
- Update README.md for significant changes
- Keep inline comments concise and helpful

## Testing

### Running Tests

```bash
# Unit tests
make test

# Integration tests
make test-integration

# E2E tests (requires more setup)
make test-e2e

# All tests
make test-all
```

### Test Requirements

- **Unit tests**: Required for all new functionality
- **Integration tests**: Required for module interactions
- **E2E tests**: Required for user-facing features
- **Benchmark tests**: Required for performance-critical code

### Test Guidelines

- Write clear, descriptive test names
- Test both success and failure cases
- Mock external dependencies
- Maintain test coverage above 80%

## Submitting Changes

### Pull Request Process

1. **Update your branch**:
   ```bash
   git checkout main
   git pull upstream main
   git checkout your-feature-branch
   git rebase main
   ```

2. **Run tests and linting**:
   ```bash
   make test-all
   make lint
   make format
   ```

3. **Create Pull Request**:
   - Use a clear, descriptive title
   - Include detailed description of changes
   - Reference related issues
   - Add screenshots for UI changes

### PR Requirements

- [ ] Tests pass
- [ ] Code is properly formatted
- [ ] Documentation is updated
- [ ] No security vulnerabilities introduced
- [ ] Performance impact assessed
- [ ] Backwards compatibility maintained

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and security scans
2. **Peer Review**: At least one team member reviews code
3. **Security Review**: Security team reviews sensitive changes
4. **Final Review**: Maintainer approves and merges

## Module-Specific Guidelines

### HODL Stablecoin Module

- Ensure price stability mechanisms are tested
- Validate mint/burn operations thoroughly
- Test edge cases for collateral management

### Equity Module

- Verify compliance with securities regulations
- Test dividend distribution accuracy
- Ensure share transfer validity

### DEX Module

- Test order matching algorithms extensively
- Validate liquidity pool mechanics
- Ensure fair price discovery

### Governance Module

- Test voting mechanisms thoroughly
- Validate proposal execution logic
- Ensure proper access controls

## Security Considerations

### Security Review Required

- Cryptographic code changes
- Authentication/authorization changes
- Financial transaction logic
- External API integrations
- Deployment configuration changes

### Security Guidelines

- Never commit secrets or credentials
- Use secure coding practices
- Follow principle of least privilege
- Validate all inputs
- Use established cryptographic libraries

## Performance Guidelines

- Profile performance-critical code
- Avoid premature optimization
- Document performance requirements
- Use benchmarks for verification
- Consider memory usage in algorithms

## Documentation Standards

### Code Documentation

```go
// ProcessTransaction processes a blockchain transaction and updates state.
// It validates the transaction, executes business logic, and commits changes.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - tx: Transaction to process
//
// Returns:
//   - Response with transaction result
//   - Error if processing fails
func ProcessTransaction(ctx context.Context, tx Transaction) (*Response, error) {
    // Implementation...
}
```

### API Documentation

- Use OpenAPI/Swagger specifications
- Include request/response examples
- Document error codes and meanings
- Provide authentication requirements

## Getting Help

### Community Support

- **Discord**: [ShareHODL Discord](https://discord.gg/sharehodl)
- **Forum**: [ShareHODL Forum](https://forum.sharehodl.io)
- **Documentation**: [docs.sharehodl.io](https://docs.sharehodl.io)

### Developer Resources

- **Development Guide**: See `docs/DEVELOPMENT.md`
- **Architecture Overview**: See `ARCHITECTURE.md`
- **API Reference**: See `docs/api/`

### Contact

- **General Questions**: dev@sharehodl.io
- **Security Issues**: security@sharehodl.io
- **Urgent Issues**: emergency@sharehodl.io

## Recognition

Contributors are recognized in:
- Release notes
- Contributors file
- Community events
- Developer documentation

Thank you for contributing to ShareHODL! Your efforts help build the future of decentralized equity markets.