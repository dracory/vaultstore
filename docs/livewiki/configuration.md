---
path: configuration.md
page-type: reference
summary: Configuration options, environment variables, and setup parameters including CryptoConfig and Password Identity settings.
tags: [configuration, setup, options, environment, cryptoconfig, password-identity]
created: 2026-02-03
updated: 2026-02-03
version: 1.1.0
---

# Configuration

This document covers all configuration options for VaultStore, including initialization parameters, environment variables, and runtime settings.

## Store Configuration

### NewStoreOptions

The main configuration structure for creating a VaultStore instance.

```go
type NewStoreOptions struct {
    // Required: Name of the vault table in the database
    VaultTableName string
    
    // Required: Name of the vault metadata table
    VaultMetaTableName string
    
    // Required: Database connection
    DB *sql.DB
    
    // Optional: Database driver name (auto-detected if not provided)
    DbDriverName string
    
    // Optional: Enable automatic database migration (default: false)
    AutomigrateEnabled bool
    
    // Optional: Enable debug logging (default: false)
    DebugEnabled bool
    
    // Optional: Database driver name (auto-detected if not provided)
    DbDriverName string
}
```

### Required Parameters

#### VaultTableName

The name of the table where vault data will be stored.

```go
// Example
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "app_vault",  // Table name
    DB:           db,
})
```

**Considerations:**
- Must be a valid SQL table name
- Should be unique per application
- Can include prefixes for multi-tenant applications

#### DB

The database connection instance.

```go
// SQLite example
db, err := sql.Open("sqlite", "./app.db")

// PostgreSQL example
db, err := sql.Open("postgres", "postgres://user:pass@localhost/db")

// MySQL example
db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/db")
```

**Requirements:**
- Must be a valid `*sql.DB` instance
- Should have appropriate permissions for table operations
- Connection pooling should be configured appropriately

#### VaultMetaTableName

The name of the metadata table for vault operations. This table stores password identity information and vault settings when PasswordIdentityEnabled is true.

```go
// Example
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "app_vault",
    VaultMetaTableName: "app_vault_meta",  // Metadata table name
    DB:                 db,
})
```

**Considerations:**
- Must be a valid SQL table name
- Should be different from VaultTableName
- Used for password identity management and vault versioning

### Optional Parameters

#### AutomigrateEnabled

Controls whether VaultStore automatically creates and updates the database schema.

```go
// Enable auto-migration
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:     "vault",
    DB:                 db,
    AutomigrateEnabled: true,  // Auto-create tables
})
```

**Behavior:**
- `true`: Automatically creates vault table if it doesn't exist
- `false`: Requires manual table creation (see Manual Schema)

#### DebugEnabled

Enables detailed logging for debugging purposes.

```go
// Enable debug mode
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    DebugEnabled:   true,  // Enable debug logging
})
```

**Debug Output:**
- SQL query logging
- Error stack traces
- Performance metrics
- Encryption/decryption operations

#### CryptoConfig

Custom cryptographic configuration for encryption operations. Allows tuning Argon2id and AES-GCM parameters.

```go
// Use default configuration
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    CryptoConfig:   vaultstore.DefaultCryptoConfig(),
})

// High security configuration
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    CryptoConfig:   vaultstore.HighSecurityCryptoConfig(),
})

// Lightweight configuration for resource-constrained environments
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    CryptoConfig:   vaultstore.LightweightCryptoConfig(),
})
```

**Configuration Options:**

```go
type CryptoConfig struct {
    // Argon2id parameters
    Iterations  int  // Number of iterations (default: 3)
    Memory      int  // Memory in bytes (default: 64MB)
    Parallelism int  // Parallelism (default: 4)
    KeyLength   int  // Key length in bytes (default: 32)

    // AES-GCM parameters
    SaltSize  int  // Salt size in bytes (default: 16)
    NonceSize int  // Nonce size in bytes (default: 12)
    TagSize   int  // Tag size in bytes (default: 16)
}
```

**Pre-configured Profiles:**

| Profile | Iterations | Memory | Use Case |
|---------|-----------|--------|----------|
| `DefaultCryptoConfig()` | 3 | 64MB | Balanced security and performance |
| `HighSecurityCryptoConfig()` | 4 | 128MB | Maximum security requirements |
| `LightweightCryptoConfig()` | 2 | 32MB | Resource-constrained environments |

#### PasswordIdentityEnabled

Enables identity-based password management for optimized bulk rekey operations. When enabled, the system tracks which records share the same password, allowing efficient bulk password changes.

```go
// Enable password identity management
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:          "vault",
    VaultMetaTableName:      "vault_meta",
    DB:                      db,
    PasswordIdentityEnabled: true,  // Enable identity management
})
```

**Benefits:**
- **Fast Bulk Rekey**: O(1) password change vs O(n) scan-and-test
- **Password Deduplication**: Automatically groups records by password
- **Efficient Migration**: Metadata links enable quick re-encryption

**Considerations:**
- Requires VaultMetaTableName to be set
- Metadata table stores password hash references (not actual passwords)
- Slightly higher storage overhead for metadata
- Recommended for applications with frequent bulk rekey operations

**Migration:**
Existing records can be migrated to use identity management via `MigrateRecordLinks()` method.

#### DbDriverName

Explicitly specify the database driver type.

```go
// Explicit driver specification
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB:             db,
    DbDriverName:   "sqlite",  // Explicit driver
})
```

**Supported Drivers:**
- `sqlite`
- `postgres`
- `mysql`
- `mssql` (experimental)

## Token Configuration

### TokenCreateOptions

Options for customizing token creation behavior.

```go
type TokenCreateOptions struct {
    // Optional: Token expiration time
    ExpiresAt time.Time
}
```

#### Token Length

Token length is specified during creation:

```go
// Create 32-character token
token, err := vault.TokenCreate(ctx, "value", "password", 32)

// Create 64-character token
token, err := vault.TokenCreate(ctx, "value", "password", 64)
```

**Recommended Lengths:**
- **16 characters**: Basic security, development use
- **32 characters**: Good security, most applications
- **64 characters**: High security, sensitive applications

#### Token Expiration

Set expiration time for tokens:

```go
// Expire in 24 hours
token, err := vault.TokenCreate(ctx, "value", "password", 32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Now().Add(24 * time.Hour),
    })

// Expire at specific time
token, err := vault.TokenCreate(ctx, "value", "password", 32,
    vaultstore.TokenCreateOptions{
        ExpiresAt: time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
    })
```

## Database Configuration

### Supported Databases

#### SQLite

**Configuration:**
```go
import (
    "database/sql"
    _ "github.com/glebarez/sqlite"
)

db, err := sql.Open("sqlite", "./vault.db")
```

**Connection String Options:**
- `./vault.db` - File-based database
- `:memory:` - In-memory database (for testing)
- `file:vault.db?cache=shared` - Shared cache mode

**Recommended Settings:**
```go
db.SetMaxOpenConns(1)  // SQLite recommends single connection
db.SetMaxIdleConns(1)
```

#### PostgreSQL

**Configuration:**
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/vaultdb?sslmode=disable")
```

**Connection String Parameters:**
- `host`: Database host (default: localhost)
- `port`: Database port (default: 5432)
- `user`: Database username
- `password`: Database password
- `dbname`: Database name
- `sslmode`: SSL mode (disable, require, verify-ca, verify-full)

**Recommended Settings:**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

#### MySQL

**Configuration:**
```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
)

db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/vaultdb")
```

**Connection String Format:**
```
username:password@protocol(address)/dbname?parameter=value
```

**Recommended Settings:**
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Connection Pooling

Configure database connection pooling for optimal performance:

```go
// General recommendation for production
db.SetMaxOpenConns(25)           // Maximum open connections
db.SetMaxIdleConns(5)            // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
db.SetConnMaxIdleTime(2 * time.Minute)   // Idle connection timeout
```

## Environment Variables

While VaultStore doesn't directly use environment variables, you can use them for configuration:

### Database Configuration

```bash
# SQLite
export VAULT_DB_PATH="./data/vault.db"

# PostgreSQL
export VAULT_DB_HOST="localhost"
export VAULT_DB_PORT="5432"
export VAULT_DB_USER="vaultuser"
export VAULT_DB_PASSWORD="vaultpass"
export VAULT_DB_NAME="vaultdb"

# MySQL
export VAULT_DB_HOST="localhost"
export VAULT_DB_PORT="3306"
export VAULT_DB_USER="vaultuser"
export VAULT_DB_PASSWORD="vaultpass"
export VAULT_DB_NAME="vaultdb"
```

### Application Configuration

```bash
# Vault settings
export VAULT_TABLE_NAME="app_vault"
export VAULT_DEBUG="true"
export VAULT_AUTO_MIGRATE="true"

# Token settings
export VAULT_DEFAULT_TOKEN_LENGTH="32"
export VAULT_DEFAULT_EXPIRY_HOURS="24"
```

### Example Usage

```go
func getDBConfig() *sql.DB {
    dbType := os.Getenv("VAULT_DB_TYPE")
    
    switch dbType {
    case "sqlite":
        dbPath := getEnvOrDefault("VAULT_DB_PATH", "./vault.db")
        db, _ := sql.Open("sqlite", dbPath)
        return db
    case "postgres":
        connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
            os.Getenv("VAULT_DB_USER"),
            os.Getenv("VAULT_DB_PASSWORD"),
            getEnvOrDefault("VAULT_DB_HOST", "localhost"),
            getEnvOrDefault("VAULT_DB_PORT", "5432"),
            os.Getenv("VAULT_DB_NAME"))
        db, _ := sql.Open("postgres", connStr)
        return db
    default:
        panic("Unsupported database type")
    }
}

func getVaultStore(db *sql.DB) (*vaultstore.StoreInterface, error) {
    tableName := getEnvOrDefault("VAULT_TABLE_NAME", "vault")
    debug := getEnvOrDefault("VAULT_DEBUG", "false") == "true"
    autoMigrate := getEnvOrDefault("VAULT_AUTO_MIGRATE", "false") == "true"
    
    return vaultstore.NewStore(vaultstore.NewStoreOptions{
        VaultTableName:     tableName,
        DB:                 db,
        DebugEnabled:       debug,
        AutomigrateEnabled: autoMigrate,
    })
}
```

## Manual Schema

If you disable auto-migration, you can create the schema manually:

### SQLite Schema

```sql
CREATE TABLE IF NOT EXISTS vault (
    id              TEXT PRIMARY KEY,
    token           TEXT UNIQUE NOT NULL,
    value           TEXT NOT NULL,
    password_hash   TEXT,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL,
    expires_at      TEXT,
    soft_deleted_at TEXT,
    data            TEXT
);

CREATE INDEX IF NOT EXISTS idx_vault_token ON vault(token);
CREATE INDEX IF NOT EXISTS idx_vault_created_at ON vault(created_at);
CREATE INDEX IF NOT EXISTS idx_vault_expires_at ON vault(expires_at);
CREATE INDEX IF NOT EXISTS idx_vault_soft_deleted_at ON vault(soft_deleted_at);

-- Metadata table for password identity management and vault settings
CREATE TABLE IF NOT EXISTS vault_meta (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    object_type  TEXT NOT NULL,
    object_id    TEXT NOT NULL,
    meta_key     TEXT NOT NULL,
    meta_value   TEXT,
    created_at   TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_vault_meta_obj ON vault_meta(object_type, object_id, meta_key);
```

### PostgreSQL Schema

```sql
CREATE TABLE IF NOT EXISTS vault (
    id              TEXT PRIMARY KEY,
    token           TEXT UNIQUE NOT NULL,
    value           TEXT NOT NULL,
    password_hash   TEXT,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL,
    expires_at      TIMESTAMP,
    soft_deleted_at TIMESTAMP,
    data            JSONB
);

CREATE INDEX IF NOT EXISTS idx_vault_token ON vault(token);
CREATE INDEX IF NOT EXISTS idx_vault_created_at ON vault(created_at);
CREATE INDEX IF NOT EXISTS idx_vault_expires_at ON vault(expires_at);
CREATE INDEX IF NOT EXISTS idx_vault_soft_deleted_at ON vault(soft_deleted_at);

-- Metadata table for password identity management and vault settings
CREATE TABLE IF NOT EXISTS vault_meta (
    id           SERIAL PRIMARY KEY,
    object_type  VARCHAR(50) NOT NULL,
    object_id    VARCHAR(64) NOT NULL,
    meta_key     VARCHAR(50) NOT NULL,
    meta_value   TEXT,
    created_at   TIMESTAMP NOT NULL,
    updated_at   TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_vault_meta_obj ON vault_meta(object_type, object_id, meta_key);
```

```sql
CREATE TABLE IF NOT EXISTS vault (
    id              VARCHAR(255) PRIMARY KEY,
    token           VARCHAR(255) UNIQUE NOT NULL,
    value           TEXT NOT NULL,
    password_hash   VARCHAR(255),
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    expires_at      DATETIME,
    soft_deleted_at DATETIME,
    data            JSON
);

CREATE INDEX idx_vault_token ON vault(token);
CREATE INDEX idx_vault_created_at ON vault(created_at);
CREATE INDEX idx_vault_expires_at ON vault(expires_at);
CREATE INDEX idx_vault_soft_deleted_at ON vault(soft_deleted_at);

-- Metadata table for password identity management and vault settings
CREATE TABLE IF NOT EXISTS vault_meta (
    id           INT AUTO_INCREMENT PRIMARY KEY,
    object_type  VARCHAR(50) NOT NULL,
    object_id    VARCHAR(64) NOT NULL,
    meta_key     VARCHAR(50) NOT NULL,
    meta_value   TEXT,
    created_at   DATETIME NOT NULL,
    updated_at   DATETIME NOT NULL
);

CREATE INDEX idx_vault_meta_obj ON vault_meta(object_type, object_id, meta_key);
```

## Performance Tuning

### Database Optimization

#### SQLite

```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = NORMAL;
PRAGMA cache_size = 1000;
PRAGMA temp_store = memory;
```

#### PostgreSQL

```sql
-- Connection pool settings in postgresql.conf
max_connections = 100
shared_buffers = 256MB
effective_cache_size = 1GB
```

#### MySQL

```sql
-- Settings in my.cnf
innodb_buffer_pool_size = 256M
innodb_log_file_size = 64M
max_connections = 100
```

### Application Tuning

```go
// Optimize for high-throughput applications
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName: "vault",
    DB: db,
    // Configure database for performance
})

// Use context with timeout for production
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

token, err := vault.TokenCreate(ctx, value, password, length)
```

## Security Configuration

### Encryption Settings

VaultStore uses AES-256-GCM encryption by default. The encryption is configured internally but you should consider:

1. **Password Strength**: Use strong passwords for token protection
2. **Token Length**: Use appropriate token lengths (32+ characters recommended)
3. **Database Security**: Secure database connections and access
4. **Environment Security**: Protect environment variables and secrets

### Access Control

```go
// Example: Wrap VaultStore with access control
type SecureVault struct {
    vault *vaultstore.StoreInterface
    permissions map[string]bool
}

func (sv *SecureVault) TokenCreate(ctx context.Context, value, password string, length int) (string, error) {
    if !sv.permissions["create"] {
        return "", errors.New("permission denied")
    }
    return sv.vault.TokenCreate(ctx, value, password, length)
}
```

## See Also

- [Getting Started](getting_started.md) - Setup and installation
- [Architecture](architecture.md) - System design overview
- [API Reference](api_reference.md) - Complete API documentation
- [Database Setup](../data_stores.md) - Database-specific information
- [Password Identity Management](modules/password_identity_management.md) - Identity-based password management

## Changelog

- **v1.1.0** (2026-02-03): Added documentation for CryptoConfig, PasswordIdentityEnabled, and VaultMetaTableName options. Added vault_meta table schema for all databases.
- **v1.0.0** (2026-02-03): Initial configuration documentation

