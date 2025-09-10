# Query Interface

This document describes the query interface for VaultStore, which provides a flexible way to search and filter records.

## Overview

The `RecordQueryInterface` allows for complex querying of records in the VaultStore. It provides a fluent interface for building queries with various filters and options.

## Creating a Query

To create a new query, use the `RecordQuery` function:

```go
import "github.com/dracory/vaultstore"

query := vaultstore.RecordQuery()
```

## Query Methods

The query interface provides the following methods for filtering and configuring the query:

### ID Filtering

- `SetID(id string)`: Filter records by a specific ID
- `SetIDIn(idIn []string)`: Filter records by a list of IDs
- `IsIDSet()`: Check if ID filter is set
- `GetID()`: Get the current ID filter
- `IsIDInSet()`: Check if IDIn filter is set
- `GetIDIn()`: Get the current IDIn filter

### Token Filtering

- `SetToken(token string)`: Filter records by a specific token
- `SetTokenIn(tokenIn []string)`: Filter records by a list of tokens
- `IsTokenSet()`: Check if token filter is set
- `GetToken()`: Get the current token filter
- `IsTokenInSet()`: Check if tokenIn filter is set
- `GetTokenIn()`: Get the current tokenIn filter

### Pagination

- `SetLimit(limit int)`: Set the maximum number of records to return
- `SetOffset(offset int)`: Set the number of records to skip
- `IsLimitSet()`: Check if limit is set
- `GetLimit()`: Get the current limit
- `IsOffsetSet()`: Check if offset is set
- `GetOffset()`: Get the current offset

### Sorting

- `SetOrderBy(orderBy string)`: Set the field to order by
- `SetSortOrder(sortOrder string)`: Set the sort order ("asc" or "desc")
- `IsOrderBySet()`: Check if orderBy is set
- `GetOrderBy()`: Get the current orderBy
- `IsSortOrderSet()`: Check if sortOrder is set
- `GetSortOrder()`: Get the current sortOrder

### Other Options

- `SetCountOnly(countOnly bool)`: Set to true to only return the count of records
- `SetSoftDeletedInclude(softDeletedInclude bool)`: Set to true to include soft-deleted records in the results
- `SetSoftDeletedOnly(softDeletedOnly bool)`: Set to true to only return soft-deleted records
- `IsCountOnlySet()`: Check if countOnly is set
- `GetCountOnly()`: Get the current countOnly value
- `IsSoftDeletedIncludeSet()`: Check if softDeletedInclude is set
- `GetSoftDeletedInclude()`: Get the current softDeletedInclude value
- `IsSoftDeletedOnlySet()`: Check if softDeletedOnly is set
- `GetSoftDeletedOnly()`: Get the current softDeletedOnly value

### Validation

- `Validate()`: Validates the query and returns an error if any constraints are violated

## Example Usage

### Basic Query

```go
query := vaultstore.RecordQuery().
    SetLimit(10).
    SetOffset(0).
    SetOrderBy("created_at").
    SetSortOrder("desc")

records, err := store.RecordList(ctx, query)
if err != nil {
    // Handle error
}
```

### Filtering by ID

```go
query := vaultstore.RecordQuery().
    SetID("123e4567-e89b-12d3-a456-426614174000")

record, err := store.RecordList(ctx, query)
if err != nil {
    // Handle error
}
```

### Filtering by Multiple IDs

```go
query := vaultstore.RecordQuery().
    SetIDIn([]string{
        "123e4567-e89b-12d3-a456-426614174000",
        "223e4567-e89b-12d3-a456-426614174001",
    })

records, err := store.RecordList(ctx, query)
if err != nil {
    // Handle error
}
```

### Counting Records

```go
query := vaultstore.RecordQuery().
    SetToken("my-token").
    SetCountOnly(true)

count, err := store.RecordCount(ctx, query)
if err != nil {
    // Handle error
}
```

### Including Soft-Deleted Records

```go
query := vaultstore.RecordQuery().
    SetSoftDeletedInclude(true)

records, err := store.RecordList(ctx, query)
if err != nil {
    // Handle error
}
```

### Retrieving Only Soft-Deleted Records

```go
query := vaultstore.RecordQuery().
    SetSoftDeletedOnly(true)

softDeletedRecords, err := store.RecordList(ctx, query)
if err != nil {
    // Handle error
}
```

## Important Notes

1. The query interface is designed to be immutable - each method returns a new instance of the query.
2. When using `SetCountOnly(true)`, the `SetLimit` and `SetOffset` methods will be ignored.
3. The default sort order is descending ("desc") if not specified.
4. By default, soft-deleted records are not included in the results. Use `SetSoftDeletedInclude(true)` to include them.
