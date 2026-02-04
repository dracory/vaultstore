---
path: llm-context.md
page-type: overview
summary: Complete codebase summary optimized for LLM consumption including pure encryption bulk rekey features.
tags: [llm, context, summary, bulk-rekey, cryptoconfig]
created: 2026-02-03
updated: 2026-02-04
version: 1.2.0
---

# LLM Context: VaultStore

## Project Summary

VaultStore is a secure value storage (data-at-rest) implementation for Go, designed as a library component for securely storing and retrieving secrets in applications. It provides token-based access to encrypted values with optional password protection, flexible querying, and soft delete functionality. The project is specifically designed as a data store component, not a complete secrets management system with user management or API endpoints.

## Key Technologies

- **Go 1.25+** - Programming language
- **GORM** - ORM for database operations
- **sb (SQL Builder)** - SQL query building
- **AES-256-GCM** - Encryption algorithm
- **PBKDF2** - Password-based key derivation
- **SQLite/PostgreSQL/MySQL** - Database-agnostic support
- **crypto/rand** - Cryptographically secure randomness

## Directory Structure

```
vaultstore/
├── docs/                   # Documentation
│   ├── livewiki/           # LiveWiki documentation
│   │   ├── modules/
│   │   │   ├── core_store.md
│   │   │   ├── encryption.md
│   │   │   ├── query_interface.md
│   │   │   ├── record_management.md
│   │   │   ├── token_operations.md
│   │   │   └── bulk_rekey.md         # Pure encryption bulk rekey
│   │   └── *.md            # Other wiki pages
│   ├── proposals/          # Design proposals
│   └── *.md                # Other docs
├── interfaces.go           # Core interfaces
├── constants.go            # Constants and CryptoConfig
├── store_*.go              # Store implementation
├── record_*.go             # Record operations
├── token_*.go              # Token operations
├── encdec*.go              # Encryption/decryption
├── functions.go            # Utility functions
├── gorm_model.go           # GORM models
├── is_token.go             # Token validation
├── sqls.go                 # SQL queries
├── meta_*.go               # Metadata operations
├── *_test.go               # Test files
├── go.mod                  # Go module
└── README.md               # Project README
```

## Core Concepts

1. **StoreInterface** - Main interface providing all vault operations including record management, token operations, database management, and identity-based password management
2. **RecordInterface** - Interface for record data operations with getters/setters for all record fields
3. **RecordQueryInterface** - Builder pattern interface for flexible query construction
4. **MetaInterface** - Interface for vault metadata operations (password identities, settings)
5. **Tokens** - Unique identifiers providing secure access to stored encrypted values
6. **Records** - Underlying data structure storing encrypted information with metadata
7. **Encryption** - AES-256-GCM encryption with Argon2id key derivation (configurable via CryptoConfig)
8. **Soft Delete** - Logical deletion mechanism with recovery capability
9. **Bulk Rekey** - Pure encryption password rotation without metadata (scan-and-test approach)
10. **Parallel Processing** - Worker pools for large dataset bulk operations

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
- `store_new_options.go` - NewStoreOptions struct with CryptoConfig and ParallelThreshold
- `store_implementation.go` - Core store logic and database operations
- `store_record_methods.go` - Record CRUD operations
- `store_token_methods.go` - Token lifecycle management
- `store_record_query.go` - Query execution and SQL generation
- `store_bulk_rekey_methods.go` - Pure encryption bulk rekey with parallel processing

### Encryption (`encdec*.go`, `constants.go`)
- `encdec.go` - Main encryption/decryption functions using AES-256-GCM with Argon2id
- `constants.go` - Constants including CryptoConfig and encryption parameters
- `encdec_test.go` - Encryption functionality tests
- `encdec_v2_test.go` - Enhanced encryption tests

### Metadata (`meta_*.go`)
- `meta_implementation.go` - Metadata operations implementation
- `meta_helpers.go` - Metadata helper functions
- `gorm_model.go` - GORM models including VaultMeta

### Interfaces (`interfaces.go`)
Defines `StoreInterface`, `RecordInterface`, `RecordQueryInterface`, and `MetaInterface` - the foundation of the entire system

## Key APIs

### Store Creation

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:          "vault",
    VaultMetaTableName:      "vault_meta",
    DB:                      db,
    AutomigrateEnabled:      true,
    DebugEnabled:            false,
    PasswordIdentityEnabled: true,              // Enable for bulk rekey optimization
    CryptoConfig:            vaultstore.DefaultCryptoConfig(), // Or HighSecurityCryptoConfig(), LightweightCryptoConfig()
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

// Change password for all tokens
// Uses pure encryption scan-and-test approach (no metadata storage)
count, err := vault.TokensChangePassword(ctx, "oldpass", "newpass")
```

### Query Operations
```go
query := vaultstore.RecordQuery().
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

### Metadata Table (vault_meta)

Stores password identities and vault settings:
- `id` (INTEGER PRIMARY KEY) - Unique metadata ID
- `object_type` (TEXT NOT NULL) - 'password_identity', 'record', or 'vault'
- `object_id` (TEXT NOT NULL) - Unique object identifier
- `meta_key` (TEXT NOT NULL) - Key for the metadata value
- `meta_value` (TEXT) - The metadata value
- `created_at` (TEXT NOT NULL) - Creation timestamp
- `updated_at` (TEXT NOT NULL) - Last update timestamp

## Security Implementation

1. **Encryption** - AES-256-GCM provides authenticated encryption
2. **Key Derivation** - Argon2id with configurable parameters (CryptoConfig)
3. **Random Generation** - crypto/rand for tokens, salts, and nonces
4. **Authentication** - Built-in integrity checking via GCM authentication tags
5. **Token Security** - Cryptographically secure random token generation
6. **Password Hashing** - Bcrypt or Argon2id for password identity verification

## Configuration Options

Store configuration includes:
- `VaultTableName` - Database table name for vault records (required)
- `VaultMetaTableName` - Database table name for metadata (required for identity features)
- `DB` - Database connection (required)
- `AutomigrateEnabled` - Automatic schema migration (optional)
- `DebugEnabled` - Debug logging (optional)
- `DbDriverName` - Database driver specification (optional)
- `CryptoConfig` - Custom cryptographic parameters (optional)
- `PasswordIdentityEnabled` - Enable identity-based password management (optional)

## Changelog

- **v1.1.0** (2026-02-03): Added identity-based password management, CryptoConfig, metadata table, and BulkRekey documentation.
- **v1.0.0** (2026-02-03): Initial LLM context documentation

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
- `github.com/dracory/database` - Database utilities
- `github.com/dracory/dataobject` - Data object utilities
- `github.com/dracory/sb` - SQL builder utilities
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
