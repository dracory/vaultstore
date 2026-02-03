---
path: data_flow.md
page-type: overview
summary: How data moves through the VaultStore system including identity-based password management flows.
tags: [data-flow, architecture, processing, lifecycle, identity]
created: 2026-02-03
updated: 2026-02-03
version: 1.1.0
---

# Data Flow

This document describes how data moves through the VaultStore system, from creation to retrieval and deletion.

## Data Lifecycle Overview

```mermaid
graph TB
    A[Input Data] --> B[Encryption]
    B --> C[Token Generation]
    C --> D[Database Storage]
    D --> E[Query Operations]
    E --> F[Decryption]
    F --> G[Output Data]
    
    D --> H[Soft Delete]
    H --> I[Recovery]
    I --> E
    
    D --> J[Hard Delete]
    J --> K[Permanent Removal]
```

## Token Creation Flow

### Step-by-Step Process

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant TokenGen as Token Generator
    participant Encrypt as Encryption Service
    participant DB as Database
    
    App->>Store: TokenCreate(value, password, length)
    Store->>TokenGen: GenerateSecureToken(length)
    TokenGen-->>Store: token
    Store->>Encrypt: EncryptValue(value, password)
    Encrypt-->>Store: encryptedValue
    Store->>DB: CreateRecord(token, encryptedValue)
    DB-->>Store: record
    Store-->>App: token
```

### Data Transformations

1. **Input**: Plain text value and optional password
2. **Token Generation**: Cryptographically secure random token
3. **Encryption**: Value encrypted using AES-256-GCM with password-derived key
4. **Storage**: Token and encrypted value stored in database
5. **Output**: Generated token returned to application

### Security Considerations

- **Token Generation**: Uses `crypto/rand` for cryptographically secure randomness
- **Encryption**: AES-256-GCM provides authenticated encryption
- **Password Handling**: Passwords are never stored in plain text
- **Key Derivation**: Uses PBKDF2 for password-based key derivation

## Token Retrieval Flow

### Step-by-Step Process

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant DB as Database
    participant Encrypt as Encryption Service
    
    App->>Store: TokenRead(token, password)
    Store->>DB: FindByToken(token)
    DB-->>Store: record
    Store->>Encrypt: DecryptValue(encryptedValue, password)
    Encrypt-->>Store: value
    Store-->>App: value
```

### Data Validation

1. **Token Existence**: Verify token exists in database
2. **Soft Delete Check**: Ensure token is not soft deleted
3. **Expiration Check**: Verify token has not expired
4. **Password Verification**: Validate password if required
5. **Decryption**: Decrypt and return the value

### Error Handling

- **Token Not Found**: Return `ErrRecordNotFound`
- **Soft Deleted**: Return `ErrRecordNotFound`
- **Expired**: Return `ErrRecordNotFound`
- **Invalid Password**: Return `ErrInvalidPassword`
- **Decryption Failed**: Return `ErrDecryptionFailed`

## Query Operations Flow

### Query Building

```mermaid
graph LR
    A[Query Builder] --> B[Validation]
    B --> C[SQL Generation]
    C --> D[Database Query]
    D --> E[Result Mapping]
    E --> F[Record Objects]
```

### Query Execution

1. **Query Construction**: Build query using builder pattern
2. **Validation**: Validate query parameters
3. **SQL Generation**: Convert query to SQL using goqu
4. **Database Execution**: Execute SQL query
5. **Result Mapping**: Map database rows to record objects
6. **Filtering**: Apply additional filters (soft delete, etc.)

### Query Types

#### Simple Query

```go
query := vaultstore.NewRecordQuery().
    SetToken("abc123")
```

#### Complex Query

```go
query := vaultstore.NewRecordQuery().
    SetTokenIn([]string{"token1", "token2"}).
    SetLimit(10).
    SetOrderBy("created_at").
    SetSortOrder("desc").
    SetSoftDeletedInclude(false)
```

## Update Operations Flow

### Token Update Process

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant DB as Database
    participant Encrypt as Encryption Service
    
    App->>Store: TokenUpdate(token, newValue, password)
    Store->>DB: FindByToken(token)
    DB-->>Store: existingRecord
    Store->>Encrypt: EncryptValue(newValue, password)
    Encrypt-->>Store: newEncryptedValue
    Store->>DB: UpdateRecord(token, newEncryptedValue)
    DB-->>Store: updatedRecord
    Store-->>App: success
```

### Update Validation

1. **Token Existence**: Verify token exists
2. **Current Password**: Validate current password if required
3. **New Encryption**: Encrypt new value with password
4. **Timestamp Update**: Update `updated_at` timestamp
5. **Atomic Update**: Perform atomic database update

## Delete Operations Flow

### Soft Delete Process

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant DB as Database
    
    App->>Store: TokenSoftDelete(token)
    Store->>DB: FindByToken(token)
    DB-->>Store: record
    Store->>DB: UpdateSoftDeletedAt(token, now)
    DB-->>Store: softDeletedRecord
    Store-->>App: success
```

### Hard Delete Process

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant DB as Database
    
    App->>Store: TokenDelete(token)
    Store->>DB: DeleteByToken(token)
    DB-->>Store: deletionResult
    Store-->>App: success
```

### Delete Strategies

| Operation | Recovery | Data Removal | Use Case |
|-----------|----------|--------------|----------|
| Soft Delete | Yes | Logical | Temporary deletion, recovery needed |
| Hard Delete | No | Physical | Permanent removal, compliance |

## Batch Operations Flow

### Expired Token Cleanup

```mermaid
graph TB
    A[Cleanup Task] --> B[Find Expired Tokens]
    B --> C{Delete Type}
    C -->|Soft Delete| D[Soft Delete Expired]
    C -->|Hard Delete| E[Hard Delete Expired]
    D --> F[Update soft_deleted_at]
    E --> G[Remove Records]
    F --> H[Return Count]
    G --> H
```

### Batch Read Operations

```go
// Multiple token read
values, err := vault.TokensRead(ctx, []string{"token1", "token2"}, "password")
```

1. **Token Validation**: Validate all tokens exist
2. **Batch Decryption**: Decrypt all values efficiently
3. **Result Mapping**: Map tokens to values
4. **Error Aggregation**: Collect and return errors

## Data Integrity Flow

### Validation Checks

```mermaid
graph TB
    A[Input Data] --> B[Format Validation]
    B --> C[Size Validation]
    C --> D[Character Validation]
    D --> E[Business Rules]
    E --> F[Database Constraints]
    F --> G[Success]
    
    B --> H[Format Error]
    C --> I[Size Error]
    D --> J[Character Error]
    E --> K[Business Rule Error]
    F --> L[Constraint Error]
```

### Integrity Measures

1. **Input Validation**: Validate all input parameters
2. **Database Constraints**: Use database constraints for data integrity
3. **Transaction Safety**: Use database transactions for atomic operations
4. **Error Handling**: Comprehensive error handling and logging

## Performance Considerations

### Database Optimization

1. **Indexing**: Proper indexes on frequently queried fields
2. **Query Optimization**: Efficient SQL generation
3. **Connection Pooling**: Leverage database connection pooling
4. **Batch Operations**: Minimize database round trips

### Memory Management

1. **Lazy Loading**: Load data only when needed
2. **Efficient Encryption**: Minimize memory overhead
3. **Resource Cleanup**: Proper resource disposal
4. **Stream Processing**: Process large datasets efficiently

## Monitoring and Observability

### Key Metrics

- **Token Creation Rate**: Number of tokens created per time period
- **Query Performance**: Average query response time
- **Error Rates**: Frequency of different error types
- **Database Performance**: Connection pool usage, query times

### Logging Points

- **Token Operations**: Create, read, update, delete operations
- **Database Operations**: SQL queries, connection events
- **Security Events**: Failed authentications, suspicious activities
- **Performance Events**: Slow queries, resource usage

## Identity-Based Data Flows

These flows apply when `PasswordIdentityEnabled` is set to `true`.

### Token Creation with Identity

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant MetaDB as Metadata DB
    participant VaultDB as Vault DB
    participant Encrypt as Encryption
    
    App->>Store: TokenCreate(value, password, length)
    Store->>MetaDB: FindIdentityID(password)
    alt Identity Exists
        MetaDB-->>Store: passwordID
    else Identity Not Found
        Store->>Encrypt: HashPassword(password)
        Encrypt-->>Store: passwordHash
        Store->>MetaDB: CreateIdentity(passwordHash)
        MetaDB-->>Store: passwordID
    end
    Store->>Encrypt: EncryptValue(value, password)
    Encrypt-->>Store: encryptedValue
    Store->>VaultDB: CreateRecord(token, encryptedValue)
    VaultDB-->>Store: record
    Store->>MetaDB: LinkRecord(record.ID, passwordID)
    Store-->>App: token
```

### Bulk Rekey Flow (Fast Path)

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant MetaDB as Metadata DB
    participant VaultDB as Vault DB
    participant Encrypt as Encryption
    
    App->>Store: BulkRekey(oldPass, newPass)
    Store->>MetaDB: FindIdentityID(oldPass)
    MetaDB-->>Store: oldPasswordID
    Store->>MetaDB: FindRecordsByIdentity(oldPasswordID)
    MetaDB-->>Store: recordIDs[]
    loop For Each Record
        Store->>VaultDB: GetRecord(recordID)
        VaultDB-->>Store: record
        Store->>Encrypt: Decrypt(record.value, oldPass)
        Encrypt-->>Store: value
        Store->>Encrypt: Encrypt(value, newPass)
        Encrypt-->>Store: newEncryptedValue
        Store->>VaultDB: UpdateRecord(recordID, newEncryptedValue)
    end
    Store->>Encrypt: HashPassword(newPass)
    Encrypt-->>Store: newPasswordHash
    Store->>MetaDB: UpdateIdentity(oldPasswordID, newPasswordHash)
    Store-->>App: count
```

### Migration Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Store as VaultStore
    participant VaultDB as Vault DB
    participant Encrypt as Encryption
    participant MetaDB as Metadata DB
    
    App->>Store: MigrateRecordLinks(password)
    Store->>VaultDB: GetAllRecords()
    VaultDB-->>Store: records[]
    loop For Each Record
        Store->>Encrypt: TryDecrypt(record, password)
        alt Decryption Success
            Encrypt-->>Store: value
            Store->>MetaDB: FindOrCreateIdentity(password)
            MetaDB-->>Store: passwordID
            Store->>MetaDB: LinkRecord(record.ID, passwordID)
        else Decryption Failed
            Encrypt-->>Store: error
        end
    end
    Store-->>App: migratedCount
```

## See Also

- [Architecture](architecture.md) - System design and patterns
- [API Reference](api_reference.md) - Complete API documentation
- [Token Operations](modules/token_operations.md) - Token-specific operations
- [Query Interface](modules/query_interface.md) - Query system details
- [Password Identity Management](modules/password_identity_management.md) - Identity-based password management

## Changelog

- **v1.1.0** (2026-02-03): Added identity-based data flow diagrams for token creation, bulk rekey, and migration operations.
- **v1.0.0** (2026-02-03): Initial data flow documentation
