# Usage Guide

This guide provides examples of how to use VaultStore to securely store and retrieve secrets.

## Scope

VaultStore is specifically designed as a data store component for securely storing and retrieving secrets. It is **not** an API or a complete secrets management system. Features such as user management, access control, and API endpoints are intentionally beyond the scope of this project.

VaultStore is meant to be integrated into your application as a library, providing the data storage layer for your secrets management needs. The application using VaultStore is responsible for implementing any additional layers such as API endpoints, user management, or access control if needed.

## Installation

```bash
go get github.com/dracory/vaultstore
```

## Basic Usage

### Creating a Store

To use VaultStore, you first need to create a store instance:

```go
import (
    "database/sql"
    "github.com/dracory/vaultstore"
    _ "github.com/mattn/go-sqlite3"
)

// Create a database connection
db, err := sql.Open("sqlite3", "vault.db")
if err != nil {
    panic(err)
}

// Create a new store
store, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "vault",
    DB:                 db,
    DbDriverName:       "sqlite3",
    AutomigrateEnabled: true,
    DebugEnabled:       false,
})
if err != nil {
    panic(err)
}
```

### Storing a Secret

To store a secret, use the `TokenCreate` method:

```go
ctx := context.Background()
value := "my-secret-value"
password := "my-password"
tokenLength := 20

token, err := store.TokenCreate(ctx, value, password, tokenLength)
if err != nil {
    panic(err)
}

fmt.Println("Your token:", token)
```

### Retrieving a Secret

To retrieve a secret, use the `TokenRead` method:

```go
ctx := context.Background()
token := "your-token"
password := "my-password"

value, err := store.TokenRead(ctx, token, password)
if err != nil {
    panic(err)
}

fmt.Println("Secret value:", value)
```

### Updating a Secret

To update a secret, use the `TokenUpdate` method:

```go
ctx := context.Background()
token := "your-token"
newValue := "my-new-secret-value"
password := "my-password"

err := store.TokenUpdate(ctx, token, newValue, password)
if err != nil {
    panic(err)
}
```

### Deleting a Secret

To delete a secret, use the `TokenDelete` method:

```go
ctx := context.Background()
token := "your-token"

err := store.TokenDelete(ctx, token)
if err != nil {
    panic(err)
}
```

### Soft Deleting a Secret

Soft deletion marks a record as deleted without actually removing it from the database. This allows for potential recovery of deleted data. To soft delete a secret, use the `TokenSoftDelete` method:

```go
ctx := context.Background()
token := "your-token"

err := store.TokenSoftDelete(ctx, token)
if err != nil {
    panic(err)
}
```

Soft-deleted records are not returned by default when using methods like `TokenExists`, `RecordFindByToken`, or `RecordList`. However, you can include soft-deleted records in your queries using the query interface:

```go
// Include soft-deleted records in a list query
query := vaultstore.RecordQuery().SetSoftDeletedInclude(true)
records, err := store.RecordList(ctx, query)

// Retrieve only soft-deleted records
onlySoftDeleted := vaultstore.RecordQuery().SetSoftDeletedOnly(true)
deletedRecords, err := store.RecordList(ctx, onlySoftDeleted)

// Find a specific soft-deleted record by token
query = vaultstore.RecordQuery().SetToken("your-token").SetSoftDeletedInclude(true)
records, err = store.RecordList(ctx, query)
if len(records) > 0 {
    // Found the soft-deleted record
    record := records[0]
    fmt.Println("Found soft-deleted record:", record.GetToken())
}
```

### Checking if a Token Exists

To check if a token exists, use the `TokenExists` method:

```go
ctx := context.Background()
token := "your-token"

exists, err := store.TokenExists(ctx, token)
if err != nil {
    panic(err)
}

if exists {
    fmt.Println("Token exists")
} else {
    fmt.Println("Token does not exist")
}
```

## Advanced Usage

### Creating a Custom Token

You can create a custom token instead of letting the system generate one:

```go
ctx := context.Background()
customToken := "my-custom-token"
value := "my-secret-value"
password := "my-password"

err := store.TokenCreateCustom(ctx, customToken, value, password)
if err != nil {
    panic(err)
}
```

### Reading Multiple Tokens

You can read multiple tokens at once:

```go
ctx := context.Background()
tokens := []string{"token1", "token2", "token3"}
password := "my-password"

values, err := store.TokensRead(ctx, tokens, password)
if err != nil {
    panic(err)
}

for token, value := range values {
    fmt.Printf("Token: %s, Value: %s\n", token, value)
}
```

### Using the Query Interface

VaultStore provides a flexible query interface for searching and filtering records:

```go
ctx := context.Background()

// Create a query to find the 10 most recently created records
query := vaultstore.RecordQuery().
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder("desc")

records, err := store.RecordList(ctx, query)
if err != nil {
    panic(err)
}

// Display the records
for _, record := range records {
    fmt.Printf("ID: %s, Token: %s\n", record.GetID(), record.GetToken())
}
```

#### Filtering by Token

```go
ctx := context.Background()
query := vaultstore.RecordQuery().
    SetToken("my-token")

record, err := store.RecordFindByToken(ctx, "my-token")
if err != nil {
    panic(err)
}
```

#### Counting Records

```go
ctx := context.Background()
query := vaultstore.RecordQuery().
    SetCountOnly(true)

count, err := store.RecordCount(ctx, query)
if err != nil {
    panic(err)
}

fmt.Printf("Total records: %d\n", count)
```

For more detailed information about the query interface, see the [Query Interface documentation](./query_interface.md).

### Working with Records Directly

If you need more control, you can work with records directly:

```go
ctx := context.Background()

// Create a new record
record := vaultstore.NewRecord()
record.SetToken("my-token")
record.SetValue("my-encrypted-value")

err := store.RecordCreate(ctx, *record)
if err != nil {
    panic(err)
}

// Find a record by token
foundRecord, err := store.RecordFindByToken(ctx, "my-token")
if err != nil {
    panic(err)
}

// Update a record
foundRecord.SetValue("new-encrypted-value")
err = store.RecordUpdate(ctx, *foundRecord)
if err != nil {
    panic(err)
}

// Delete a record
err = store.RecordDeleteByID(ctx, foundRecord.ID())
if err != nil {
    panic(err)
}
