---
path: llm-context.md
page-type: overview
summary: Complete codebase summary optimized for LLM consumption.
tags: [llm, context, summary]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
---

# LLM Context: VaultStore

## Project Summary

VaultStore is a secure value storage (data-at-rest) implementation for Go, designed as a library component for securely storing and retrieving secrets in applications. It provides token-based access to encrypted values with optional password protection, flexible querying, and soft delete functionality. The project is specifically designed as a data store component, not a complete secrets management system with user management or API endpoints.

## Key Technologies

- **Go 1.25+** - Programming language
- **GORM** - ORM for database operations
- **goqu** - Query builder for complex queries
- **AES-256-GCM** - Encryption algorithm
- **PBKDF2** - Password-based key derivation
- **SQLite/PostgreSQL/MySQL** - Database-agnostic support
- **crypto/rand** - Cryptographically secure randomness

## Directory Structure

```
vaultstore/
├── docs/                   # Documentation
│   ├── livewiki/           # LiveWiki documentation
│   ├── proposals/          # Design proposals
│   └── *.md                # Other docs
├── interfaces.go           # Core interfaces
├── store_*.go             # Store implementation
├── record_*.go             # Record operations
├── token_*.go              # Token operations
├── encdec*.go              # Encryption/decryption
├── functions.go           # Utility functions
├── consts.go               # Constants
├── gorm_model.go           # GORM models
├── is_token.go             # Token validation
├── sqls.go                 # SQL queries
├── *_test.go               # Test files
├── go.mod                  # Go module
└── README.md               # Project README
```

## Core Concepts

1. **StoreInterface** - Main interface providing all vault operations including record management, token operations, and database management
2. **RecordInterface** - Interface for record data operations with getters/setters for all record fields
3. **RecordQueryInterface** - Builder pattern interface for flexible query construction
4. **Tokens** - Unique identifiers providing secure access to stored encrypted values
5. **Records** - Underlying data structure storing encrypted information with metadata
6. **Encryption** - AES-256-GCM encryption with optional password-based protection using PBKDF2
7. **Soft Delete** - Logical deletion mechanism with recovery capability

## Common Patterns

1. **Interface Segregation** - Multiple focused interfaces rather than one large interface
2. **Builder Pattern** - Query system uses builder pattern for flexible construction
3. **Factory Pattern** - Store creation using factory function with options
4. **Repository Pattern** - Store implementation acts as repository abstracting database operations
5. **Method Chaining** - Query builders and record setters support fluent chaining
6. **Error Handling** - Comprehensive error types with specific error values for different failure modes

## Important Files

### Core Interfaces (`interfaces.go`)
Defines `StoreInterface`, `RecordInterface`, and `RecordQueryInterface` - the foundation of the entire system

### Store Implementation (`store_*.go`)
- `store_new.go` - Factory function for creating store instances
- `store_implementation.go` - Core store logic and database operations
- `store_record_methods.go` - Record CRUD operations
- `store_token_methods.go` - Token lifecycle management
- `store_record_query.go` - Query execution and SQL generation

### Encryption (`encdec*.go`)
- `encdec.go` - Main encryption/decryption functions using AES-256-GCM
- `encdec_test.go` - Encryption functionality tests
- `encdec_v2_test.go` - Enhanced encryption tests

### Models (`gorm_model.go`)
GORM database models defining the vault table structure and relationships

### Token Validation (`is_token.go`)
Token format validation and generation utilities

### SQL Queries (`sqls.go`)
Predefined SQL queries and database operations

## Key APIs

### Store Creation
```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "vault",
    DB:                 db,
    AutomigrateEnabled: true,
    DebugEnabled:       false,
})
```

### Token Operations
```go
// Create token
token, err := vault.TokenCreate(ctx, "value", "password", 32)

// Read value
value, err := vault.TokenRead(ctx, token, "password")

// Update value
err = vault.TokenUpdate(ctx, token, "new_value", "password")

// Delete token
err = vault.TokenDelete(ctx, token)
```

### Query Operations
```go
query := vaultstore.NewRecordQuery().
    SetToken("abc123").
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder("desc")

records, err := vault.RecordList(ctx, query)
```

## Database Schema

The vault table stores encrypted data with the following structure (database-agnostic):
- `id` (TEXT PRIMARY KEY) - Unique record identifier
- `token` (TEXT UNIQUE NOT NULL) - Access token
- `value` (TEXT NOT NULL) - Encrypted stored value
- `password_hash` (TEXT) - Optional password hash
- `created_at` (TEXT NOT NULL) - Creation timestamp
- `updated_at` (TEXT NOT NULL) - Last update timestamp
- `expires_at` (TEXT) - Optional expiration timestamp
- `soft_deleted_at` (TEXT) - Soft delete timestamp
- `data` (TEXT) - JSON metadata

**Note**: The schema is database-agnostic and compatible with SQLite, PostgreSQL, and MySQL through GORM.

## Security Implementation

1. **Encryption** - AES-256-GCM provides authenticated encryption
2. **Key Derivation** - PBKDF2 with 100,000 iterations for password-based keys
3. **Random Generation** - crypto/rand for tokens, salts, and nonces
4. **Authentication** - Built-in integrity checking via GCM authentication tags
5. **Token Security** - Cryptographically secure random token generation

## Error Handling

Common error types include:
- `ErrRecordNotFound` - Record or token not found
- `ErrTokenAlreadyExists` - Token uniqueness violation
- `ErrInvalidPassword` - Password verification failed
- `ErrDecryptionFailed` - Encryption/decryption operation failed
- `ErrValidationFailed` - Input validation errors

## Testing Strategy

Comprehensive test coverage including:
- Unit tests for all major components
- Integration tests for different databases
- Security tests for encryption functionality
- Performance benchmarks for critical operations
- Error handling validation

## Dependencies

External dependencies:
- `github.com/doug-martin/goqu/v9` - Query builder
- `github.com/dracory/database` - Database utilities
- `github.com/dracory/dataobject` - Data object utilities
- `github.com/dracory/sb` - String builder utilities
- `github.com/dracory/uid` - UID generation
- `github.com/dromara/carbon/v2` - Date/time utilities
- `github.com/glebarez/sqlite` - SQLite driver
- `github.com/samber/lo` - Functional programming utilities
- `golang.org/x/crypto` - Cryptographic functions
- `gorm.io/gorm` - ORM framework

## Configuration Options

Store configuration includes:
- `VaultTableName` - Database table name (required)
- `DB` - Database connection (required)
- `AutomigrateEnabled` - Automatic schema migration (optional)
- `DebugEnabled` - Debug logging (optional)
- `DbDriverName` - Database driver specification (optional)

## Performance Considerations

1. **Database Indexing** - Proper indexes on token, created_at, expires_at, soft_deleted_at
2. **Connection Pooling** - Configurable database connection pooling
3. **Query Optimization** - Column selection, result limiting, efficient filtering
4. **Memory Management** - Lazy loading, change tracking, resource cleanup
5. **Encryption Performance** - Hardware acceleration support, configurable PBKDF2 iterations

## Use Cases

VaultStore is ideal for:
- Storing API keys and secrets securely
- Managing temporary access tokens
- Implementing secure configuration storage
- Handling sensitive user data
- Providing encrypted data persistence
- Building custom secrets management systems

## Architecture Patterns

The system follows layered architecture with:
- **Application Layer** - User applications
- **VaultStore Library** - Core functionality
- **Core Services** - Encryption, validation, migration
- **Data Layer** - GORM ORM, database drivers, databases

## Extension Points

The system supports:
- Custom encryption implementations
- Custom validation logic
- Database-specific optimizations
- Additional metadata fields
- Custom query filters

## Licensing

The project is licensed under AGPL-3.0 with commercial licenses available for commercial use.
