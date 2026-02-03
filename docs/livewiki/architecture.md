---
path: architecture.md
page-type: overview
summary: System architecture, design patterns, and key architectural decisions including identity-based password management.
tags: [architecture, design, patterns, system, identity, metadata]
created: 2026-02-03
updated: 2026-02-03
version: 1.1.0
---

# Architecture

This document describes the architecture of VaultStore, including design patterns, key decisions, and system components.

## High-Level Architecture

VaultStore follows a layered architecture pattern with clear separation of concerns:

```mermaid
graph TB
    subgraph "Application Layer"
        A[Your Application]
    end
    
    subgraph "VaultStore Library"
        B[Store Interface]
        C[Token Operations]
        D[Record Management]
        E[Query Interface]
    end
    
    subgraph "Core Services"
        F[Encryption Service]
        G[Validation Service]
        H[Migration Service]
    end
    
    subgraph "Data Layer"
        I[GORM ORM]
        J[Database Driver]
        K[Database]
    end
    
    A --> B
    B --> C
    B --> D
    B --> E
    C --> F
    D --> F
    E --> G
    B --> H
    H --> I
    I --> J
    J --> K
```

## Core Components

### 1. Store Interface (`StoreInterface`)

The main interface that defines all vault operations. It provides:

- **Record Management**: CRUD operations for records
- **Token Operations**: Token lifecycle management
- **Query Operations**: Flexible querying capabilities
- **Database Management**: Migration and connection handling

### 2. Record System

Records are the fundamental data storage units:

```go
type RecordInterface interface {
    // Data access
    Data() map[string]string
    DataChanged() map[string]string
    
    // Getters
    GetID() string
    GetToken() string
    GetValue() string
    GetCreatedAt() string
    GetUpdatedAt() string
    GetExpiresAt() string
    GetSoftDeletedAt() string
    
    // Setters
    SetID(id string) RecordInterface
    SetToken(token string) RecordInterface
    SetValue(value string) RecordInterface
    // ... other setters
}
```

### 3. Token System

Tokens provide secure access to stored values:

- **Unique Generation**: Cryptographically secure random tokens
- **Password Protection**: Optional password encryption
- **Expiration**: Time-based access control
- **Soft Delete**: Recoverable deletion mechanism

### 4. Query Interface

Flexible querying system using the builder pattern:

```go
query := vaultstore.NewRecordQuery().
    SetToken("abc123").
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder("desc").
    SetSoftDeletedInclude(false)
```

## Design Patterns

### 1. Interface Segregation

VaultStore uses multiple focused interfaces rather than one large interface:

- `StoreInterface` - Main store operations
- `RecordInterface` - Record data operations
- `RecordQueryInterface` - Query building operations

### 2. Builder Pattern

The query system uses the builder pattern for flexible query construction:

```go
query := vaultstore.NewRecordQuery().
    SetToken(token).
    SetLimit(limit).
    SetOrderBy("created_at")
```

### 3. Factory Pattern

Store creation uses a factory pattern with options:

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "vault",
    DB:                 db,
    AutomigrateEnabled: true,
})
```

### 4. Repository Pattern

The store implementation acts as a repository, abstracting database operations:

```go
type storeImplementation struct {
    vaultTableName          string
    vaultMetaTableName      string
    automigrateEnabled      bool
    db                      *sql.DB
    gormDB                  *gorm.DB
    dbDriverName            string
    debugEnabled            bool
    cryptoConfig            *CryptoConfig
    passwordIdentityEnabled bool
}
```

## Data Flow

### Token Creation Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant Enc as Encryption
    participant DB as Database
    
    App->>Store: TokenCreate(value, password)
    Store->>Enc: GenerateToken()
    Enc-->>Store: token
    Store->>Enc: EncryptValue(value, password)
    Enc-->>Store: encryptedValue
    Store->>DB: CreateRecord(token, encryptedValue)
    DB-->>Store: record
    Store-->>App: token
```

### Token Read Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant DB as Database
    participant Enc as Encryption
    
    App->>Store: TokenRead(token, password)
    Store->>DB: FindByToken(token)
    DB-->>Store: record
    Store->>Enc: DecryptValue(encryptedValue, password)
    Enc-->>Store: value
    Store-->>App: value
```

## Security Architecture

### Encryption Strategy

1. **Value Encryption**: Stored values are encrypted using AES-256-GCM
2. **Password Protection**: Optional password-based encryption
3. **Token Security**: Cryptographically secure random token generation
4. **Data at Rest**: All sensitive data is encrypted in the database

### Access Control

```mermaid
graph LR
    A[Token] --> B{Password Protected?}
    B -->|Yes| C[Verify Password]
    B -->|No| D[Direct Access]
    C --> E[Decrypt Value]
    D --> E
    E --> F[Return Value]
```

## Database Schema

### Vault Table Structure

```sql
CREATE TABLE vault (
    id              TEXT PRIMARY KEY,
    token           TEXT UNIQUE NOT NULL,
    value           TEXT NOT NULL,           -- Encrypted value
    password_hash   TEXT,                   -- Optional password hash
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL,
    expires_at      TEXT,                   -- Optional expiration
    soft_deleted_at TEXT,                   -- Soft delete timestamp
    data            TEXT                    -- JSON metadata
);

CREATE INDEX IF NOT EXISTS idx_vault_soft_deleted_at ON vault(soft_deleted_at);
```

### Metadata Table Structure

The vault_meta table stores password identities and vault settings:

```sql
CREATE TABLE vault_meta (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    object_type  TEXT NOT NULL,    -- 'password_identity', 'record', 'vault'
    object_id    TEXT NOT NULL,    -- Unique ID for the object
    meta_key     TEXT NOT NULL,    -- 'hash', 'password_id', 'version'
    meta_value   TEXT,             -- Value (hash, ID, or setting)
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);

CREATE INDEX idx_vault_meta_obj ON vault_meta(object_type, object_id, meta_key);
```

#### Object Types

**Password Identity Objects:**
- `object_type`: `password_identity`
- `object_id`: `p_<uuid>` - Unique password ID
- `meta_key`: `hash`
- `meta_value`: Bcrypt/Argon2 hash of the password

**Record Link Objects:**
- `object_type`: `record`
- `object_id`: `r_<uuid>` - Record ID
- `meta_key`: `password_id`
- `meta_value`: Reference to password identity ID

**Vault Settings Objects:**
- `object_type`: `vault`
- `object_id`: `settings`
- `meta_key`: `version`
- `meta_value`: Vault version string

## Identity-Based Password Management

### Architecture Overview

The identity-based password management system uses a generic metadata table to track relationships between records and passwords:

```mermaid
graph TB
    subgraph "Password Identity"
        PI[Password Identity] --> |stores| PH[Password Hash]
    end
    
    subgraph "Records"
        R1[Record 1] --> |links to| PI
        R2[Record 2] --> |links to| PI
        R3[Record 3] --> |links to| PI
    end
    
    subgraph "Metadata Table"
        M1[password_identity p_xxx hash $2y$...]
        M2[record r_001 password_id p_xxx]
        M3[record r_002 password_id p_xxx]
    end
```

### Key Benefits

1. **Fast Bulk Rekey**: O(1) vs O(n) complexity
2. **Password Deduplication**: Automatic grouping by password
3. **Backward Compatible**: Existing records work without migration
4. **Optional Feature**: Can be enabled/disabled per store instance

### Migration Strategy

1. **On-Access Migration**: Records are linked to identities when accessed
2. **Batch Migration**: `MigrateRecordLinks()` processes all records for a password
3. **Version Tracking**: Vault settings track migration state

## Migration Strategy

### Auto-Migration

VaultStore provides automatic database migration:

```go
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "vault",
    DB:                 db,
    AutomigrateEnabled: true,  // Auto-create tables
})
```

### Migration Process

1. **Check Table Existence**: Verify if vault table exists
2. **Create Table**: Create table with proper schema if needed
3. **Index Creation**: Create necessary indexes for performance
4. **Validation**: Verify table structure

## Performance Considerations

### Database Optimization

1. **Indexing**: Token and ID fields are indexed
2. **Query Optimization**: Efficient SQL generation
3. **Connection Pooling**: Leverages database connection pooling
4. **Batch Operations**: Support for bulk operations where applicable

### Memory Management

1. **Lazy Loading**: Records loaded on demand
2. **Efficient Encryption**: Minimal memory overhead for encryption
3. **Resource Cleanup**: Proper resource management

## Error Handling Strategy

### Error Types

```go
// Validation errors
var ErrTokenRequired = errors.New("token is required")
var ErrPasswordRequired = errors.New("password is required")

// Database errors
var ErrRecordNotFound = errors.New("record not found")
var ErrTokenAlreadyExists = errors.New("token already exists")

// Encryption errors
var ErrInvalidPassword = errors.New("invalid password")
var ErrDecryptionFailed = errors.New("decryption failed")
```

### Error Handling Patterns

1. **Graceful Degradation**: Non-critical errors don't crash the system
2. **Contextual Errors**: Errors include relevant context
3. **Recovery Strategies**: Retry mechanisms for transient failures
4. **Logging**: Comprehensive error logging for debugging

## Extension Points

### Custom Encryption

VaultStore allows custom encryption implementations:

```go
type Encryptor interface {
    Encrypt(value string, password string) (string, error)
    Decrypt(encryptedValue string, password string) (string, error)
}
```

### Custom Validators

Add custom validation logic:

```go
type Validator interface {
    ValidateToken(token string) error
    ValidatePassword(password string) error
}
```

## See Also

- [API Reference](api_reference.md) - Detailed API documentation
- [Modules](modules/) - Individual module documentation
- [Getting Started](getting_started.md) - Setup and usage guide
- [Configuration](configuration.md) - Configuration options
- [Password Identity Management](modules/password_identity_management.md) - Identity-based password management

## Changelog

- **v1.1.0** (2026-02-03): Added documentation for vault_meta table structure and identity-based password management architecture.
- **v1.0.0** (2026-02-03): Initial architecture documentation
