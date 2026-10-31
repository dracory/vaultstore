---
path: llm-context.md
page-type: overview
summary: Complete codebase summary optimized for LLM consumption including pure encryption bulk rekey features.
tags: [llm, context, summary, bulk-rekey, cryptoconfig]
created: 2026-02-03
updated: 2026-10-31
version: 1.3.0
---

# LLM Context: VaultStore

## Project Summary

VaultStore is a secure value storage (data-at-rest) implementation for Go, designed as a library component for securely storing and retrieving secrets in applications. It provides token-based access to encrypted values with optional password protection, flexible querying, and soft delete functionality. The project is specifically designed as a data store component, not a complete secrets management system with user management or API endpoints.

## Key Technologies

- **Go 1.27+** - Programming language
- **neat** - Database layer (ORM + schema, multi-dialect)
- **AES-256-GCM** - Encryption algorithm
- **Argon2id** - Password-based key derivation
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
│   │   │   └── tokens_change_password.md  # Pure encryption bulk rekey
│   │   └── *.md            # Other wiki pages
│   ├── proposals/          # Design proposals
│   └── *.md                # Other docs
├── interfaces.go           # Core interfaces
├── constants.go            # Constants and CryptoConfig
├── new.go                  # Store factory (NewStore) and options
├── store.go                # Store implementation, migrations, query builder
├── store_record_methods.go # Record CRUD operations
├── store_record_query.go   # Query execution and filtering
├── store_token_methods.go  # Token lifecycle management
├── store_tokens_change_password_methods.go # Pure encryption bulk rekey
├── record.go               # Record implementation (neat soft-delete model)
├── meta.go                 # Vault metadata (password identities, settings)
├── vault_settings.go       # Vault settings storage
├── encdec.go               # Encryption/decryption (AES-256-GCM + Argon2id)
├── functions.go            # Utility functions
├── is_token.go             # Token validation
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

### Store Implementation (`store*.go`, `new.go`)
- `new.go` - Factory function (`NewStore`) and `NewStoreOptions` struct with CryptoConfig and ParallelThreshold
- `store.go` - Core store logic, migrations, and query builder
- `store_record_methods.go` - Record CRUD operations
- `store_token_methods.go` - Token lifecycle management
- `store_record_query.go` - Query execution and filtering
- `store_tokens_change_password_methods.go` - Pure encryption bulk rekey with parallel processing

### Encryption (`encdec*.go`, `constants.go`)
- `encdec.go` - Main encryption/decryption functions using AES-256-GCM with Argon2id
- `constants.go` - Constants including CryptoConfig and encryption parameters
- `encdec_test.go` - Encryption functionality tests
- `encdec_v2_test.go` - Enhanced encryption tests

### Metadata (`meta.go`, `vault_settings.go`)
- `meta.go` - Metadata operations for password identities and vault settings
- `vault_settings.go` - Vault settings storage (meta table)

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
- `id` (STRING PRIMARY KEY, 40) - Unique record identifier
- `token` (STRING UNIQUE, 40) - Access token
- `value` (TEXT) - Encrypted stored value
- `created_at` (DATETIME) - Creation timestamp
- `updated_at` (DATETIME) - Last update timestamp
- `expires_at` (DATETIME) - Optional expiration timestamp
- `soft_deleted_at` (DATETIME) - Soft delete timestamp (max-date sentinel for "not deleted")

**Note**: The schema is database-agnostic and compatible with SQLite, PostgreSQL, and MySQL through neat.

### Metadata Table (vault_meta)

Stores password identities and vault settings:
- `id` (STRING PRIMARY KEY, 40) - Unique metadata ID
- `object_type` (STRING, 50) - 'password_identity', 'record', or 'vault'
- `object_id` (STRING, 64) - Unique object identifier
- `meta_key` (STRING, 50) - Key for the metadata value
- `meta_value` (TEXT) - The metadata value

## Security Implementation

1. **Encryption** - AES-256-GCM provides authenticated encryption
2. **Key Derivation** - Argon2id with configurable parameters (CryptoConfig)
3. **Random Generation** - crypto/rand for tokens, salts, and nonces
4. **Authentication** - Built-in integrity checking via GCM authentication tags
5. **Token Security** - Cryptographically secure random token generation
6. **Password Hashing** - Argon2id for password identity verification

## Configuration Options

Store configuration includes:
- `VaultTableName` - Database table name for vault records (required)
- `VaultMetaTableName` - Database table name for metadata (required for identity features)
- `DB` - Database connection (required)
- `DbDriverName` - Database driver specification (optional)
- `AutomigrateEnabled` - Automatic schema migration (optional)
- `DebugEnabled` - Debug logging (optional)
- `CryptoConfig` - Custom cryptographic parameters (optional)
- `ParallelThreshold` - Threshold for switching to parallel bulk rekey processing (optional)
- `PasswordAllowEmpty`, `PasswordMinLength`, `PasswordRequireLowercase/Uppercase/Numbers/Symbols` - Password policy (optional)

## Changelog

- **v1.3.0** (2026-10-31): Refreshed for the neat migration - replaced GORM/sb/PBKDF2 references, updated dependencies, schema, and file layout.
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
- `github.com/dracory/neat` - Database layer (ORM, schema, multi-dialect)
- `github.com/dromara/carbon/v2` - Date/time utilities
- `github.com/samber/lo` - Functional programming utilities
- `golang.org/x/crypto` - Cryptographic functions (Argon2id, AES-GCM)
- `modernc.org/sqlite` - Pure Go SQLite driver (no CGO)

Indirect (notable): `github.com/google/uuid` (UUID generation), `github.com/spf13/cast`, `golang.org/x/exp`, `golang.org/x/sys`, `golang.org/x/text`, `modernc.org/libc`.

## Performance Considerations

1. **Database Indexing** - Proper indexes on token, created_at, expires_at, soft_deleted_at
2. **Connection Pooling** - Configurable database connection pooling
3. **Query Optimization** - Column selection, result limiting, efficient filtering
4. **Memory Management** - Lazy loading, change tracking, resource cleanup
5. **Encryption Performance** - Configurable Argon2id parameters via CryptoConfig

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
- **Data Layer** - neat (database layer), database drivers, databases

## Extension Points

The system supports:
- Custom encryption implementations
- Custom validation logic
- Database-specific optimizations
- Additional metadata fields
- Custom query filters

## Licensing

The project is licensed under AGPL-3.0 with commercial licenses available for commercial use.
