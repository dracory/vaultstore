---
path: modules/token_operations.md
page-type: module
summary: Token lifecycle management and secure access operations including BulkRekey functionality.
tags: [module, token, operations, security, bulk-rekey]
created: 2026-02-03
updated: 2026-02-03
version: 1.1.0
---

# Token Operations Module

The token operations module manages the complete lifecycle of tokens, providing secure access to stored values through token-based authentication and encryption.

## Overview

The token operations module handles:
- Token generation and validation
- Secure value encryption and decryption
- Token lifecycle management (create, read, update, delete)
- Expiration handling and cleanup
- Password-based access control

## Core Operations

### Token Creation

#### TokenCreate

Creates a new token with automatically generated token value:

```go
func (s *storeImplementation) TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (string, error)
```

**Parameters:**
- `ctx context.Context`: Operation context
- `value string`: Value to encrypt and store
- `password string`: Optional password for encryption
- `tokenLength int`: Length of generated token (recommended: 32+)
- `options ...TokenCreateOptions`: Optional creation parameters

**Returns:**
- `string`: Generated token
- `error`: Error if creation fails

**Process:**
1. Generate cryptographically secure random token
2. Encrypt value using password (if provided)
3. Create record with token and encrypted value
4. Set creation and optional expiration timestamps
5. Store in database
6. Return generated token

#### TokenCreateCustom

Creates a token with a custom token value:

```go
func (s *storeImplementation) TokenCreateCustom(ctx context.Context, token string, value string, password string, options ...TokenCreateOptions) error
```

**Use Cases:**
- Predictable token values for testing
- Integration with existing token systems
- Custom token formats

### Token Retrieval

#### TokenRead

Retrieves and decrypts a value using a token and password:

```go
func (s *storeImplementation) TokenRead(ctx context.Context, token string, password string) (string, error)
```

**Process:**
1. Find record by token
2. Verify token is not soft deleted or expired
3. Decrypt value using password
4. Return decrypted value

**Security Features:**
- Validates token existence and status
- Checks expiration time
- Verifies password for decryption
- Prevents access to deleted/expired tokens

#### TokensRead

Batch read multiple tokens efficiently:

```go
func (s *storeImplementation) TokensRead(ctx context.Context, tokens []string, password string) (map[string]string, error)
```

**Benefits:**
- Single database query for multiple tokens
- Efficient batch decryption
- Error aggregation for partial failures

### Token Updates

#### TokenUpdate

Updates a token's encrypted value:

```go
func (s *storeImplementation) TokenUpdate(ctx context.Context, token string, value string, password string) error
```

**Process:**
1. Find existing record by token
2. Validate current access (password if required)
3. Encrypt new value with password
4. Update record with new encrypted value
5. Update modification timestamp

#### TokenRenew

Updates token expiration time:

```go
func (s *storeImplementation) TokenRenew(ctx context.Context, token string, expiresAt time.Time) error
```

**Use Cases:**
- Extend access for active sessions
- Renew temporary credentials
- Implement token refresh policies

### Token Deletion

#### TokenSoftDelete

Logical deletion with recovery capability:

```go
func (s *storeImplementation) TokenSoftDelete(ctx context.Context, token string) error
```

**Features:**
- Sets `soft_deleted_at` timestamp
- Token remains in database for recovery
- Excluded from normal queries
- Can be restored if needed

#### TokenDelete

Permanent deletion of token:

```go
func (s *storeImplementation) TokenDelete(ctx context.Context, token string) error
```

**Features:**
- Physical removal from database
- Cannot be recovered
- Compliance with data retention policies
- Complete data destruction

### Token Validation

#### TokenExists

Check if token exists and is accessible:

```go
func (s *storeImplementation) TokenExists(ctx context.Context, token string) (bool, error)
```

**Validation Checks:**
- Token exists in database
- Token is not soft deleted
- Token has not expired
- Returns boolean result

## Token Creation Options

### TokenCreateOptions

```go
type TokenCreateOptions struct {
    ExpiresAt time.Time // Optional expiration time
}
```

### Expiration Management

#### Setting Expiration

```go
// Expire in 24 hours
token, err := vault.TokenCreate(ctx, value, password, 32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Now().Add(24 * time.Hour),
    })

// Expire at specific time
token, err := vault.TokenCreate(ctx, value, password, 32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
    })

// No expiration (permanent)
token, err := vault.TokenCreate(ctx, value, password, 32)
```

#### Expiration Cleanup

```go
// Soft delete expired tokens
count, err := vault.TokensExpiredSoftDelete(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Soft deleted %d expired tokens\n", count)

// Hard delete expired tokens
count, err = vault.TokensExpiredDelete(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Hard deleted %d expired tokens\n", count)
```

## Security Features

### Token Generation

#### Cryptographically Secure Tokens

```go
// Token generation uses crypto/rand for security
func generateSecureToken(length int) (string, error) {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}
```

#### Token Length Recommendations

| Length | Security Level | Use Case |
|--------|----------------|----------|
| 16 chars | Basic | Development, testing |
| 32 chars | Good | Most applications |
| 64 chars | High | Sensitive data |

### Encryption Integration

#### Password-Based Encryption

```go
// Value encryption process
func encryptValue(value, password string) (string, error) {
    // Derive key from password using PBKDF2
    key := deriveKey(password, salt)
    
    // Encrypt using AES-256-GCM
    encrypted, err := aesGCMEncrypt(key, value)
    if err != nil {
        return "", err
    }
    
    return encrypted, nil
}
```

#### Security Measures

- **AES-256-GCM**: Authenticated encryption
- **PBKDF2**: Password-based key derivation
- **Random Salts**: Unique salt per encryption
- **Authentication Tags**: Prevent tampering

### Access Control

#### Password Protection

```go
// Password validation
func validatePassword(password string) error {
    if password == "" {
        return nil // No password is allowed
    }
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    return nil
}
```

#### Token Status Validation

```go
// Comprehensive token validation
func (s *storeImplementation) validateTokenAccess(ctx context.Context, token string) error {
    record, err := s.RecordFindByToken(ctx, token)
    if err != nil {
        return err
    }
    
    // Check soft delete
    if record.GetSoftDeletedAt() != "" {
        return ErrRecordNotFound
    }
    
    // Check expiration
    if record.GetExpiresAt() != "" {
        expiresAt, _ := time.Parse(time.RFC3339, record.GetExpiresAt())
        if time.Now().After(expiresAt) {
            return ErrRecordNotFound
        }
    }
    
    return nil
}
```

## Usage Examples

### Basic Token Operations

```go
// Create a token
token, err := vault.TokenCreate(context.Background(), 
    "my_secret_value", 
    "my_password", 
    32)
if err != nil {
    log.Fatal(err)
}

// Check if token exists
exists, err := vault.TokenExists(context.Background(), token)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Token exists: %t\n", exists)

// Read the value
value, err := vault.TokenRead(context.Background(), token, "my_password")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Value: %s\n", value)

// Update the value
err = vault.TokenUpdate(context.Background(), token, "new_value", "my_password")
if err != nil {
    log.Fatal(err)
}

// Delete the token
err = vault.TokenDelete(context.Background(), token)
if err != nil {
    log.Fatal(err)
}
```

### Advanced Token Management

```go
// Create token with expiration
token, err := vault.TokenCreate(context.Background(), 
    "temporary_secret", 
    "",  // No password
    32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Now().Add(1 * time.Hour),
    })

// Renew token expiration
err = vault.TokenRenew(context.Background(), 
    token, 
    time.Now().Add(24 * time.Hour))

// Batch read multiple tokens
tokens := []string{"token1", "token2", "token3"}
values, err := vault.TokensRead(context.Background(), tokens, "password")
if err != nil {
    log.Fatal(err)
}

for token, value := range values {
    fmt.Printf("%s: %s\n", token, value)
}
```

### Token Lifecycle Management

```go
// Create token for user session
func createSessionToken(vault StoreInterface, userID string, sessionData string) (string, error) {
    token, err := vault.TokenCreate(context.Background(),
        sessionData,
        "", // No password for session tokens
        32,
        vaultstore.TokenCreateOptions{
            ExpiresAt: time.Now().Add(24 * time.Hour),
        })
    if err != nil {
        return "", err
    }
    
    // Store token-to-user mapping in application layer
    err = storeUserToken(userID, token)
    return token, err
}

// Validate and read session
func validateSession(vault StoreInterface, token string) (string, error) {
    exists, err := vault.TokenExists(context.Background(), token)
    if err != nil || !exists {
        return "", errors.New("invalid session")
    }
    
    sessionData, err := vault.TokenRead(context.Background(), token, "")
    if err != nil {
        return "", err
    }
    
    return sessionData, nil
}
```

## Bulk Password Changes

### BulkRekey

Changes the password for all tokens encrypted with a specific password. This is useful for password rotation scenarios.

```go
func (s *storeImplementation) BulkRekey(ctx context.Context, oldPassword, newPassword string) (int, error)
```

**Parameters:**
- `ctx context.Context`: Context for the operation
- `oldPassword string`: Current password used for encryption
- `newPassword string`: New password for re-encryption

**Returns:**
- `int`: Number of tokens re-encrypted
- `error`: Error if operation fails

**Performance:**

| Mode | Complexity | Description |
|------|------------|-------------|
| **With Identity** | O(1) | Direct metadata lookup (fast) |
| **Without Identity** | O(n) | Scan all records (slow) |

**Example:**

```go
// Rekey all tokens using "oldpass" to use "newpass"
count, err := vault.BulkRekey(ctx, "oldpass", "newpass")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Re-encrypted %d tokens\n", count)
```

**With Password Identity Management:**

```go
// Enable identity management for fast bulk rekey
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:          "vault",
    VaultMetaTableName:      "vault_meta",
    DB:                      db,
    PasswordIdentityEnabled: true,  // Enable for O(1) rekey
})

// Fast bulk rekey - only touches relevant records
count, err := vault.BulkRekey(ctx, "oldpass", "newpass")
```

See [Password Identity Management](password_identity_management.md) for more details.

## Performance Optimization

### Batch Operations

```go
// Efficient batch token creation
func createMultipleTokens(vault StoreInterface, values []string) ([]string, error) {
    tokens := make([]string, 0, len(values))
    
    for _, value := range values {
        token, err := vault.TokenCreate(context.Background(), value, "", 32)
        if err != nil {
            return tokens, err
        }
        tokens = append(tokens, token)
    }
    
    return tokens, nil
}
```

### Caching Strategies

```go
// Token existence caching
type TokenCache struct {
    cache map[string]bool
    mutex sync.RWMutex
    ttl   time.Duration
}

func (tc *TokenCache) CheckExists(vault StoreInterface, token string) (bool, error) {
    tc.mutex.RLock()
    if exists, found := tc.cache[token]; found {
        tc.mutex.RUnlock()
        return exists, nil
    }
    tc.mutex.RUnlock()
    
    // Cache miss - check database
    exists, err := vault.TokenExists(context.Background(), token)
    if err != nil {
        return false, err
    }
    
    // Update cache
    tc.mutex.Lock()
    tc.cache[token] = exists
    tc.mutex.Unlock()
    
    return exists, nil
}
```

## Error Handling

### Common Token Errors

```go
var (
    ErrTokenNotFound      = errors.New("token not found")
    ErrTokenExpired       = errors.New("token has expired")
    ErrTokenDeleted       = errors.New("token has been deleted")
    ErrInvalidPassword    = errors.New("invalid password")
    ErrTokenAlreadyExists = errors.New("token already exists")
    ErrInvalidTokenLength = errors.New("invalid token length")
)
```

### Error Handling Patterns

```go
// Comprehensive error handling
func safeTokenRead(vault StoreInterface, token, password string) (string, error) {
    // Check if token exists
    exists, err := vault.TokenExists(context.Background(), token)
    if err != nil {
        return "", fmt.Errorf("token check failed: %w", err)
    }
    if !exists {
        return "", ErrTokenNotFound
    }
    
    // Attempt to read
    value, err := vault.TokenRead(context.Background(), token, password)
    if err != nil {
        if errors.Is(err, ErrInvalidPassword) {
            return "", ErrInvalidPassword
        }
        return "", fmt.Errorf("token read failed: %w", err)
    }
    
    return value, nil
}
```

## Testing

### Unit Tests

```go
func TestTokenLifecycle(t *testing.T) {
    vault := createTestStore(t)
    ctx := context.Background()
    
    // Create
    token, err := vault.TokenCreate(ctx, "test_value", "password", 32)
    require.NoError(t, err)
    assert.NotEmpty(t, token)
    
    // Exists
    exists, err := vault.TokenExists(ctx, token)
    require.NoError(t, err)
    assert.True(t, exists)
    
    // Read
    value, err := vault.TokenRead(ctx, token, "password")
    require.NoError(t, err)
    assert.Equal(t, "test_value", value)
    
    // Update
    err = vault.TokenUpdate(ctx, token, "new_value", "password")
    require.NoError(t, err)
    
    // Verify update
    value, err = vault.TokenRead(ctx, token, "password")
    require.NoError(t, err)
    assert.Equal(t, "new_value", value)
    
    // Delete
    err = vault.TokenDelete(ctx, token)
    require.NoError(t, err)
    
    // Verify deletion
    exists, err = vault.TokenExists(ctx, token)
    require.NoError(t, err)
    assert.False(t, exists)
}
```

### Security Tests

```go
func TestTokenSecurity(t *testing.T) {
    vault := createTestStore(t)
    ctx := context.Background()
    
    // Test with password
    token, err := vault.TokenCreate(ctx, "secret", "password", 32)
    require.NoError(t, err)
    
    // Test wrong password
    _, err = vault.TokenRead(ctx, token, "wrong_password")
    assert.Error(t, err)
    assert.True(t, errors.Is(err, ErrInvalidPassword))
    
    // Test empty password
    _, err = vault.TokenRead(ctx, token, "")
    assert.Error(t, err)
    assert.True(t, errors.Is(err, ErrInvalidPassword))
}
```

## Best Practices

### Security

1. **Use strong passwords** for token protection
2. **Generate tokens with sufficient length** (32+ characters)
3. **Set appropriate expiration times** for temporary data
4. **Implement proper error handling** without information leakage

### Performance

1. **Batch operations** when handling multiple tokens
2. **Use context with timeout** for production operations
3. **Implement caching** for frequently accessed tokens
4. **Monitor token creation rates** for abuse detection

### Operations

1. **Validate input parameters** before operations
2. **Handle all error cases** appropriately
3. **Use soft delete** for data recovery needs
4. **Implement cleanup** for expired tokens

## See Also

- [Core Store](core_store.md) - Main store implementation
- [Record Management](record_management.md) - Record operations
- [Encryption](encryption.md) - Encryption and decryption
- [API Reference](../api_reference.md) - Complete API documentation
- [Password Identity Management](password_identity_management.md) - Identity-based password management

## Changelog

- **v1.1.0** (2026-02-03): Added BulkRekey documentation and password identity management references.
- **v1.0.0** (2026-02-03): Initial token operations documentation
