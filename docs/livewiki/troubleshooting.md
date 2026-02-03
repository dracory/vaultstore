---
path: troubleshooting.md
page-type: tutorial
summary: Common issues, solutions, and debugging tips for VaultStore.
tags: [troubleshooting, debugging, issues, solutions]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
---

# Troubleshooting

This document covers common issues, their solutions, and debugging tips for VaultStore.

## Common Issues

### Database Connection Issues

#### Issue: Database connection fails

**Symptoms:**
```
vault store: DB is required
database connection failed
```

**Causes:**
- Database server not running
- Incorrect connection parameters
- Network connectivity issues
- Insufficient permissions

**Solutions:**

1. **Check Database Server:**
```bash
# SQLite
ls -la ./vault.db

# PostgreSQL
pg_isready -h localhost -p 5432

# MySQL
mysqladmin ping -h localhost
```

2. **Verify Connection String:**
```go
// Test connection separately
db, err := sql.Open("sqlite", "./vault.db")
if err != nil {
    log.Fatal("Connection error:", err)
}

// Test ping
err = db.Ping()
if err != nil {
    log.Fatal("Ping failed:", err)
}
```

3. **Check Permissions:**
```bash
# File permissions for SQLite
chmod 666 vault.db

# Database permissions for PostgreSQL
psql -c "\l"  # List databases
```

#### Issue: SQLite database locked

**Symptoms:**
```
database is locked
database table is locked
```

**Causes:**
- Multiple connections to SQLite file
- Uncommitted transactions
- Concurrent access issues

**Solutions:**

1. **Use WAL Mode:**
```go
db, err := sql.Open("sqlite", "./vault.db?cache=shared&mode=rwc")
if err != nil {
    log.Fatal(err)
}

// Enable WAL mode
_, err = db.Exec("PRAGMA journal_mode=WAL")
if err != nil {
    log.Fatal(err)
}
```

2. **Limit Connections:**
```go
db.SetMaxOpenConns(1)  // SQLite recommends single connection
db.SetMaxIdleConns(1)
```

3. **Close Connections Properly:**
```go
defer db.Close()
defer vault.Close()  // If available
```

### Token Operations Issues

#### Issue: Token not found

**Symptoms:**
```
record not found
ErrRecordNotFound
```

**Causes:**
- Token was soft deleted
- Token expired
- Incorrect token value
- Case sensitivity issues

**Solutions:**

1. **Check Token Existence:**
```go
exists, err := vault.TokenExists(ctx, token)
if err != nil {
    log.Fatal("Check failed:", err)
}
if !exists {
    log.Println("Token does not exist")
}
```

2. **Check Soft Delete Status:**
```go
record, err := vault.RecordFindByToken(ctx, token)
if err == nil {
    if record.GetSoftDeletedAt() != "" {
        log.Println("Token was soft deleted at:", record.GetSoftDeletedAt())
    }
}
```

3. **Check Expiration:**
```go
record, err := vault.RecordFindByToken(ctx, token)
if err == nil {
    expiresAt := record.GetExpiresAt()
    if expiresAt != "" {
        expiryTime, _ := time.Parse(time.RFC3339, expiresAt)
        if time.Now().After(expiryTime) {
            log.Println("Token expired at:", expiryTime)
        }
    }
}
```

#### Issue: Invalid password error

**Symptoms:**
```
invalid password
ErrInvalidPassword
decryption failed
```

**Causes:**
- Incorrect password provided
- Password encoding issues
- Encryption/decryption mismatch
- Corrupted encrypted data

**Solutions:**

1. **Verify Password:**
```go
// Use the exact password used during creation
originalPassword := "mySecretPassword123"
value, err := vault.TokenRead(ctx, token, originalPassword)
if err != nil {
    log.Printf("Password verification failed: %v", err)
}
```

2. **Check Password Encoding:**
```go
// Ensure no extra whitespace
password := strings.TrimSpace(password)
password = strings.Trim(password, "\n\r")
```

3. **Test Encryption Separately:**
```go
// Test encryption/decryption
testValue := "test"
testPassword := "password"
encrypted, err := encrypt(testValue, testPassword)
if err != nil {
    log.Fatal("Encryption failed:", err)
}

decrypted, err := decrypt(encrypted, testPassword)
if err != nil {
    log.Fatal("Decryption failed:", err)
}

if decrypted != testValue {
    log.Fatal("Encryption/decryption mismatch")
}
```

### Migration Issues

#### Issue: Auto-migration fails

**Symptoms:**
```
table creation failed
column already exists
migration error
```

**Causes:**
- Insufficient database permissions
- Existing table with different schema
- Database connection issues
- Constraint violations

**Solutions:**

1. **Check Permissions:**
```go
// Test table creation permission
_, err := db.Exec("CREATE TABLE test_table (id TEXT)")
if err != nil {
    log.Fatal("Table creation permission denied:", err)
}
db.Exec("DROP TABLE test_table")
```

2. **Check Existing Schema:**
```sql
-- SQLite
.schema vault

-- PostgreSQL
\d vault

-- MySQL
DESCRIBE vault;
```

3. **Manual Migration:**
```sql
-- Create table manually
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
```

### Performance Issues

#### Issue: Slow query performance

**Symptoms:**
- Queries take long time to execute
- High CPU usage during queries
- Memory usage increases over time

**Causes:**
- Missing database indexes
- Large result sets
- Inefficient queries
- Connection pool exhaustion

**Solutions:**

1. **Check Indexes:**
```sql
-- Verify indexes exist
EXPLAIN QUERY PLAN SELECT * FROM vault WHERE token = 'abc123';

-- Create missing indexes
CREATE INDEX idx_vault_token ON vault(token);
CREATE INDEX idx_vault_created_at ON vault(created_at);
```

2. **Optimize Queries:**
```go
// Use specific columns instead of SELECT *
query := vaultstore.NewRecordQuery().
    SetColumns([]string{"id", "token", "created_at"}).
    SetLimit(100)  // Limit result size
```

3. **Monitor Connection Pool:**
```go
// Check connection pool stats
stats := db.Stats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("In use: %d\n", stats.InUse)
fmt.Printf("Idle: %d\n", stats.Idle)
```

### Memory Issues

#### Issue: Memory leaks

**Symptoms:**
- Memory usage increases continuously
- Out of memory errors
- Garbage collection pressure

**Causes:**
- Unclosed database connections
- Large objects retained in memory
- Goroutine leaks

**Solutions:**

1. **Profile Memory:**
```bash
# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

2. **Check Resource Cleanup:**
```go
defer db.Close()
defer ctx.Cancel()

// Check for goroutine leaks
runtime.GC()
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Alloc = %v\n", m.Alloc)
```

3. **Monitor Goroutines:**
```go
// Check goroutine count
go func() {
    for {
        time.Sleep(10 * time.Second)
        fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
    }
}()
```

## Debugging Techniques

### Enable Debug Mode

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    DebugEnabled:   true,  // Enable debug logging
})
```

### Logging

Add comprehensive logging:

```go
func debugLog(format string, args ...interface{}) {
    if debugEnabled {
        log.Printf(format, args...)
    }
}

// Usage
debugLog("Creating token with length: %d", tokenLength)
debugLog("Database query: %s", query)
debugLog("Encryption time: %v", time.Since(start))
```

### Database Query Logging

```go
// Wrap database driver for logging
type loggingDriver struct {
    *sql.Driver
}

func (d *loggingDriver) Open(name string) (driver.Conn, error) {
    conn, err := d.Driver.Open(name)
    if err != nil {
        return nil, err
    }
    return &loggingConn{Conn: conn}, nil
}

type loggingConn struct {
    driver.Conn
}

func (c *loggingConn) Query(query string, args []driver.Value) (driver.Rows, error) {
    log.Printf("Query: %s, Args: %v", query, args)
    return c.Conn.Query(query, args)
}
```

### Error Analysis

Enhanced error reporting:

```go
type VaultError struct {
    Operation string
    Token     string
    Cause     error
    Timestamp time.Time
}

func (e *VaultError) Error() string {
    return fmt.Sprintf("Vault operation '%s' failed for token '%s' at %v: %v",
        e.Operation, e.Token, e.Timestamp, e.Cause)
}

// Usage
if err != nil {
    return &VaultError{
        Operation: "TokenRead",
        Token:     token,
        Cause:     err,
        Timestamp: time.Now(),
    }
}
```

## Testing Scenarios

### Database Connection Test

```go
func TestDatabaseConnection(t *testing.T) {
    db, err := sql.Open("sqlite", ":memory:")
    require.NoError(t, err)
    
    err = db.Ping()
    require.NoError(t, err)
    
    defer db.Close()
}
```

### Encryption Test

```go
func TestEncryptionDecryption(t *testing.T) {
    value := "test value"
    password := "test password"
    
    encrypted, err := encrypt(value, password)
    require.NoError(t, err)
    assert.NotEmpty(t, encrypted)
    
    decrypted, err := decrypt(encrypted, password)
    require.NoError(t, err)
    assert.Equal(t, value, decrypted)
}
```

### Token Lifecycle Test

```go
func TestTokenLifecycle(t *testing.T) {
    vault := createTestStore(t)
    ctx := context.Background()
    
    // Create
    token, err := vault.TokenCreate(ctx, "value", "password", 32)
    require.NoError(t, err)
    
    // Exists
    exists, err := vault.TokenExists(ctx, token)
    require.NoError(t, err)
    assert.True(t, exists)
    
    // Read
    value, err := vault.TokenRead(ctx, token, "password")
    require.NoError(t, err)
    assert.Equal(t, "value", value)
    
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

## Performance Monitoring

### Metrics Collection

```go
type Metrics struct {
    TokenCreations   int64
    TokenReads       int64
    TokenUpdates     int64
    TokenDeletes     int64
    QueryTime        time.Duration
    EncryptionTime   time.Duration
    DatabaseTime     time.Duration
}

func (m *Metrics) RecordOperation(op string, duration time.Duration) {
    switch op {
    case "TokenCreate":
        atomic.AddInt64(&m.TokenCreations, 1)
    case "TokenRead":
        atomic.AddInt64(&m.TokenReads, 1)
    // ... other operations
    }
}
```

### Benchmark Testing

```go
func BenchmarkTokenOperations(b *testing.B) {
    vault := createTestStore(&testing.T{})
    ctx := context.Background()
    
    b.Run("Create", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _, err := vault.TokenCreate(ctx, "value", "password", 32)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("Read", func(b *testing.B) {
        // Create test token
        token, _ := vault.TokenCreate(ctx, "value", "password", 32)
        
        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _, err := vault.TokenRead(ctx, token, "password")
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

## Getting Help

### Community Resources

- **GitHub Issues**: [Report bugs and request features](https://github.com/dracory/vaultstore/issues)
- **GitHub Discussions**: [Ask questions and share ideas](https://github.com/dracory/vaultstore/discussions)
- **Documentation**: [Complete documentation](https://github.com/dracory/vaultstore/docs)

### Reporting Issues

When reporting issues, include:

1. **Go Version**: `go version`
2. **VaultStore Version**: Check `go.mod`
3. **Database Type and Version**
4. **Operating System**
5. **Minimal Reproducible Example**
6. **Error Messages and Stack Traces**
7. **Expected vs Actual Behavior**

### Issue Template

```markdown
## Bug Report
**VaultStore Version**: vX.Y.Z
**Go Version**: 1.25.X
**Database**: SQLite X.X.X
**OS**: Linux/macOS/Windows

### Description
Brief description of the issue.

### Steps to Reproduce
1. Create store with these options...
2. Call this method...
3. Observe this error...

### Expected Behavior
What should happen.

### Actual Behavior
What actually happens.

### Error Message
```
Paste error message here
```

### Additional Context
Any other relevant information.
```

## See Also

- [Getting Started](getting_started.md) - Setup and usage
- [API Reference](api_reference.md) - Complete API documentation
- [Development](development.md) - Development workflow
- [Configuration](configuration.md) - Configuration options
