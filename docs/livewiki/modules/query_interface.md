---
path: modules/query_interface.md
page-type: module
summary: Flexible query building and execution system for record retrieval.
tags: [module, query, interface, filtering]
created: 2026-02-03
updated: 2026-02-04
version: 1.1.0
---

# Query Interface Module

The query interface module provides a flexible, type-safe query building system for retrieving vault records with various filters, sorting, and pagination options.

## Overview

The query interface module enables:
- Flexible record filtering and searching
- Type-safe query building with method chaining
- Pagination and result limiting
- Column selection optimization
- Soft delete handling

## Core Interface

### RecordQueryInterface

Defines the contract for query building:

```go
type RecordQueryInterface interface {
    // Validation
    Validate() error

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
}
```

## Query Building

### Factory Function

```go
func RecordQuery() RecordQueryInterface
```

Creates a new query builder instance with default values.

### Method Chaining

The query interface supports fluent method chaining:

```go
query := vaultstore.RecordQuery().
    SetToken("abc123").
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder(vaultstore.DESC).
    SetSoftDeletedInclude(false)
```

## Filter Options

### ID Filtering

#### Single ID

```go
// Filter by specific ID
query := vaultstore.RecordQuery().
    SetID("550e8400-e29b-41d4-a716-446655440000")
```

#### Multiple IDs

```go
// Filter by list of IDs
query := vaultstore.RecordQuery().
    SetIDIn([]string{
        "550e8400-e29b-41d4-a716-446655440000",
        "550e8400-e29b-41d4-a716-446655440001",
        "550e8400-e29b-41d4-a716-446655440002",
    })
```

### Token Filtering

#### Single Token

```go
// Filter by specific token
query := vaultstore.RecordQuery().
    SetToken("my_token_abc123")
```

#### Multiple Tokens

```go
// Filter by list of tokens
query := vaultstore.RecordQuery().
    SetTokenIn([]string{
        "token1",
        "token2",
        "token3",
    })
```

### Soft Delete Filtering

```go
// Include soft deleted records
query := vaultstore.RecordQuery().
    SetSoftDeletedInclude(true)

// Exclude soft deleted records (default)
query := vaultstore.RecordQuery().
    SetSoftDeletedInclude(false)
```

## Pagination

### Limit and Offset

```go
// Get first 10 records
query := vaultstore.RecordQuery().
    SetLimit(10).
    SetOffset(0)

// Get records 11-20 (second page)
query := vaultstore.RecordQuery().
    SetLimit(10).
    SetOffset(10)
```

### Pagination Helper

```go
// Pagination helper function
func paginateQuery(page, pageSize int) RecordQueryInterface {
    offset := (page - 1) * pageSize
    return vaultstore.RecordQuery().
        SetLimit(pageSize).
        SetOffset(offset)
}

// Usage
page := 2
pageSize := 25
query := paginateQuery(page, pageSize)
```

## Sorting

### Order By

```go
// Sort by creation time (ascending)
query := vaultstore.RecordQuery().
    SetOrderBy("created_at").
    SetSortOrder(vaultstore.ASC)

// Sort by creation time (descending)
query := vaultstore.RecordQuery().
    SetOrderBy("created_at").
    SetSortOrder(vaultstore.DESC)
```

### Sortable Columns

Supported sort columns:
- `id` - Record ID
- `token` - Token value
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp
- `expires_at` - Expiration timestamp
- `soft_deleted_at` - Soft delete timestamp

## Column Selection

### Select Specific Columns

```go
// Select only specific columns for performance
query := vaultstore.RecordQuery().
    SetColumns([]string{"id", "token", "created_at"})
```

### Common Column Sets

```go
// Lightweight query (no value data)
func lightweightQuery() RecordQueryInterface {
    return vaultstore.RecordQuery().
        SetColumns([]string{"id", "token", "created_at", "expires_at"})
}

// Full query (include all data)
func fullQuery() RecordQueryInterface {
    return vaultstore.RecordQuery()
    // No SetColumns() - selects all columns
}

// Metadata query (only timestamps)
func metadataQuery() RecordQueryInterface {
    return vaultstore.RecordQuery().
        SetColumns([]string{"id", "created_at", "updated_at", "expires_at"})
}
```

## Special Query Types

### Count Query

```go
// Count records matching criteria
query := vaultstore.RecordQuery().
    SetToken("abc123").
    SetCountOnly(true)

count, err := vault.RecordCount(context.Background(), query)
```

### Existence Check

```go
// Check if any records match criteria
func exists(vault StoreInterface, token string) (bool, error) {
    query := vaultstore.RecordQuery().
        SetToken(token).
        SetLimit(1).
        SetCountOnly(true)

    count, err := vault.RecordCount(context.Background(), query)
    return count > 0, err
}
```

## Query Execution

### Record List

```go
// Execute query and get records
records, err := vault.RecordList(context.Background(), query)
if err != nil {
    log.Fatal(err)
}

for _, record := range records {
    fmt.Printf("Token: %s, Created: %s\n", 
        record.GetToken(), record.GetCreatedAt())
}
```

### Record Count

```go
// Count matching records
count, err := vault.RecordCount(context.Background(), query)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d records\n", count)
```

## Usage Examples

### Basic Queries

```go
// Find record by token
query := vaultstore.RecordQuery().
    SetToken("my_token")

record, err := vault.RecordFindByToken(context.Background(), "my_token")
if err != nil {
    log.Fatal(err)
}
```

### Advanced Filtering

```go
// Find multiple tokens, exclude deleted, sort by creation
query := vaultstore.RecordQuery().
    SetTokenIn([]string{"token1", "token2", "token3"}).
    SetSoftDeletedInclude(false).
    SetOrderBy("created_at").
    SetSortOrder("desc").
    SetLimit(10)

records, err := vault.RecordList(context.Background(), query)
if err != nil {
    log.Fatal(err)
}
```

### Pagination Example

```go
// Paginated token listing
func listTokensPaginated(vault StoreInterface, page, pageSize int) ([]RecordInterface, error) {
    query := vaultstore.RecordQuery().
        SetSoftDeletedInclude(false).
        SetOrderBy("created_at").
        SetSortOrder("desc").
        SetLimit(pageSize).
        SetOffset((page - 1) * pageSize)

    return vault.RecordList(context.Background(), query)
}

// Usage
page := 1
pageSize := 25
records, err := listTokensPaginated(vault, page, pageSize)
```

### Search Functionality

```go
// Search tokens by pattern (using LIKE)
func searchTokens(vault StoreInterface, pattern string, limit int) ([]RecordInterface, error) {
    // This would require extending the query interface
    // For now, we can use token filtering with exact matches
    query := vaultstore.RecordQuery().
        SetSoftDeletedInclude(false).
        SetOrderBy("token").
        SetSortOrder("asc").
        SetLimit(limit)

    return vault.RecordList(context.Background(), query)
}
```

## Performance Optimization

### Column Selection

```go
// Optimize for performance by selecting only needed columns
func getTokenMetadata(vault StoreInterface, token string) (RecordInterface, error) {
    query := vaultstore.RecordQuery().
        SetToken(token).
        SetColumns([]string{"id", "token", "created_at", "expires_at"})

    records, err := vault.RecordList(context.Background(), query)
    if err != nil {
        return nil, err
    }
    
    if len(records) == 0 {
        return nil, ErrRecordNotFound
    }
    
    return records[0], nil
}
```

### Index Utilization

The query interface leverages database indexes:

```sql
-- Important indexes for query performance
CREATE INDEX idx_vault_token ON vault(token);
CREATE INDEX idx_vault_created_at ON vault(created_at);
CREATE INDEX idx_vault_expires_at ON vault(expires_at);
CREATE INDEX idx_vault_soft_deleted_at ON vault(soft_deleted_at);
```

### Query Planning

```go
// Efficient query for recent tokens
func getRecentTokens(vault StoreInterface, since time.Time, limit int) ([]RecordInterface, error) {
    query := vaultstore.RecordQuery().
        SetSoftDeletedInclude(false).
        SetOrderBy("created_at").
        SetSortOrder("desc").
        SetLimit(limit)

    return vault.RecordList(context.Background(), query)
}
```

## Implementation Details

### Query Validation

```go
func (q *recordQueryImplementation) Validate() error {
    // Validate limit
    if q.IsLimitSet() && q.GetLimit() < 0 {
        return errors.New("limit cannot be negative")
    }
    
    // Validate offset
    if q.IsOffsetSet() && q.GetOffset() < 0 {
        return errors.New("offset cannot be negative")
    }
    
    // Validate sort order
    if q.IsSortOrderSet() {
        sortOrder := q.GetSortOrder()
        if sortOrder != "asc" && sortOrder != "desc" {
            return errors.New("sort order must be 'asc' or 'desc'")
        }
    }
    
    return nil
}
```

### SQL Generation

The query interface uses the `sb` (SQL Builder) package for SQL generation. Query parameters are stored internally and converted to SQL when executed:

```go
const (
    ASC  = "asc"
    DESC = "desc"
)

type recordQueryImpl struct {
    properties map[string]interface{}
}

// Properties are stored and retrieved via getter/setter methods
// Example: SetToken("abc") stores in properties["token"]
```

The store implementation processes queries directly, building SQL using the `sb` package based on the query properties.

## Error Handling

### Common Query Errors

```go
var (
    ErrInvalidQuery      = errors.New("invalid query parameters")
    ErrInvalidLimit      = errors.New("invalid limit value")
    ErrInvalidOffset     = errors.New("invalid offset value")
    ErrInvalidSortOrder  = errors.New("invalid sort order")
    ErrInvalidColumn     = errors.New("invalid column name")
)
```

### Error Handling Patterns

```go
func safeQueryExecution(vault StoreInterface, query RecordQueryInterface) ([]RecordInterface, error) {
    // Validate query
    if err := query.Validate(); err != nil {
        return nil, fmt.Errorf("query validation failed: %w", err)
    }
    
    // Execute query
    records, err := vault.RecordList(context.Background(), query)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    
    return records, nil
}
```

## Testing

### Unit Tests

```go
func TestQueryBuilder(t *testing.T) {
    query := vaultstore.RecordQuery().
        SetToken("test_token").
        SetLimit(10).
        SetOrderBy("created_at").
        SetSortOrder("desc").
        SetSoftDeletedInclude(false)

    // Test query properties
    assert.True(t, query.IsTokenSet())
    assert.Equal(t, "test_token", query.GetToken())
    assert.True(t, query.IsLimitSet())
    assert.Equal(t, 10, query.GetLimit())
    assert.True(t, query.IsOrderBySet())
    assert.Equal(t, "created_at", query.GetOrderBy())
    assert.True(t, query.IsSortOrderSet())
    assert.Equal(t, "desc", query.GetSortOrder())
    assert.False(t, query.GetSoftDeletedInclude())
}

func TestQueryValidation(t *testing.T) {
    // Test invalid limit
    query := vaultstore.RecordQuery().SetLimit(-1)
    err := query.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "limit cannot be negative")

    // Test invalid sort order
    query = vaultstore.RecordQuery().SetSortOrder("invalid")
    err = query.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "sort order must be 'asc' or 'desc'")
}
```

### Integration Tests

```go
func TestQueryExecution(t *testing.T) {
    vault := createTestStore(t)
    ctx := context.Background()

    // Create test records
    tokens := []string{"token1", "token2", "token3"}
    for _, token := range tokens {
        _, err := vault.TokenCreate(ctx, "test_value", "", 32)
        require.NoError(t, err)
    }

    // Test query
    query := vaultstore.RecordQuery().
        SetTokenIn(tokens).
        SetLimit(10).
        SetOrderBy("created_at").
        SetSortOrder("desc")

    records, err := vault.RecordList(ctx, query)
    require.NoError(t, err)
    assert.Len(t, records, 3)

    // Test count
    count, err := vault.RecordCount(ctx, query)
    require.NoError(t, err)
    assert.Equal(t, int64(3), count)
}
```

## Best Practices

### Performance

1. **Select specific columns** when you don't need all data
2. **Use appropriate limits** to prevent large result sets
3. **Leverage database indexes** through proper filtering
4. **Avoid SELECT *** in production queries

### Security

1. **Validate all input parameters** before query execution
2. **Use parameterized queries** through the database abstraction layer
3. **Implement proper access controls** at the application level
4. **Log query patterns** for security monitoring

### Design

1. **Build queries incrementally** using method chaining
2. **Validate queries** before execution
3. **Handle errors gracefully** with proper context
4. **Use appropriate pagination** for large datasets

## See Also

- [Record Management](record_management.md) - Record operations
- [Core Store](core_store.md) - Store implementation
- [API Reference](../api_reference.md) - Complete API documentation
- [Database Schema](../data_stores.md) - Database structure and indexes

## Changelog

- **v1.1.0** (2026-02-04): Removed goqu dependency, changed factory function from `NewRecordQuery()` to `RecordQuery()`, added ASC/DESC constants, updated SQL generation documentation
- **v1.0.0** (2026-02-03): Initial creation
