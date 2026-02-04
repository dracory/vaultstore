---
path: cheatsheet.md
page-type: reference
summary: Quick reference for common VaultStore operations and patterns.
tags: [cheatsheet, reference, quick-start, patterns]
created: 2026-02-03
updated: 2026-02-04
version: 1.1.0
---

# VaultStore Cheatsheet

Quick reference for common VaultStore operations and patterns.

## Setup

### Basic Store Creation

```go
import (
    "database/sql"
    "context"
    "log"
    
    "github.com/dracory/vaultstore"
    _ "github.com/glebarez/sqlite"
)

// Create database connection
db, err := sql.Open("sqlite", "./vault.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create vault store
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "vault",
    DB:                 db,
    AutomigrateEnabled: true,
    DebugEnabled:       false,
})
if err != nil {
    log.Fatal(err)
}
```

### Production Configuration

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "production_vault",
    DB:                 db,
    AutomigrateEnabled: false,  // Manual migration
    DebugEnabled:       false,  // No debug logs
    DbDriverName:       "postgres",
})
```

## Token Operations

### Create Token

```go
// Basic token creation
token, err := vault.TokenCreate(context.Background(), "my_secret", "", 32)

// With password protection
token, err := vault.TokenCreate(context.Background(), "my_secret", "password123", 32)

// With expiration
token, err := vault.TokenCreate(context.Background(), "my_secret", "", 32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Now().Add(24 * time.Hour),
    })

// Custom token
err := vault.TokenCreateCustom(context.Background(), "my_token", "my_secret", "", 32)
```

### Read Token

```go
// Read with password
value, err := vault.TokenRead(context.Background(), token, "password123")

// Read without password
value, err := vault.TokenRead(context.Background(), token, "")

// Batch read
values, err := vault.TokensRead(context.Background(), 
    []string{"token1", "token2", "token3"}, "password")
```

### Update Token

```go
// Update value
err := vault.TokenUpdate(context.Background(), token, "new_secret", "password")

// Renew expiration
err := vault.TokenRenew(context.Background(), token, 
    time.Now().Add(48 * time.Hour))
```

### Delete Token

```go
// Soft delete (recoverable)
err := vault.TokenSoftDelete(context.Background(), token)

// Hard delete (permanent)
err := vault.TokenDelete(context.Background(), token)
```

### Check Token

```go
// Check if token exists
exists, err := vault.TokenExists(context.Background(), token)
if err != nil {
    log.Fatal(err)
}
if !exists {
    log.Println("Token not found")
}
```

## Record Operations

### Create Record

```go
// Create new record
record := vaultstore.NewRecord().
    SetToken("my_token").
    SetValue("encrypted_value").
    SetCreatedAt(time.Now().Format(time.RFC3339))

err := vault.RecordCreate(context.Background(), record)
```

### Find Records

```go
// Find by token
record, err := vault.RecordFindByToken(context.Background(), "my_token")

// Find by ID
record, err := vault.RecordFindByID(context.Background(), "record_id")

// List records
query := vaultstore.RecordQuery().SetLimit(10)
records, err := vault.RecordList(context.Background(), query)
```

### Update Record

```go
// Find existing record
record, err := vault.RecordFindByToken(context.Background(), "my_token")
if err != nil {
    log.Fatal(err)
}

// Update fields
record.SetValue("new_value").
    SetData(map[string]string{
        "category": "credentials",
        "owner":    "user123",
    })

// Save changes
err = vault.RecordUpdate(context.Background(), record)
```

### Delete Record

```go
// Find record
record, err := vault.RecordFindByToken(context.Background(), "my_token")
if err != nil {
    log.Fatal(err)
}

// Soft delete
err = vault.RecordSoftDelete(context.Background(), record)

// Hard delete by token
err = vault.RecordDeleteByToken(context.Background(), "my_token")
```

## Query Operations

### Basic Queries

```go
// Find by token
query := vaultstore.RecordQuery().SetToken("my_token")

// Find multiple tokens
query := vaultstore.RecordQuery().
    SetTokenIn([]string{"token1", "token2", "token3"})

// Find by ID
query := vaultstore.RecordQuery().SetID("record_id")
```

### Filtering

```go
// Exclude soft deleted records
query := vaultstore.RecordQuery().
    SetSoftDeletedInclude(false)

// Include soft deleted records
query := vaultstore.RecordQuery().
    SetSoftDeletedInclude(true)

// Select specific columns
query := vaultstore.RecordQuery().
    SetColumns([]string{"id", "token", "created_at"})
```

### Sorting and Pagination

```go
// Sort by creation date (newest first)
query := vaultstore.RecordQuery().
    SetOrderBy("created_at").
    SetSortOrder("desc")

// Paginate results
query := vaultstore.RecordQuery().
    SetLimit(25).
    SetOffset(0)  // Page 1

query = vaultstore.RecordQuery().
    SetLimit(25).
    SetOffset(25) // Page 2
```

### Counting

```go
// Count all records
query := vaultstore.RecordQuery()
count, err := vault.RecordCount(context.Background(), query)

// Count filtered records
query = vaultstore.RecordQuery().
    SetSoftDeletedInclude(false)
count, err = vault.RecordCount(context.Background(), query)
```

## Common Patterns

### Session Management

```go
// Create session token
func createSession(vault StoreInterface, userID string, sessionData string) (string, error) {
    token, err := vault.TokenCreate(context.Background(),
        sessionData,
        "", // No password for sessions
        32,
        vaultstore.TokenCreateOptions{
            ExpiresAt: time.Now().Add(24 * time.Hour),
        })
    return token, err
}

// Validate session
func validateSession(vault StoreInterface, token string) (string, error) {
    exists, err := vault.TokenExists(context.Background(), token)
    if err != nil || !exists {
        return "", errors.New("invalid session")
    }
    
    return vault.TokenRead(context.Background(), token, "")
}
```

### API Key Storage

```go
// Store API key
func storeAPIKey(vault StoreInterface, keyName, apiKey string) (string, error) {
    return vault.TokenCreate(context.Background(),
        apiKey,
        "master_password", // Use application master password
        32)
}

// Retrieve API key
func getAPIKey(vault StoreInterface, token string) (string, error) {
    return vault.TokenRead(context.Background(), token, "master_password")
}
```

### Temporary Credentials

```go
// Create temporary access
func createTempAccess(vault StoreInterface, data string, duration time.Duration) (string, error) {
    return vault.TokenCreate(context.Background(),
        data,
        "", // No password for temporary access
        32,
        vaultstore.TokenCreateOptions{
            ExpiresAt: time.Now().Add(duration),
        })
}

// Cleanup expired tokens
func cleanupExpired(vault StoreInterface) error {
    count, err := vault.TokensExpiredSoftDelete(context.Background())
    if err != nil {
        return err
    }
    log.Printf("Cleaned up %d expired tokens", count)
    return nil
}
```

## Error Handling

### Common Error Patterns

```go
// Safe token read
func safeTokenRead(vault StoreInterface, token, password string) (string, error) {
    // Check existence first
    exists, err := vault.TokenExists(context.Background(), token)
    if err != nil {
        return "", fmt.Errorf("token check failed: %w", err)
    }
    if !exists {
        return "", vaultstore.ErrRecordNotFound
    }
    
    // Attempt read
    value, err := vault.TokenRead(context.Background(), token, password)
    if err != nil {
        if errors.Is(err, vaultstore.ErrInvalidPassword) {
            return "", vaultstore.ErrInvalidPassword
        }
        return "", fmt.Errorf("token read failed: %w", err)
    }
    
    return value, nil
}
```

### Error Types

```go
// Common errors
vaultstore.ErrRecordNotFound     // Token/record doesn't exist
vaultstore.ErrInvalidPassword    // Wrong password
vaultstore.ErrDecryptionFailed   // Encryption/decryption error
vaultstore.ErrTokenAlreadyExists // Token uniqueness violation
```

## Database Setup

### SQLite

```go
db, err := sql.Open("sqlite", "./vault.db")
db.SetMaxOpenConns(1)  // SQLite recommendation
db.SetMaxIdleConns(1)
```

### PostgreSQL

```go
db, err := sql.Open("postgres", "postgres://user:pass@localhost/db?sslmode=disable")
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### MySQL

```go
db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## Security Best Practices

### Token Security

```go
// Use sufficient token length
token, err := vault.TokenCreate(ctx, value, password, 32) // Minimum 32 chars

// Set expiration for temporary data
token, err := vault.TokenCreate(ctx, value, password, 32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Now().Add(24 * time.Hour),
    })
```

### Password Security

```go
// Use strong passwords
password := "my_strong_password_123!"

// Never store passwords in logs
log.Printf("Token created: %s", token) // OK
log.Printf("Password: %s", password)  // BAD - Don't do this
```

### Error Handling

```go
// Don't leak information in errors
if err != nil {
    log.Printf("Operation failed: %v", err) // OK for debugging
    return errors.New("internal error")    // OK for user response
}
```

## Performance Tips

### Query Optimization

```go
// Select only needed columns
query := vaultstore.RecordQuery().
    SetColumns([]string{"id", "token", "created_at"})

// Limit results
query = vaultstore.RecordQuery().
    SetLimit(100)

// Use indexes effectively
query = vaultstore.RecordQuery().
    SetToken("specific_token") // Uses token index
```

### Batch Operations

```go
// Batch token creation
tokens := make([]string, 0, len(values))
for _, value := range values {
    token, err := vault.TokenCreate(ctx, value, "", 32)
    if err != nil {
        return tokens, err
    }
    tokens = append(tokens, token)
}
```

## Testing

### Test Setup

```go
func createTestStore(t *testing.T) StoreInterface {
    db, err := sql.Open("sqlite", ":memory:")
    require.NoError(t, err)
    
    vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
        VaultTableName:     "test_vault",
        DB:                 db,
        AutomigrateEnabled: true,
    })
    require.NoError(t, err)
    return vault
}
```

### Common Test Patterns

```go
func TestTokenLifecycle(t *testing.T) {
    vault := createTestStore(t)
    ctx := context.Background()
    
    // Create
    token, err := vault.TokenCreate(ctx, "value", "password", 32)
    require.NoError(t, err)
    
    // Read
    value, err := vault.TokenRead(ctx, token, "password")
    require.NoError(t, err)
    assert.Equal(t, "value", value)
    
    // Delete
    err = vault.TokenDelete(ctx, token)
    require.NoError(t, err)
}
```

## Debugging

### Enable Debug Mode

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    DebugEnabled:   true, // Enable debug logging
})
```

### Common Debugging Commands

```go
// Check database connection
err = db.Ping()
if err != nil {
    log.Fatal("Database connection failed:", err)
}

// Check store configuration
log.Printf("Table: %s", vault.GetVaultTableName())
log.Printf("Driver: %s", vault.GetDbDriverName())

// Test basic operation
token, err := vault.TokenCreate(ctx, "test", "", 32)
if err != nil {
    log.Printf("Token creation failed: %v", err)
}
```

## Migration

### Manual Schema

```sql
CREATE TABLE vault (
    id              TEXT PRIMARY KEY,
    token           TEXT UNIQUE NOT NULL,
    value           TEXT NOT NULL,
    password_hash   TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL,
    expires_at      TEXT,
    soft_deleted_at TEXT,
    data            TEXT
);

CREATE INDEX idx_vault_token ON vault(token);
CREATE INDEX idx_vault_created_at ON vault(created_at);
CREATE INDEX idx_vault_expires_at ON vault(expires_at);
CREATE INDEX idx_vault_soft_deleted_at ON vault(soft_deleted_at);
```

## See Also

- [Getting Started](getting_started.md) - Complete setup guide
- [API Reference](api_reference.md) - Full API documentation
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Development](development.md) - Development workflow and testing

## Changelog

- **v1.1.0** (2026-02-04): Updated all query examples to use `RecordQuery()` instead of `NewRecordQuery()`
- **v1.0.0** (2026-02-03): Initial creation
