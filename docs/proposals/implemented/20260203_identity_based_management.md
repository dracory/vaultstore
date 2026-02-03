# Proposal: Identity-Based Password Management (Single Table)

*Date: February 3, 2026*
*Replaces: 20260203_identity_based_management.md*

## Overview

This proposal implements a "Password Identity" model using a **Single Generic Metadata Table**. 

Instead of creating separate tables for identities and metadata, we leverage the flexible `vault_meta` structure to store **both** the password identities themselves and the links between records and those identities.

## Core Concept

We assume the number of unique passwords in the vault is small (< 100). This allows us to find a password identity by iterating through the stored hashes ("Try-and-Verify") rather than using a complex Blind Index or Peppering system.

## Schema Changes

### 1. `vault_meta` Table

We use a single, generic table for all metadata needs.

```go
type VaultMeta struct {
    ID         uint   `gorm:"primaryKey"`
    ObjectType string `gorm:"index:idx_obj_key;size:50;column:object_type"` 
    ObjectID   string `gorm:"index:idx_obj_key;size:64;column:object_id"`   
    Key        string `gorm:"index:idx_obj_key;size:50;column:meta_key"`    
    Value      string `gorm:"type:text;column:meta_value"`                 
}
```

### 2. Data Structures

We utilize this table for two distinct types of data:

#### A. Password Identity Objects
Stores the verification hash for a specific password.

| Column | Value | Description |
| :--- | :--- | :--- |
| **ObjectType** | `password_identity` | Defines this row as a Password Identity |
| **ObjectID** | `p_<uuid>` | Unique ID for this password (the "Password ID") |
| **Key** | `hash` | Indicates this value is the verification hash |
| **Value** | `$2y$10$...` | The bcrypt (or Argon2) hash of the password |

#### B. Record Links
Links a generic record to a password identity.

| Column | Value | Description |
| :--- | :--- | :--- |
| **ObjectType** | `record` | Defines this row as metadata for a specific Record |
| **ObjectID** | `r_<uuid>` | The Record ID |
| **Key** | `password_id` | The reference key |
| **Value** | `p_<uuid>` | The ID of the Password Identity (Foreign Key-ish) |

#### C. Global Vault Settings
Stores global configuration or state, such as versioning, to assist with migration and compatibility.

| Column | Value | Description |
| :--- | :--- | :--- |
| **ObjectType** | `vault` | Defines this row as a Global Vault Setting |
| **ObjectID** | `settings` | Singleton ID for global settings |
| **Key** | `version` | The configuration key |
| **Value** | `1.1` | The current vault version |

## Configuration

To maintain strict backward compatibility and allow users to opt-in to metadata storage features, a new flag is added to `NewStoreOptions`:

```go
type NewStoreOptions struct {
    // ... existing fields ...

    // PasswordIdentityEnabled enables the identity-based password management 
    // feature.
    PasswordIdentityEnabled bool
}
```

- **Schema:** The `vault_meta` table is **always** created by `AutoMigrate()`, regardless of this setting. This ensures the generic metadata capability is always available for future features or custom use cases.
- **Feature Logic:** The `PasswordIdentityEnabled` flag specifically controls the **Password Identity** workflow (creating identity objects and linking records to them).
    - If `false` (default): No password identity metadata is created/updated.
    - If `true`: The system actively manages password identities and links as described below.

## Workflow

The specific workflows for **Password Identity** below are only active if `PasswordIdentityEnabled` is set to `true`.

### 1. Finding an Identity (The "Scan")

To find if a password already exists in the system:

```go
func (s *store) findIdentityID(password string) (string, error) {
    // 1. Fetch all identity hashes
    // SELECT object_id, meta_value FROM vault_meta WHERE object_type = 'password_identity' AND meta_key = 'hash'
    type IdentityRow struct {
        ObjectID string
        Value    string
    }
    var rows []IdentityRow
    s.db.Table("vault_meta").
        Where("object_type = ? AND meta_key = ?", "password_identity", "hash").
        Scan(&rows)

    // 2. Iterate and Verify
    for _, row := range rows {
        // Standard bcrypt check (slow but safe)
        if VerifyPassword(password, row.Value) { 
            return row.ObjectID, nil
        }
    }
    return "", ErrNotFound
}
```

### 2. Writing a Record (Set)

When `Set(token, value, password)` is called:

1.  **Encrypt** value (standard v2).
2.  **Resolve Identity:**
    *   Call `findIdentityID(password)`.
    *   **If found:** use existing `passwordID`.
    *   **If not found:** 
        *   Generate new `passwordID` (e.g., `uuid`).
        *   Compute `hash = bcrypt(password)`.
        *   Insert into `vault_meta`: `(ObjectType="password_identity", ObjectID=passwordID, Key="hash", Value=hash)`.
3.  **Link Record:**
    *   Save the encrypted record to `vault` table.
    *   Upsert into `vault_meta`: `(ObjectType="record", ObjectID=record.ID, Key="password_id", Value=passwordID)`.

### 3. Bulk Rekeying

```go
func (s *store) BulkRekey(oldPass, newPass string) (int, error) {
    // Optimization: If Identity feature is enabled, use metadata lookup
    if s.passwordIdentityEnabled {
        return s.bulkRekeyFast(oldPass, newPass)
    }
    
    // Fallback: Scan-and-Test (Slower)
    // If metadata is disabled, we must try to decrypt every record.
    return s.bulkRekeyScan(oldPass, newPass)
}
```

#### A. Fast Path (Identity Enabled)
1.  Find Identity for `oldPass` via metadata scan.
2.  Query `vault_meta` for all records linked to this Identity ID.
3.  Load, Decrypt, and Re-encrypt only those specific records.
4.  Update metadata links to the new password identity.

#### B. Fallback Path (Identity Disabled)
If the feature is disabled, `BulkRekey` still functions but is computationally more expensive.
1.  Iterate through **all** records in the `vault` table.
2.  For each record:
    *   Attempt `Decrypt(value, oldPass)`.
    *   If decryption fails (wrong password): Skip.
    *   If decryption succeeds:
        *   `Encrypt(value, newPass)`.
        *   `Update(record)`.
3.  **Performance Warning:** Since v2 encryption uses `Argon2id` (memory-hard), attempting to decrypt thousands of records with the "wrong" password will be slow. This is acceptable for maintenance tasks but should be documented.

## Pros & Cons

### Pros

1.  **Flexibility:** The generic schema (`ObjectType`, `ObjectID`, `Key`, `Value`) allows adding new feature support (e.g., expiry, access control, tags, settings) without *any* future SQL schema changes.
2.  **Operational Simplicity:** Only one new table (`vault_meta`) to manage, backup, and monitor. Eliminates the complexity of managing multiple specialized tables for every new feature.
3.  **Backward Compatibility:** The design does not modify the core `vault` table, ensuring existing code continues to function (albeit without metadata features) without breaking changes.
4.  **Standard Security:** Uses standard, salted hashing (bcrypt/Argon2) for password verification. Unlike "Blind Indexing", it does not require managing a secret "pepper" key in the application configuration.
5.  **Smart Migration:** The inclusion of `Global Vault Settings` (versioning) allows for optimized migration paths, enabling the system to skip expensive checks once the vault is marked as up-to-date.

### Cons

1.  **Performance (Identity Scan):** Identifying a password requires fetching *all* identity hashes and verifying them one-by-one.
    *   *Mitigation:* This design assumes a "Small N" scenario (records < millions, unique passwords < 100). For this scale, the overhead is negligible.
2.  **Data Redundancy:** Storing string keys (e.g., "object_type", "meta_key") for every row increases storage size compared to normalized tables with integer IDs.
    *   *Mitigation:* Modern storage is cheap, and compression handles repeated strings well.
3.  **Referential Integrity:** The generic `ObjectID` prevents the use of strict SQL `FOREIGN KEY` constraints (e.g., `ON DELETE CASCADE`).
    *   *Mitigation:* Application logic must strictly handle cleanup (e.g., deleting a record must explicitly delete its metadata).
4.  **Type Safety:** The `Value` column is text-based. Storing non-string data (booleans, integers, dates) requires application-level parsing and casting.

## Migration Strategy

### 1. Versioning
We use the global setting `vault/settings/version` to track the vault's state.
- **Version < 1.1 (or missing):** Indicates the vault contains records that may not have "Password Identities" linked.
- **Version >= 1.1:** Indicates the vault is fully migrated (or empty/new).

### 2. Migration Logic
Existing records have no metadata.
- **On-Access:** When `Get(token, pass)` succeeds:
  1. Check if the record has a linked `password_id`.
  2. If missing, run the "Link Logic" (Find/Create Identity -> Link Record).
- **Batch Migration Tool:** A function `Migrate(password)` that:
  1. Scans all records.
  2. Tries to decrypt each with `password`.
  3. Links successful ones to the identity.
  4. Once all records are linked, updates `vault/settings/version` to `1.1`.