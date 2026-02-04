---
path: api_reference.md
page-type: reference
summary: Complete API reference for VaultStore interfaces and methods including pure encryption bulk rekey.
tags: [api, reference, interfaces, methods, bulk-rekey]
created: 2026-02-03
updated: 2026-02-04
version: 1.2.0
---

# API Reference

This document provides a complete reference for all VaultStore APIs, interfaces, and methods.

## Core Interfaces

### StoreInterface

The main interface that provides all vault operations.

#### Methods

```go
// Database Management
AutoMigrate() error
EnableDebug(debug bool)
GetDbDriverName() string
GetVaultTableName() string
GetMetaTableName() string

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

// Pure Encryption Bulk Rekey
BulkRekey(ctx context.Context, oldPassword, newPassword string) (int, error)

// Vault Settings
GetVaultVersion(ctx context.Context) (string, error)
SetVaultVersion(ctx context.Context, version string) error
GetVaultSetting(ctx context.Context, key string) (string, error)
SetVaultSetting(ctx context.Context, key, value string) error
```

### RecordInterface

Interface for record data operations.

#### Methods

```go
// Data Access
Data() map[string]string
DataChanged() map[string]string

// Getters
GetCreatedAt() string
GetExpiresAt() string
GetSoftDeletedAt() string
GetID() string
GetToken() string
GetUpdatedAt() string
GetValue() string

// Setters (return self for chaining)
SetCreatedAt(createdAt string) RecordInterface
SetExpiresAt(expiresAt string) RecordInterface
SetSoftDeletedAt(softDeletedAt string) RecordInterface
SetID(id string) RecordInterface
SetToken(token string) RecordInterface
SetUpdatedAt(updatedAt string) RecordInterface
SetValue(value string) RecordInterface
```

### RecordQueryInterface

Interface for building flexible queries.

#### Methods

```go
// Validation
Validate() error
toSelectDataset(store StoreInterface) (*goqu.SelectDataset, []any, error)

// Column Selection
GetColumns() []string
SetColumns(columns []string) RecordQueryInterface
IsColumnsSet() bool

// ID Filtering
IsIDSet() bool
GetID() string
SetID(id string) RecordQueryInterface
IsIDInSet() bool
GetIDIn() []string
SetIDIn(idIn []string) RecordQueryInterface

// Token Filtering
IsTokenSet() bool
GetToken() string
SetToken(token string) RecordQueryInterface
IsTokenInSet() bool
GetTokenIn() []string
SetTokenIn(tokenIn []string) RecordQueryInterface

// Pagination
IsOffsetSet() bool
GetOffset() int
SetOffset(offset int) RecordQueryInterface
IsLimitSet() bool
GetLimit() int
SetLimit(limit int) RecordQueryInterface

// Ordering
IsOrderBySet() bool
GetOrderBy() string
SetOrderBy(orderBy string) RecordQueryInterface
IsSortOrderSet() bool
GetSortOrder() string
SetSortOrder(sortOrder string) RecordQueryInterface

// Special Options
IsCountOnlySet() bool
GetCountOnly() bool
SetCountOnly(countOnly bool) RecordQueryInterface
IsSoftDeletedIncludeSet() bool
GetSoftDeletedInclude() bool
SetSoftDeletedInclude(softDeletedInclude bool) RecordQueryInterface
```

## Factory Functions

### NewStore

Creates a new vault store instance.

```go
func NewStore(opts NewStoreOptions) (*storeImplementation, error)
```

#### Parameters

```go
type NewStoreOptions struct {
    VaultTableName     string  // Required: Table name for vault
    DB                 *sql.DB // Required: Database connection
    AutomigrateEnabled bool    // Optional: Enable auto migration
    DebugEnabled       bool    // Optional: Enable debug logging
    DbDriverName       string  // Optional: Database driver name
    CryptoConfig       *CryptoConfig // Optional: Custom Argon2id parameters
    ParallelThreshold  int     // Optional: Threshold for parallel bulk rekey (default: 10000)
}
```

#### Returns

- `*storeImplementation`: Store instance
- `error`: Error if creation fails

### NewRecordQuery

Creates a new query builder instance.

```go
func NewRecordQuery() *recordQueryImplementation
```

#### Returns

- `*recordQueryImplementation`: Query builder instance

## Token Operations

### TokenCreate

Creates a new token with automatic token generation.

```go
func (s *storeImplementation) TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (string, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `value string`: Value to encrypt and store
- `password string`: Password for encryption (optional, can be empty)
- `tokenLength int`: Length of generated token
- `options ...TokenCreateOptions`: Optional creation options

#### Returns

- `string`: Generated token
- `error`: Error if creation fails

#### TokenCreateOptions

```go
type TokenCreateOptions struct {
    ExpiresAt time.Time // Token expiration time
}
```

### TokenCreateCustom

Creates a token with a custom token value.

```go
func (s *storeImplementation) TokenCreateCustom(ctx context.Context, token string, value string, password string, options ...TokenCreateOptions) error
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Custom token value
- `value string`: Value to encrypt and store
- `password string`: Password for encryption (optional)
- `options ...TokenCreateOptions`: Optional creation options

#### Returns

- `error`: Error if creation fails

### TokenRead

Reads a value using a token and password.

```go
func (s *storeImplementation) TokenRead(ctx context.Context, token string, password string) (string, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to look up
- `password string`: Password for decryption

#### Returns

- `string`: Decrypted value
- `error`: Error if read fails

### TokenUpdate

Updates a token's value.

```go
func (s *storeImplementation) TokenUpdate(ctx context.Context, token string, value string, password string) error
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to update
- `value string`: New value to encrypt
- `password string`: Password for encryption

#### Returns

- `error`: Error if update fails

### TokenDelete

Hard deletes a token (permanent deletion).

```go
func (s *storeImplementation) TokenDelete(ctx context.Context, token string) error
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to delete

#### Returns

- `error`: Error if deletion fails

### TokenSoftDelete

Soft deletes a token (recoverable).

```go
func (s *storeImplementation) TokenSoftDelete(ctx context.Context, token string) error
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to soft delete

#### Returns

- `error`: Error if soft delete fails

### TokenExists

Checks if a token exists.

```go
func (s *storeImplementation) TokenExists(ctx context.Context, token string) (bool, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to check

#### Returns

- `bool`: True if token exists
- `error`: Error if check fails

### TokenRenew

Updates a token's expiration time.

```go
func (s *storeImplementation) TokenRenew(ctx context.Context, token string, expiresAt time.Time) error
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to renew
- `expiresAt time.Time`: New expiration time

#### Returns

- `error`: Error if renewal fails

## Record Operations

### RecordCreate

Creates a new record.

```go
func (s *storeImplementation) RecordCreate(ctx context.Context, record RecordInterface) error
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `record RecordInterface`: Record to create

#### Returns

- `error`: Error if creation fails

### RecordFindByToken

Finds a record by token.

```go
func (s *storeImplementation) RecordFindByToken(ctx context.Context, token string) (RecordInterface, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `token string`: Token to search for

#### Returns

- `RecordInterface`: Found record
- `error`: Error if not found or search fails

### RecordList

Lists records based on query criteria.

```go
func (s *storeImplementation) RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `query RecordQueryInterface`: Query criteria

#### Returns

- `[]RecordInterface`: List of matching records
- `error`: Error if query fails

### RecordCount

Counts records based on query criteria.

```go
func (s *storeImplementation) RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation
- `query RecordQueryInterface`: Query criteria

#### Returns

- `int64`: Count of matching records
- `error`: Error if count fails

## Query Examples

### Basic Query

```go
query := vaultstore.NewRecordQuery().
    SetToken("abc123").
    SetLimit(10)

records, err := vault.RecordList(context.Background(), query)
```

### Advanced Query

```go
query := vaultstore.NewRecordQuery().
    SetTokenIn([]string{"token1", "token2", "token3"}).
    SetLimit(20).
    SetOffset(0).
    SetOrderBy("created_at").
    SetSortOrder("desc").
    SetSoftDeletedInclude(false)

records, err := vault.RecordList(context.Background(), query)
```

### Count Query

```go
query := vaultstore.NewRecordQuery().
    SetSoftDeletedInclude(false)

count, err := vault.RecordCount(context.Background(), query)
```

## Error Types

### Common Errors

```go
var ErrTokenRequired = errors.New("token is required")
var ErrPasswordRequired = errors.New("password is required")
var ErrRecordNotFound = errors.New("record not found")
var ErrTokenAlreadyExists = errors.New("token already exists")
var ErrInvalidPassword = errors.New("invalid password")
var ErrDecryptionFailed = errors.New("decryption failed")
```

### Validation Errors

Validation errors are returned when input parameters are invalid:

- Empty token when required
- Invalid token format
- Missing required fields

### Database Errors

Database errors occur during database operations:

- Connection failures
- Constraint violations
- Query execution failures

### Encryption Errors

Encryption errors occur during cryptographic operations:

- Invalid password
- Corrupted encrypted data
- Encryption algorithm failures

## Pure Encryption Bulk Rekey

### BulkRekey

Changes the password for all records encrypted with a specific password using pure encryption scan-and-test approach. No password metadata is stored, providing maximum security against correlation attacks.

```go
func (s *storeImplementation) BulkRekey(ctx context.Context, oldPassword, newPassword string) (int, error)
```

#### Parameters

- `ctx context.Context`: Context for the operation (supports cancellation)
- `oldPassword string`: Current password used for encryption
- `newPassword string`: New password to use for re-encryption

#### Returns

- `int`: Number of records re-encrypted
- `error`: Error if operation fails (returns partial count with wrapped error on cancellation)

#### Behavior

**Pure Encryption Approach (No Metadata):**
1. Retrieve all records from the vault
2. For small datasets (< parallelThreshold): Use sequential processing
3. For large datasets (>= parallelThreshold): Use parallel processing with 10 workers
4. For very large datasets (> 1000 records): Use cursor-based pagination to prevent memory exhaustion
5. Attempt decryption of each record with oldPassword
6. Re-encrypt successful decryptions with newPassword
7. Update records in database

**Performance Characteristics:**
- Small datasets (< 1000 records): Sequential processing, ~1-2 seconds
- Medium datasets (1000-10000 records): Sequential processing, ~10-20 seconds  
- Large datasets (> 10000 records): Parallel processing with 10 workers
- Very large datasets (> 1000 records): Cursor-based streaming to prevent memory exhaustion

#### Example

```go
// Rekey all records using "oldpass" to use "newpass"
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

count, err := vault.BulkRekey(ctx, "oldpass", "newpass")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Printf("Partial rekey completed: %d records\n", count)
    } else {
        log.Fatal(err)
    }
}
fmt.Printf("Re-encrypted %d records\n", count)
```

### ParallelThreshold Configuration

Configure the threshold for switching between sequential and parallel processing:

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:    "vault",
    DB:                db,
    ParallelThreshold: 5000, // Use parallel for datasets >= 5000 records
})
```

- Default: 10000 records
- Set to 0 to use default
- Lower values increase parallelism overhead
- Higher values may miss parallelization benefits

## Vault Settings

### GetVaultVersion

Retrieves the current vault version from metadata.

```go
func (s *storeImplementation) GetVaultVersion(ctx context.Context) (string, error)
```

Returns the vault version string (e.g., "0.26.0") or empty string if not set.

### SetVaultVersion

Sets the vault version in metadata.

```go
func (s *storeImplementation) SetVaultVersion(ctx context.Context, version string) error
```

Used to track vault state for migration purposes.

### GetVaultSetting

Retrieves a custom vault setting by key.

```go
func (s *storeImplementation) GetVaultSetting(ctx context.Context, key string) (string, error)
```

### SetVaultSetting

Sets a custom vault setting.

```go
func (s *storeImplementation) SetVaultSetting(ctx context.Context, key, value string) error
```

Allows storing arbitrary key-value pairs in vault metadata.

## Changelog

- **v1.2.0** (2026-02-04): Removed identity-based password management methods (MigrateRecordLinks, IsVaultMigrated, MarkVaultMigrated). Updated BulkRekey documentation for pure encryption approach. Added ParallelThreshold configuration option.
- **v1.1.0** (2026-02-03): Added documentation for identity-based password management (BulkRekey, MigrateRecordLinks) and vault settings methods.
- **v1.0.0** (2026-02-03): Initial API reference documentation

## See Also

- [Getting Started](getting_started.md) - Usage examples
- [Architecture](architecture.md) - System design
- [Query Interface](modules/query_interface.md) - Detailed query documentation
- [Token Operations](modules/token_operations.md) - Token-specific operations
