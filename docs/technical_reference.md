# Technical Reference

This document provides a detailed technical reference for the VaultStore implementation.

## Database Schema

VaultStore uses a single table to store records. The table name is configurable when creating a new store.

The table has the following columns:

| Column | Type | Description |
|--------|------|-------------|
| id | String | Primary key, a unique identifier for the record (Human Friendly UUID) |
| vault_token | Long Text | Unique token used to access the secret |
| vault_value | Long Text | The encrypted secret value |
| created_at | DateTime | Timestamp when the record was created |
| updated_at | DateTime | Timestamp when the record was last updated |
| soft_deleted_at | DateTime | Timestamp when the record was soft deleted (MAX_DATE if not deleted) |

### Metadata Table

When `PasswordIdentityEnabled` is true, a metadata table is used to track password identities:

| Column | Type | Description |
|--------|------|-------------|
| id | Integer | Primary key, auto-increment |
| object_type | String | Type: 'password_identity', 'record', or 'vault' |
| object_id | String | Unique identifier for the object |
| meta_key | String | Key for the metadata entry |
| meta_value | String | Value of the metadata entry |
| created_at | DateTime | Timestamp when the entry was created |
| updated_at | DateTime | Timestamp when the entry was last updated |

## Record Structure

The `Record` struct is defined as follows:

```go
type Record struct {
    dataobject.DataObject
}
```

It embeds `dataobject.DataObject` for storing data. The struct provides methods for accessing and modifying the record's fields:

- `ID()` / `SetID(id string)`: Get/set the record's ID
- `Token()` / `SetToken(token string)`: Get/set the record's token
- `Value()` / `SetValue(value string)`: Get/set the record's encrypted value
- `CreatedAt()` / `SetCreatedAt(createdAt string)`: Get/set the record's creation timestamp
- `UpdatedAt()` / `SetUpdatedAt(updatedAt string)`: Get/set the record's update timestamp
- `SoftDeletedAt()` / `SetSoftDeletedAt(softDeletedAt string)`: Get/set the record's soft deletion timestamp

## Query Interface

The `RecordQueryInterface` provides a flexible way to search and filter records in the VaultStore. It is implemented by the `recordQueryImpl` struct and offers methods for building complex queries.

Key features of the query interface:

- Filtering by ID or token (single or multiple values)
- Pagination with limit and offset
- Sorting with order by and sort direction
- Option to include soft-deleted records with `SetSoftDeletedInclude(true)`
- Option to retrieve only soft-deleted records with `SetSoftDeletedOnly(true)`
- Option to only count records

For detailed usage examples, see the [Query Interface documentation](./query_interface.md).

## Store Implementation

The `Store` struct implements the `StoreInterface` and provides methods for interacting with the database.

### Initialization

The store is initialized with the `NewStore` function, which takes a `NewStoreOptions` struct:

```go
type NewStoreOptions struct {
    VaultTableName          string
    VaultMetaTableName      string
    DB                      *sql.DB
    DbDriverName            string
    AutomigrateEnabled      bool
    DebugEnabled            bool
    CryptoConfig            *CryptoConfig
    PasswordIdentityEnabled bool
}
```

### Auto-Migration

If `AutomigrateEnabled` is set to `true`, the store will automatically create the necessary table in the database if it doesn't exist.

### Encryption and Decryption

VaultStore uses **AES-256-GCM** encryption with **Argon2id** key derivation for protecting secret values. The encryption is password-based, meaning you need the correct password to decrypt and access the secret value.

### CryptoConfig

You can configure the cryptographic parameters via `CryptoConfig`:

```go
// Default configuration (balanced)
CryptoConfig: vaultstore.DefaultCryptoConfig()

// High security configuration
CryptoConfig: vaultstore.HighSecurityCryptoConfig()

// Lightweight configuration (resource-constrained)
CryptoConfig: vaultstore.LightweightCryptoConfig()
```

The `encode` function encrypts a value with a password:

```go
func encode(value string, password string) (string, error)
```

The `decode` function decrypts a value with a password:

```go
func decode(value string, password string) (string, error)
```

### Token Generation

Tokens are generated using the `GenerateToken` function, which creates a random string of the specified length:

```go
func GenerateToken(length int) string
```

## Error Handling

VaultStore returns errors for various scenarios:

- Database errors
- Token not found
- Decryption errors (e.g., wrong password)
- Invalid parameters

It's important to check for errors when calling VaultStore methods.

## Thread Safety

VaultStore is designed to be thread-safe. It uses database transactions to ensure data consistency when multiple goroutines access the store simultaneously.

## Performance Considerations

VaultStore is designed for secure storage of secrets, not for high-performance data access. The encryption and decryption operations can be CPU-intensive, especially for large values.

For better performance, consider:

- Caching frequently accessed secrets
- Using shorter tokens (but not too short to compromise security)
- Optimizing database access (e.g., using indexes)

## Security Considerations

- Store passwords securely; they are used to encrypt and decrypt secrets
- Use HTTPS when transmitting tokens and passwords
- Implement proper access control to restrict who can create, read, update, and delete secrets
- Consider using a secure key management system for storing encryption keys
