---
path: modules/core_store.md
page-type: module
summary: Core store implementation and main interface documentation including CryptoConfig and Password Identity options.
tags: [module, core, store, interface, cryptoconfig, identity]
created: 2026-02-03
updated: 2026-02-03
version: 1.1.0
---

# Core Store Module

The core store module provides the main implementation of VaultStore's functionality through the `StoreInterface`. This module serves as the central hub for all vault operations.

## Overview

The core store module is responsible for:
- Database connection management
- Record lifecycle management
- Token operations orchestration
- Query execution
- Database migrations

## Main Interface

### StoreInterface

The primary interface that defines all vault operations:

```go
type StoreInterface interface {
    // Database Management
    AutoMigrate() error
    EnableDebug(debug bool)
    GetDbDriverName() string
    GetVaultTableName() string

    // Record Operations
    RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error)
    RecordCreate(ctx context.Context, record RecordInterface) error
    RecordDeleteByID(ctx context.Context, recordID string) error
    RecordDeleteByToken(ctx context.Context, token string) error
    RecordFindByID(ctx context.Context, recordID string) (RecordInterface, error)
    RecordFindByToken(ctx context.Context, token string) (RecordInterface, error)
    RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error)
    RecordSoftDelete(ctx context.Context, record RecordInterface) error
    RecordSoftDeleteByID(ctx context.Context, recordID string) error
    RecordSoftDeleteByToken(ctx context.Context, token string) error
    RecordUpdate(ctx context.Context, record RecordInterface) error

    // Token Operations
    TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (string, error)
    TokenCreateCustom(ctx context.Context, token string, value string, password string, options ...TokenCreateOptions) error
    TokenDelete(ctx context.Context, token string) error
    TokenExists(ctx context.Context, token string) (bool, error)
    TokenRead(ctx context.Context, token string, password string) (string, error)
    TokenRenew(ctx context.Context, token string, expiresAt time.Time) error
    TokensExpiredSoftDelete(ctx context.Context) (count int64, err error)
    TokensExpiredDelete(ctx context.Context) (count int64, err error)
    TokenSoftDelete(ctx context.Context, token string) error
    TokenUpdate(ctx context.Context, token string, value string, password string) error
    TokensRead(ctx context.Context, tokens []string, password string) (map[string]string, error)
}
```

## Implementation Structure

### storeImplementation

The concrete implementation of the store interface:

```go
type storeImplementation struct {
    vaultTableName     string
    automigrateEnabled bool
    db                 *sql.DB
    gormDB             *gorm.DB
    dbDriverName       string
    debugEnabled       bool
}
```

### Key Components

#### Database Management

- **Connection Handling**: Manages both `*sql.DB` and `*gorm.DB` connections
- **Driver Detection**: Auto-detects database driver type
- **Migration Support**: Automatic table creation and schema updates

#### Record Management

- **CRUD Operations**: Complete create, read, update, delete functionality
- **Soft Delete**: Logical deletion with recovery capability
- **Query Support**: Flexible querying with multiple filters

#### Token Operations

- **Token Generation**: Cryptographically secure token creation
- **Encryption Integration**: Seamless encryption/decryption handling
- **Lifecycle Management**: Complete token lifecycle from creation to deletion

## Factory Function

### NewStore

Creates a new store instance with configuration:

```go
func NewStore(opts NewStoreOptions) (*storeImplementation, error)
```

#### Configuration Options

```go
type NewStoreOptions struct {
    VaultTableName          string
    VaultMetaTableName      string
    DB                      *sql.DB
    DbDriverName            string
    AutomigrateEnabled      bool
    DebugEnabled            bool
    CryptoConfig            *CryptoConfig
    PasswordIdentityEnabled bool
}
```

#### Validation

The factory function validates:
- `VaultTableName` is required and non-empty
- `DB` is required and not nil
- Database connection is valid

## Database Operations

### Auto Migration

Automatic database schema management:

```go
func (s *storeImplementation) AutoMigrate() error
```

**Features:**
- Creates vault table if it doesn't exist
- Adds missing columns
- Creates necessary indexes
- Validates table structure

**Supported Databases:**
- SQLite
- PostgreSQL
- MySQL
- Other GORM-supported databases

### Connection Management

The store manages two database connections:

1. **SQL Connection** (`*sql.DB`): Direct SQL operations
2. **GORM Connection** (`*gorm.DB`): ORM operations

```go
// Get database driver information
func (s *storeImplementation) GetDbDriverName() string

// Get vault table name
func (s *storeImplementation) GetVaultTableName() string
```

## Debug Support

### Debug Mode

Enable detailed logging for debugging:

```go
func (s *storeImplementation) EnableDebug(debug bool)
```

**Debug Information:**
- SQL query logging
- Error stack traces
- Performance metrics
- Encryption operations

## Usage Examples

### Basic Setup

```go
// Create store with auto-migration
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "app_vault",
    DB:                 db,
    AutomigrateEnabled: true,
    DebugEnabled:       false,
})
if err != nil {
    log.Fatal(err)
}
```

### Advanced Configuration

```go
// Production configuration
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "production_vault",
    DB:                 db,
    AutomigrateEnabled: false,  // Manual migration in production
    DebugEnabled:       false,  // No debug in production
    DbDriverName:       "postgres",
})
if err != nil {
    log.Fatal(err)
}

// Enable debug for development
vault.EnableDebug(true)
```

## Error Handling

### Common Errors

- **ValidationError**: Invalid configuration parameters
- **DatabaseError**: Database connection or operation failures
- **MigrationError**: Schema migration failures
- **NotFoundError**: Record or token not found

### Error Patterns

```go
// Check for specific errors
if errors.Is(err, vaultstore.ErrRecordNotFound) {
    // Handle not found case
}

// Check error types
var dbErr *database.Error
if errors.As(err, &dbErr) {
    // Handle database-specific error
}
```

## Performance Considerations

### Connection Pooling

The store leverages database connection pooling:

```go
// Configure connection pool (example for PostgreSQL)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Query Optimization

- Uses prepared statements for repeated queries
- Implements proper database indexing
- Supports query result limiting and pagination

### Memory Management

- Efficient memory usage for large result sets
- Proper resource cleanup and connection management
- Minimal object allocation during operations

## Dependencies

### Internal Dependencies

- **Record Module**: For record data structures
- **Token Module**: For token operations
- **Encryption Module**: For data encryption/decryption
- **Query Module**: For query building and execution

### External Dependencies

- **GORM**: ORM for database operations
- **goqu**: Query builder for complex queries
- **Database Drivers**: SQLite, PostgreSQL, MySQL drivers

## Testing

### Unit Tests

Comprehensive test coverage for:
- Store creation and configuration
- Database operations
- Error handling
- Migration functionality

### Integration Tests

Database-specific integration tests:
- SQLite integration
- PostgreSQL integration
- MySQL integration

### Test Utilities

```go
// Helper for creating test store
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
```

## Best Practices

### Configuration

1. **Use descriptive table names** for multi-tenant applications
2. **Enable debug mode** only in development environments
3. **Configure connection pooling** appropriately for your database
4. **Use manual migration** in production environments

### Error Handling

1. **Always check return values** from store operations
2. **Handle specific error types** appropriately
3. **Log errors** with sufficient context
4. **Implement retry logic** for transient failures

### Performance

1. **Use context with timeout** for production operations
2. **Limit query results** to prevent memory issues
3. **Monitor connection pool** metrics
4. **Profile operations** for performance bottlenecks

## See Also

- [Store Interface](../api_reference.md#storeinterface) - Complete API reference
- [Record Management](record_management.md) - Record operations
- [Token Operations](token_operations.md) - Token-specific operations
- [Configuration](../configuration.md) - Configuration options
- [Password Identity Management](password_identity_management.md) - Identity-based password management

## Changelog

- **v1.1.0** (2026-02-03): Updated NewStoreOptions documentation with CryptoConfig, PasswordIdentityEnabled, and VaultMetaTableName fields.
- **v1.0.0** (2026-02-03): Initial core store module documentation
