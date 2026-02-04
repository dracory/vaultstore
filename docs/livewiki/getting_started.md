---
path: getting_started.md
page-type: tutorial
summary: Installation, setup, and quick start guide for VaultStore.
tags: [tutorial, installation, setup, quickstart]
created: 2026-02-03
updated: 2026-02-04
version: 1.1.0
---

# Getting Started

This guide will help you get VaultStore up and running in your Go application.

## Prerequisites

- Go 1.25 or higher
- A database (SQLite is recommended for getting started)
- Basic understanding of Go programming

## Installation

Add VaultStore to your project:

```bash
go get -u github.com/dracory/vaultstore
```

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/dracory/vaultstore"
    _ "github.com/glebarez/sqlite"
)

func main() {
    // Create database connection
    db, err := sql.Open("sqlite", "./vault.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create vault store
    vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
        VaultTableName:     "my_vault",
        DB:                 db,
        AutomigrateEnabled: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Your vault store is ready to use!
    log.Println("VaultStore initialized successfully")
}
```

### 2. Store and Retrieve a Secret

```go
// Create a token with a secret value
token, err := vault.TokenCreate(context.Background(), "my_secret_value", "my_password", 20)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Token created: %s\n", token)

// Check if token exists
exists, err := vault.TokenExists(context.Background(), token)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Token exists: %t\n", exists)

// Read the value using the token
value, err := vault.TokenRead(context.Background(), token, "my_password")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Retrieved value: %s\n", value)
```

## Configuration Options

### NewStoreOptions

```go
type NewStoreOptions struct {
    // Required: Name of the vault table
    VaultTableName string
    
    // Required: Database connection
    DB *sql.DB
    
    // Optional: Enable automatic database migration
    AutomigrateEnabled bool
    
    // Optional: Enable debug logging
    DebugEnabled bool
    
    // Optional: Database driver name (auto-detected if not provided)
    DbDriverName string
}
```

### Token Creation Options

```go
// TokenCreate with custom options
token, err := vault.TokenCreate(context.Background(), 
    "secret_value", 
    "password", 
    32, // token length
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Now().Add(24 * time.Hour), // expires in 24 hours
    })
```

## Database Setup

### SQLite (Recommended for Development)

```go
import (
    "database/sql"
    _ "github.com/glebarez/sqlite"
)

db, err := sql.Open("sqlite", "./vault.db")
```

### PostgreSQL

```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", "postgres://user:password@localhost/vaultdb?sslmode=disable")
```

### MySQL

```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/vaultdb")
```

## Common Operations

### Token Management

```go
// Create a custom token
err := vault.TokenCreateCustom(context.Background(), 
    "my_custom_token", 
    "secret_value", 
    "password")

// Update a token's value
err := vault.TokenUpdate(context.Background(), 
    token, 
    "new_value", 
    "password")

// Soft delete a token (recoverable)
err := vault.TokenSoftDelete(context.Background(), token)

// Hard delete a token (permanent)
err := vault.TokenDelete(context.Background(), token)

// Renew token expiration
err := vault.TokenRenew(context.Background(), 
    token, 
    time.Now().Add(24 * time.Hour))
```

### Query Operations

```go
// Find record by token
record, err := vault.RecordFindByToken(context.Background(), token)
if err != nil {
    log.Fatal(err)
}

// List records with query
query := vaultstore.RecordQuery().
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder("desc")

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

## Best Practices

1. **Use strong passwords** for token protection
2. **Set appropriate expiration times** for temporary secrets
3. **Enable soft delete** for data recovery options
4. **Use context with timeout** for production applications
5. **Handle errors gracefully** and check all return values
6. **Use connection pooling** for database connections

## Troubleshooting

### Common Issues

- **Database connection errors**: Verify database is running and credentials are correct
- **Migration failures**: Ensure database user has CREATE TABLE permissions
- **Token not found**: Check if token was soft deleted or expired
- **Password mismatch**: Verify password matches what was used during token creation

### Debug Mode

Enable debug mode for detailed logging:

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "my_vault",
    DB:                 db,
    AutomigrateEnabled: true,
    DebugEnabled:       true, // Enable debug logging
})
```

## See Also

- [API Reference](api_reference.md) - Complete API documentation
- [Configuration](configuration.md) - Detailed configuration options
- [Architecture](architecture.md) - System design and patterns
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

## Changelog

- **v1.1.0** (2026-02-04): Updated query builder example to use `RecordQuery()` instead of `NewRecordQuery()`
- **v1.0.0** (2026-02-03): Initial creation
