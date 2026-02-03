---
path: modules/record_management.md
page-type: module
summary: Record CRUD operations and data management functionality.
tags: [module, record, crud, data-management]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
---

# Record Management Module

The record management module handles all CRUD operations for vault records, providing the fundamental data storage and retrieval capabilities of VaultStore.

## Overview

The record management module is responsible for:
- Record creation and validation
- Record retrieval and querying
- Record updates and modification tracking
- Record deletion (both soft and hard delete)
- Data integrity and validation

## Core Interface

### RecordInterface

Defines the contract for record data operations:

```go
type RecordInterface interface {
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
}
```

## Record Structure

### Data Fields

Each record contains the following core fields:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique record identifier (UUID) |
| `token` | string | Access token for the record |
| `value` | string | Encrypted stored value |
| `created_at` | string | Record creation timestamp |
| `updated_at` | string | Last modification timestamp |
| `expires_at` | string | Optional expiration timestamp |
| `soft_deleted_at` | string | Soft deletion timestamp |
| `data` | string | Additional JSON metadata |

### Metadata Support

Records support additional metadata through the `Data()` method:

```go
// Get record metadata
metadata := record.Data()
// Returns: map[string]string

// Get changed fields for updates
changed := record.DataChanged()
// Returns: map[string]string of modified fields
```

## Implementation

### recordImplementation

The concrete implementation of the record interface:

```go
type recordImplementation struct {
    id              string
    token           string
    value           string
    createdAt       string
    updatedAt       string
    expiresAt       string
    softDeletedAt   string
    data            map[string]string
    dataChanged     map[string]string
}
```

### Key Features

#### Immutable ID and Token

- Record ID and token cannot be changed after creation
- Ensures data integrity and prevents token conflicts

#### Change Tracking

- `DataChanged()` tracks modifications for efficient updates
- Only modified fields are updated in the database
- Reduces database load and improves performance

#### Flexible Metadata

- JSON-based metadata storage for extensibility
- Arbitrary key-value pairs for application-specific data
- Change tracking for metadata updates

## CRUD Operations

### Create

```go
func (s *storeImplementation) RecordCreate(ctx context.Context, record RecordInterface) error
```

**Process:**
1. Validate record data
2. Generate unique ID if not provided
3. Set creation timestamp
4. Encrypt value if needed
5. Insert into database

**Validation:**
- Token must be unique
- Required fields must be present
- Data format validation

### Read

#### Find by ID

```go
func (s *storeImplementation) RecordFindByID(ctx context.Context, recordID string) (RecordInterface, error)
```

#### Find by Token

```go
func (s *storeImplementation) RecordFindByToken(ctx context.Context, token string) (RecordInterface, error)
```

#### List with Query

```go
func (s *storeImplementation) RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error)
```

**Features:**
- Flexible filtering and sorting
- Pagination support
- Soft delete filtering
- Column selection optimization

### Update

```go
func (s *storeImplementation) RecordUpdate(ctx context.Context, record RecordInterface) error
```

**Process:**
1. Track changed fields using `DataChanged()`
2. Update only modified columns
3. Update timestamp
4. Maintain data integrity

**Optimizations:**
- Partial updates for performance
- Change tracking to minimize database writes
- Validation before updates

### Delete

#### Soft Delete

```go
func (s *storeImplementation) RecordSoftDelete(ctx context.Context, record RecordInterface) error
func (s *storeImplementation) RecordSoftDeleteByID(ctx context.Context, recordID string) error
func (s *storeImplementation) RecordSoftDeleteByToken(ctx context.Context, token string) error
```

**Features:**
- Logical deletion with recovery capability
- Timestamp tracking for deletion time
- Exclusion from normal queries

#### Hard Delete

```go
func (s *storeImplementation) RecordDeleteByID(ctx context.Context, recordID string) error
func (s *storeImplementation) RecordDeleteByToken(ctx context.Context, token string) error
```

**Features:**
- Permanent data removal
- Physical deletion from database
- Compliance with data retention policies

## Query Operations

### Record Count

```go
func (s *storeImplementation) RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error)
```

**Features:**
- Efficient counting with database optimization
- Support for all query filters
- Pagination metadata

### Query Filtering

Records can be filtered by:

- **ID**: Exact match or list of IDs
- **Token**: Exact match or list of tokens
- **Creation Date**: Date range filtering
- **Expiration**: Expired vs non-expired
- **Soft Delete**: Include or exclude soft deleted records
- **Custom Metadata**: JSON field queries

## Usage Examples

### Creating a Record

```go
// Create new record
record := vaultstore.NewRecord().
    SetToken("my_token").
    SetValue("encrypted_value").
    SetCreatedAt(time.Now().Format(time.RFC3339))

err := vault.RecordCreate(context.Background(), record)
if err != nil {
    log.Fatal(err)
}
```

### Updating a Record

```go
// Find existing record
record, err := vault.RecordFindByToken(context.Background(), "my_token")
if err != nil {
    log.Fatal(err)
}

// Update value and metadata
record.SetValue("new_encrypted_value").
    SetData(map[string]string{
        "category": "credentials",
        "owner":    "user123",
    })

err = vault.RecordUpdate(context.Background(), record)
if err != nil {
    log.Fatal(err)
}
```

### Querying Records

```go
// Build query
query := vaultstore.NewRecordQuery().
    SetTokenIn([]string{"token1", "token2"}).
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder("desc").
    SetSoftDeletedInclude(false)

// Execute query
records, err := vault.RecordList(context.Background(), query)
if err != nil {
    log.Fatal(err)
}

// Count records
count, err := vault.RecordCount(context.Background(), query)
if err != nil {
    log.Fatal(err)
}
```

### Soft Delete and Recovery

```go
// Soft delete
record, err := vault.RecordFindByToken(context.Background(), "my_token")
if err != nil {
    log.Fatal(err)
}

err = vault.RecordSoftDelete(context.Background(), record)
if err != nil {
    log.Fatal(err)
}

// Find soft deleted records
query := vaultstore.NewRecordQuery().
    SetSoftDeletedInclude(true)

deletedRecords, err := vault.RecordList(context.Background(), query)
if err != nil {
    log.Fatal(err)
}
```

## Data Integrity

### Validation

Records undergo comprehensive validation:

```go
func (r *recordImplementation) Validate() error {
    if r.token == "" {
        return errors.New("token is required")
    }
    if r.value == "" {
        return errors.New("value is required")
    }
    // Additional validations...
    return nil
}
```

### Constraints

- **Token Uniqueness**: Enforced at database level
- **Required Fields**: Validated before operations
- **Data Format**: Timestamp and data format validation
- **Size Limits**: Maximum field length constraints

### Transactions

Critical operations use database transactions:

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

// Perform operations
err = s.createRecordInTx(tx, record)
if err != nil {
    return err
}

return tx.Commit()
```

## Performance Optimization

### Database Indexing

Optimized indexes for common queries:

```sql
-- Primary indexes
CREATE INDEX idx_vault_token ON vault(token);
CREATE INDEX idx_vault_id ON vault(id);

-- Query optimization indexes
CREATE INDEX idx_vault_created_at ON vault(created_at);
CREATE INDEX idx_vault_expires_at ON vault(expires_at);
CREATE INDEX idx_vault_soft_deleted_at ON vault(soft_deleted_at);
```

### Query Optimization

- **Column Selection**: Only select required columns
- **Limit Results**: Prevent large result sets
- **Efficient Filtering**: Use database-level filtering
- **Connection Pooling**: Reuse database connections

### Memory Management

- **Lazy Loading**: Load data only when needed
- **Change Tracking**: Minimize memory allocations
- **Resource Cleanup**: Proper connection and statement cleanup

## Error Handling

### Common Errors

- **ErrRecordNotFound**: Record doesn't exist
- **ErrTokenAlreadyExists**: Token uniqueness violation
- **ErrValidationFailed**: Data validation errors
- **ErrDatabaseError**: Database operation failures

### Error Recovery

```go
// Handle record not found
record, err := vault.RecordFindByToken(ctx, token)
if errors.Is(err, vaultstore.ErrRecordNotFound) {
    // Create new record
    record = createNewRecord(token)
    err = vault.RecordCreate(ctx, record)
}
```

## Testing

### Unit Tests

Comprehensive test coverage for:
- Record creation and validation
- CRUD operations
- Query functionality
- Error handling
- Change tracking

### Test Utilities

```go
// Helper for creating test records
func createTestRecord(token, value string) RecordInterface {
    return vaultstore.NewRecord().
        SetToken(token).
        SetValue(value).
        SetCreatedAt(time.Now().Format(time.RFC3339))
}

// Helper for comparing records
func assertRecordsEqual(t *testing.T, expected, actual RecordInterface) {
    assert.Equal(t, expected.GetID(), actual.GetID())
    assert.Equal(t, expected.GetToken(), actual.GetToken())
    assert.Equal(t, expected.GetValue(), actual.GetValue())
}
```

## Best Practices

### Record Design

1. **Use meaningful tokens** for easier identification
2. **Set appropriate expiration times** for temporary data
3. **Include relevant metadata** for application context
4. **Use soft delete** for data recovery needs

### Performance

1. **Limit query results** to prevent memory issues
2. **Use specific filters** instead of retrieving all records
3. **Batch operations** when processing multiple records
4. **Monitor query performance** regularly

### Data Management

1. **Validate input data** before creating records
2. **Handle errors gracefully** with proper logging
3. **Use transactions** for multi-record operations
4. **Implement cleanup** for expired records

## See Also

- [Core Store](core_store.md) - Main store implementation
- [Query Interface](query_interface.md) - Query building and execution
- [Token Operations](token_operations.md) - Token-specific operations
- [API Reference](../api_reference.md) - Complete API documentation
