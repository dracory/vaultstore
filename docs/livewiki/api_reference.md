---
path: api_reference.md
page-type: reference
summary: Complete API reference for VaultStore interfaces and methods.
tags: [api, reference, interfaces, methods]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
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

## See Also

- [Getting Started](getting_started.md) - Usage examples
- [Architecture](architecture.md) - System design
- [Query Interface](modules/query_interface.md) - Detailed query documentation
- [Token Operations](modules/token_operations.md) - Token-specific operations
