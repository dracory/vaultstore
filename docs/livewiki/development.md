---
path: development.md
page-type: tutorial
summary: Development workflow, testing, and contributing guidelines.
tags: [development, testing, contributing, workflow]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
---

# Development

This document covers the development workflow, testing procedures, and contribution guidelines for VaultStore.

## Development Environment Setup

### Prerequisites

- Go 1.25 or higher
- Git
- Make (optional, for task management)
- Docker (optional, for database testing)

### Clone and Setup

```bash
# Clone the repository
git clone https://github.com/dracory/vaultstore.git
cd vaultstore

# Install dependencies
go mod download

# Run tests to verify setup
go test ./...
```

### Development Tools

#### Task Management (Taskfile)

VaultStore uses Task for task management:

```bash
# Install Task (if not already installed)
go install github.com/go-task/task/v3/cmd/task@latest

# List available tasks
task --list

# Run common tasks
task test          # Run all tests
task build         # Build the project
task lint          # Run linters
task fmt           # Format code
```

#### Available Tasks

```yaml
# From taskfile.yml
version: '3'

tasks:
  default:
    desc: Run tests
    cmds:
      - go test ./...

  test:
    desc: Run all tests
    cmds:
      - go test -v ./...

  test-coverage:
    desc: Run tests with coverage
    cmds:
      - go test -coverprofile=coverage.out ./...
      - go tool cover -html=coverage.out -o coverage.html

  build:
    desc: Build the project
    cmds:
      - go build ./...

  lint:
    desc: Run linters
    cmds:
      - golangci-lint run

  fmt:
    desc: Format code
    cmds:
      - go fmt ./...

  clean:
    desc: Clean build artifacts
    cmds:
      - go clean
      - rm -f coverage.out coverage.html
```

## Code Structure

### Project Layout

```
vaultstore/
├── docs/                   # Documentation
│   ├── livewiki/           # LiveWiki documentation
│   ├── proposals/          # Design proposals
│   └── *.md                # Other docs
├── .github/                # GitHub workflows
│   └── workflows/
│       └── tests.yml       # CI/CD pipeline
├── interfaces.go           # Core interfaces
├── store_*.go             # Store implementation
├── record_*.go             # Record operations
├── token_*.go              # Token operations
├── encdec*.go              # Encryption/decryption
├── functions.go           # Utility functions
├── consts.go               # Constants
├── gorm_model.go           # GORM models
├── is_token.go             # Token validation
├── sqls.go                 # SQL queries
├── *_test.go               # Test files
├── go.mod                  # Go module
├── go.sum                  # Go checksums
├── README.md               # Project README
└── LICENSE                 # License file
```

### Core Components

#### Interfaces (`interfaces.go`)

Defines the main interfaces:
- `StoreInterface` - Main store operations
- `RecordInterface` - Record data operations
- `RecordQueryInterface` - Query building operations

#### Store Implementation (`store_*.go`)

Main store implementation files:
- `store_new.go` - Store factory function
- `store_implementation.go` - Core store logic
- `store_record_methods.go` - Record-related methods
- `store_token_methods.go` - Token-related methods
- `store_record_query.go` - Query implementation

#### Encryption (`encdec*.go`)

Encryption and decryption utilities:
- `encdec.go` - Main encryption/decryption functions
- `encdec_test.go` - Encryption tests
- `encdec_v2_test.go` - Enhanced encryption tests

#### Models (`gorm_model.go`)

GORM database models and schema definitions.

## Testing

### Test Structure

#### Unit Tests

Each module has corresponding test files:
- `store_implementation_test.go`
- `store_record_methods_test.go`
- `store_token_methods_test.go`
- `record_implementation_test.go`
- `encdec_test.go`
- `functions_test.go`
- `is_token_test.go`

#### Integration Tests

- `benchmark_test.go` - Performance benchmarks
- Database-specific tests

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test -run TestTokenCreate ./...

# Run benchmarks
go test -bench=. ./...

# Run race condition tests
go test -race ./...
```

### Test Database Setup

Tests use in-memory SQLite databases by default:

```go
func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    return db
}
```

### Writing Tests

#### Test Structure

```go
func TestTokenCreate(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    vault, err := NewStore(NewStoreOptions{
        VaultTableName:     "test_vault",
        DB:                 db,
        AutomigrateEnabled: true,
    })
    require.NoError(t, err)

    // Test
    token, err := vault.TokenCreate(context.Background(), "test_value", "password", 32)
    require.NoError(t, err)
    assert.NotEmpty(t, token)

    // Verify
    exists, err := vault.TokenExists(context.Background(), token)
    require.NoError(t, err)
    assert.True(t, exists)
}
```

#### Test Helpers

```go
// Helper function to create test store
func createTestStore(t *testing.T) StoreInterface {
    db := setupTestDB(t)
    vault, err := NewStore(NewStoreOptions{
        VaultTableName:     "test_vault",
        DB:                 db,
        AutomigrateEnabled: true,
    })
    require.NoError(t, err)
    return vault
}

// Helper function to create test token
func createTestToken(t *testing.T, vault StoreInterface) string {
    token, err := vault.TokenCreate(context.Background(), "test_value", "password", 32)
    require.NoError(t, err)
    return token
}
```

### Benchmark Testing

```go
func BenchmarkTokenCreate(b *testing.B) {
    vault := createTestStore(&testing.T{})
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := vault.TokenCreate(ctx, "test_value", "password", 32)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Code Quality

### Linting

VaultStore uses golangci-lint for code quality:

```bash
# Install golangci-lint
go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run specific linter
golangci-lint run --enable-all
```

### Configuration

`.golangci.yml` configuration:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gosec
    - misspell
    - unconvert
    - dupl
    - goconst
    - gocyclo

linters-settings:
  goconst:
    min-len: 3
    min-occurrences: 3
  gocyclo:
    min-complexity: 10
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Import organization
goimports -w .

# Check formatting
gofmt -l .
```

## Contributing

### Contribution Workflow

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/new-feature`
3. **Make** your changes
4. **Test** your changes: `go test ./...`
5. **Lint** your code: `golangci-lint run`
6. **Commit** your changes: `git commit -m "Add new feature"`
7. **Push** to your fork: `git push origin feature/new-feature`
8. **Create** a Pull Request

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

**Examples:**
```
feat(token): add token expiration support

Add optional expiration time for tokens to enable
time-limited access to stored secrets.

Closes #123
```

```
fix(encryption): handle empty password correctly

Fix decryption when password is empty string.
Previously this would cause panic.
```

### Pull Request Guidelines

#### PR Template

```markdown
## Description
Brief description of the change.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests for new functionality
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] CHANGELOG.md updated (if applicable)
```

#### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Code Review**: At least one maintainer must review
3. **Approval**: PR requires approval before merge
4. **Merge**: Squash and merge to maintain clean history

## Release Process

### Version Management

VaultStore follows semantic versioning (SemVer):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

1. **Update** version in `go.mod`
2. **Update** CHANGELOG.md
3. **Create** git tag: `git tag v1.2.3`
4. **Push** tag: `git push origin v1.2.3`
5. **GitHub Actions** will automatically create release

### Changelog

Maintain `docs/changelog.md` with:

```markdown
# Changelog

## [1.2.3] - 2026-02-03

### Added
- Token expiration support
- Custom token creation

### Fixed
- Password validation bug
- Memory leak in encryption

### Changed
- Improved error messages
- Updated dependencies

### Deprecated
- Old token validation method (will be removed in 2.0.0)

### Security
- Improved encryption key derivation
```

## Performance Optimization

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.

# Memory profiling
go test -memprofile=mem.prof -bench=.

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. -benchmem

# Compare benchmarks
go test -bench=. -count=5 | benchstat
```

### Optimization Strategies

1. **Database Optimization**
   - Proper indexing
   - Connection pooling
   - Query optimization

2. **Memory Management**
   - Reduce allocations
   - Pool reuse
   - Efficient data structures

3. **Encryption Optimization**
   - Key reuse where safe
   - Efficient cipher operations
   - Minimize data copying

## Debugging

### Debug Mode

Enable debug logging:

```go
vault, err := NewStore(NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    DebugEnabled:   true,
})
```

### Logging

VaultStore uses standard library logging:

```go
// Enable debug logging to see SQL queries
vault.EnableDebug(true)

// Check debug status
if vault.GetDbDriverName() == "sqlite" {
    fmt.Println("Using SQLite database")
}
```

### Common Debugging Scenarios

#### Database Issues

```go
// Check database connection
db, err := sql.Open("sqlite", "./test.db")
if err != nil {
    log.Fatal("Database connection failed:", err)
}

// Test database
err = db.Ping()
if err != nil {
    log.Fatal("Database ping failed:", err)
}
```

#### Encryption Issues

```go
// Test encryption/decryption
value := "test value"
password := "test password"
encrypted, err := encrypt(value, password)
if err != nil {
    log.Fatal("Encryption failed:", err)
}

decrypted, err := decrypt(encrypted, password)
if err != nil {
    log.Fatal("Decryption failed:", err)
}

if decrypted != value {
    log.Fatal("Decryption mismatch")
}
```

## See Also

- [Architecture](architecture.md) - System design overview
- [API Reference](api_reference.md) - Complete API documentation
- [Getting Started](getting_started.md) - Setup and usage guide
- [GitHub Repository](https://github.com/dracory/vaultstore) - Source code and issues
