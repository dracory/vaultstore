---
path: conventions.md
page-type: reference
summary: Coding and documentation standards for VaultStore contributors.
tags: [conventions, standards, guidelines, contributing]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
---

# Conventions

This document outlines the coding, documentation, and contribution conventions for VaultStore.

## Code Style

### Go Formatting

Follow standard Go formatting conventions:

```bash
# Format code
go fmt ./...

# Import organization
goimports -w .

# Lint code
golangci-lint run
```

### Naming Conventions

#### Package Names

- Use short, lowercase package names
- Avoid unnecessary abbreviations
- Package name should match directory name

```go
package vaultstore  // Good
package vs         // Bad - too abbreviated
package vault_store // Bad - underscores not recommended
```

#### Interface Names

- Interface names should be descriptive and end with "Interface"
- Use simple, method-based names for one-method interfaces

```go
type StoreInterface interface { ... }           // Good
type RecordInterface interface { ... }         // Good
type Validator interface {                     // Good - one method
    Validate() error
}
```

#### Struct Names

- Use PascalCase for exported types
- Use camelCase for unexported types
- Be descriptive but concise

```go
type storeImplementation struct { ... }  // Good - unexported
type RecordQuery struct { ... }         // Good - exported
type VaultStore struct { ... }          // Good
```

#### Method Names

- Use PascalCase for exported methods
- Use camelCase for unexported methods
- Use descriptive names that indicate action

```go
func (s *storeImplementation) TokenCreate(...) (string, error) { ... }  // Good
func (s *storeImplementation) validateToken(...) error { ... }          // Good
func (s *storeImplementation) create(...) { ... }                      // Bad - not descriptive
```

#### Variable Names

- Use camelCase for local variables
- Use short names for limited scope (i, err, ctx)
- Use descriptive names for broader scope

```go
// Good
func TokenCreate(ctx context.Context, value string, password string) (string, error) {
    token, err := generateToken(32)
    if err != nil {
        return "", err
    }
    
    encryptedValue, err := encrypt(value, password)
    if err != nil {
        return "", err
    }
    
    return token, nil
}

// Bad - too abbreviated
func TC(c context.Context, v string, p string) (string, error) {
    t, e := genTkn(32)
    if e != nil {
        return "", e
    }
    
    ev, e := enc(v, p)
    if e != nil {
        return "", e
    }
    
    return t, nil
}
```

### Function Organization

#### Function Structure

```go
func (s *storeImplementation) TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (string, error) {
    // 1. Input validation
    if err := s.validateTokenCreateParams(value, password, tokenLength); err != nil {
        return "", err
    }
    
    // 2. Generate token
    token, err := s.generateSecureToken(tokenLength)
    if err != nil {
        return "", fmt.Errorf("token generation failed: %w", err)
    }
    
    // 3. Encrypt value
    encryptedValue, err := encrypt(value, password)
    if err != nil {
        return "", fmt.Errorf("encryption failed: %w", err)
    }
    
    // 4. Create record
    record := s.newRecord(token, encryptedValue, options...)
    
    // 5. Store in database
    if err := s.RecordCreate(ctx, record); err != nil {
        return "", fmt.Errorf("record creation failed: %w", err)
    }
    
    return token, nil
}
```

#### Error Handling

- Always handle errors
- Use fmt.Errorf for error wrapping
- Define specific error types

```go
// Good
func (s *storeImplementation) TokenRead(ctx context.Context, token string, password string) (string, error) {
    record, err := s.RecordFindByToken(ctx, token)
    if err != nil {
        return "", fmt.Errorf("token lookup failed: %w", err)
    }
    
    if record.GetSoftDeletedAt() != "" {
        return "", ErrRecordNotFound
    }
    
    value, err := decrypt(record.GetValue(), password)
    if err != nil {
        return "", fmt.Errorf("decryption failed: %w", err)
    }
    
    return value, nil
}

// Bad - ignoring errors
func (s *storeImplementation) TokenRead(ctx context.Context, token string, password string) string {
    record, _ := s.RecordFindByToken(ctx, token)  // Bad - ignoring error
    value, _ := decrypt(record.GetValue(), password)  // Bad - ignoring error
    return value
}
```

## Documentation Standards

### Function Documentation

Use standard Go doc format:

```go
// TokenCreate creates a new token with the specified value and optional password protection.
// The token is generated using cryptographically secure random generation and the value
// is encrypted using AES-256-GCM. If a password is provided, it will be required for
// subsequent read operations.
//
// Parameters:
//   - ctx: Context for the operation
//   - value: The value to encrypt and store
//   - password: Optional password for encryption (can be empty)
//   - tokenLength: Length of the generated token (recommended: 32+ characters)
//   - options: Optional token creation options (expiration, etc.)
//
// Returns:
//   - string: The generated token
//   - error: Error if creation fails
//
// Example:
//   token, err := vault.TokenCreate(ctx, "my_secret", "password123", 32)
//   if err != nil {
//       log.Fatal(err)
//   }
func (s *storeImplementation) TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (string, error) {
    // Implementation...
}
```

### Type Documentation

```go
// StoreInterface defines the main interface for vault operations including
// record management, token operations, and database management.
//
// Implementations should be thread-safe and handle context cancellation
// appropriately. All operations should return meaningful errors that
// can be inspected by callers.
type StoreInterface interface {
    // Methods...
}
```

### Package Documentation

```go
// Package vaultstore provides a secure value storage implementation for Go.
// It offers token-based access to encrypted values with optional password
// protection, flexible querying, and soft delete functionality.
//
// Basic usage:
//
//   vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
//       VaultTableName:     "vault",
//       DB:                 db,
//       AutomigrateEnabled: true,
//   })
//   if err != nil {
//       log.Fatal(err)
//   }
//
//   token, err := vault.TokenCreate(ctx, "secret_value", "password", 32)
//   if err != nil {
//       log.Fatal(err)
//   }
//
//   value, err := vault.TokenRead(ctx, token, "password")
//   if err != nil {
//       log.Fatal(err)
//   }
package vaultstore
```

## Testing Conventions

### Test File Naming

- Test files should end with `_test.go`
- Use the same name as the file being tested

```
store_implementation.go    -> store_implementation_test.go
token_operations.go        -> token_operations_test.go
encryption.go             -> encryption_test.go
```

### Test Function Naming

- Test functions should start with `Test`
- Use descriptive names that indicate what is being tested
- Use subtests for related test cases

```go
func TestTokenCreate(t *testing.T) {
    t.Run("with valid parameters", func(t *testing.T) {
        // Test valid token creation
    })
    
    t.Run("with empty value", func(t *testing.T) {
        // Test error handling
    })
    
    t.Run("with password protection", func(t *testing.T) {
        // Test password-protected tokens
    })
}
```

### Test Structure

```go
func TestTokenLifecycle(t *testing.T) {
    // Setup
    vault := createTestStore(t)
    ctx := context.Background()
    
    // Test
    token, err := vault.TokenCreate(ctx, "test_value", "password", 32)
    require.NoError(t, err)
    assert.NotEmpty(t, token)
    
    // Verify
    exists, err := vault.TokenExists(ctx, token)
    require.NoError(t, err)
    assert.True(t, exists)
    
    // Cleanup (if needed)
    err = vault.TokenDelete(ctx, token)
    require.NoError(t, err)
}
```

### Test Helpers

```go
// Helper function for creating test store
func createTestStore(t *testing.T) StoreInterface {
    db, err := sql.Open("sqlite", ":memory:")
    require.NoError(t, err)
    
    vault, err := NewStore(NewStoreOptions{
        VaultTableName:     "test_vault",
        DB:                 db,
        AutomigrateEnabled: true,
    })
    require.NoError(t, err)
    return vault
}

// Helper function for creating test token
func createTestToken(t *testing.T, vault StoreInterface) string {
    token, err := vault.TokenCreate(context.Background(), "test_value", "password", 32)
    require.NoError(t, err)
    return token
}
```

## Error Handling Conventions

### Error Types

Define specific error types for different failure modes:

```go
var (
    // Validation errors
    ErrTokenRequired     = errors.New("token is required")
    ErrPasswordRequired  = errors.New("password is required")
    ErrInvalidTokenLength = errors.New("invalid token length")
    
    // Database errors
    ErrRecordNotFound    = errors.New("record not found")
    ErrTokenAlreadyExists = errors.New("token already exists")
    
    // Security errors
    ErrInvalidPassword   = errors.New("invalid password")
    ErrDecryptionFailed  = errors.New("decryption failed")
)
```

### Error Wrapping

Use fmt.Errorf for error wrapping with context:

```go
// Good
return "", fmt.Errorf("token generation failed: %w", err)

// Bad
return "", err  // Loses context
```

### Error Messages

- Write clear, actionable error messages
- Include relevant context without sensitive data
- Use consistent formatting

```go
// Good
return "", fmt.Errorf("token '%s' not found or expired", token)

// Bad - too generic
return "", errors.New("error")

// Bad - includes sensitive data
return "", fmt.Errorf("failed to decrypt '%s' with password '%s'", value, password)
```

## Commit Message Conventions

### Format

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions/changes
- `chore`: Maintenance tasks

### Examples

```
feat(token): add token expiration support

Add optional expiration time for tokens to enable
time-limited access to stored secrets. Tokens can be
created with ExpiresAt option and will be automatically
excluded from queries after expiration.

Closes #123
```

```
fix(encryption): handle empty password correctly

Fix decryption when password is empty string.
Previously this would cause panic in key derivation.

Fixes #456
```

## Code Review Conventions

### Review Checklist

- [ ] Code follows project style guidelines
- [ ] Functions are properly documented
- [ ] Error handling is comprehensive
- [ ] Tests are provided and pass
- [ ] No sensitive data in logs or errors
- [ ] Performance considerations addressed
- [ ] Security implications considered

### Review Comments

- Be constructive and specific
- Explain the "why" behind suggestions
- Use issue numbers for tracking
- Keep comments focused on code quality

## Documentation Conventions

### LiveWiki Standards

All LiveWiki pages must include the enhanced metadata header:

```markdown
---
path: filename.md
page-type: overview|reference|tutorial|module|changelog
summary: One-line description of this page's content and purpose.
tags: [tag1, tag2, tag3]
created: YYYY-MM-DD
updated: YYYY-MM-DD
version: X.Y.Z
---
```

### Page Types

- `overview`: High-level introductions and architectural summaries
- `reference`: API docs, configuration options, technical specifications
- `tutorial`: Step-by-step guides and getting started content
- `module`: Documentation for a specific module/package
- `changelog`: Version history and change logs

### Cross-References

Include "See Also" sections at the bottom of each page:

```markdown
## See Also

- [Getting Started](getting_started.md) - Setup and usage guide
- [API Reference](api_reference.md) - Complete API documentation
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
```

## Security Conventions

### Sensitive Data

- Never log passwords or tokens
- Never include sensitive data in error messages
- Use secure random generation for tokens
- Validate all input parameters

### Security Testing

- Test encryption/decryption with wrong passwords
- Test token generation uniqueness
- Test edge cases for security functions
- Include security-focused test cases

## Performance Conventions

### Database Operations

- Use appropriate database indexes
- Limit query results with pagination
- Select only needed columns
- Use connection pooling

### Memory Management

- Avoid memory leaks in long-running operations
- Use efficient data structures
- Clean up resources properly
- Profile memory usage in tests

## Version Management

### Semantic Versioning

Follow SemVer for version numbers:
- `MAJOR.MINOR.PATCH`
- Increment MAJOR for breaking changes
- Increment MINOR for new features (backward compatible)
- Increment PATCH for bug fixes (backward compatible)

### Changelog Maintenance

Maintain `docs/changelog.md` with:
- Version number and release date
- Added, Fixed, Changed, Deprecated sections
- Security section for security-related changes
- Links to relevant issues or PRs

## See Also

- [Development](development.md) - Development workflow and testing
- [Getting Started](getting_started.md) - Setup and installation
- [API Reference](api_reference.md) - Complete API documentation
- [Contributing Guide](https://github.com/dracory/vaultstore/blob/main/CONTRIBUTING.md) - GitHub contribution guidelines
