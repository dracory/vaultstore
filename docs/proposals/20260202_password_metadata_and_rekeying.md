# Password Metadata and Bulk Rekeying

*Proposal Date: February 2, 2026*

## Status: Draft

## Overview

This proposal adds password metadata storage and bulk rekeying capabilities to VaultStore, building on the already-implemented v2 encryption with Argon2id.

## Context

The versioned encryption feature (implemented in `20250202_versioned_encryption.md`) already provides:
- ✅ Argon2id KDF for v2 encryption
- ✅ Per-record random salt embedded in ciphertext

What remains to be implemented:
- ❌ Password hash column for password identification
- ❌ Bulk rekeying by password identity

## Proposed Changes

### 1. Schema Changes

Create a separate metadata table for flexible key-value storage:

```go
type vaultMeta struct {
    RecordID string `gorm:"index;column:record_id"`
    Key      string `gorm:"index;column:key"`
    Value    string `gorm:"column:value"`
}
```

**Why separate meta table?**
- More flexible (can add any metadata later)
- Doesn't bloat main table
- No joins needed (use `WHERE id IN (...)`)
- Indexed on `(key, value)` for fast lookups

**Password metadata stored as:**
- Key: `password_hash`
- Value: bcrypt hash of the password

### 2. Password Verification

Add a helper to verify a password without full decryption:

```go
// Verify that a password matches the stored hash for a token
VerifyPassword(ctx context.Context, token string, password string) (bool, error)
```

Implementation:
- Load the record by token
- Use `bcrypt.CompareHashAndPassword()` to verify
- Returns `true` if password matches, `false` otherwise

### 3. Record Identification by Password Identity

Provide a way to list records by password identity:
- Derive the bcrypt hash for a given password
- Query meta table for matching records
- Fetch records by IDs

```go
// Get bcrypt hash for password
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Get record IDs from meta table
var recordIDs []string
db.Table("vault_meta").
    Where("key = ? AND value = ?", "password_hash", hash).
    Pluck("record_id", &recordIDs)

// Get records by IDs
records := db.Table("vault").Where("id IN ?", recordIDs).Find(...)
```

### 4. Bulk Rekeying

Implement a bulk rekey workflow:

```go
// Bulk rekey all records associated with a given password
BulkRekey(ctx context.Context, oldPassword string, newPassword string) (count int, err error)
```

Implementation:
- Identify all records with matching `password_hash` from meta table
- For each record:
  - Decrypt with `oldPassword`
  - Re-encrypt with `newPassword`
  - Generate and store new bcrypt hash for `newPassword` in meta table
- Run in batches to avoid long transactions
- Allow caller to control transaction boundaries

## Implementation Details

- Modify the database schema to add `password_hash` column
- Update the record implementation to store and expose password hash
- Add bcrypt helpers for password hashing and verification
- Implement password verification, password-based record selection, and rekey helpers
- Ensure new field is included in soft delete and migration flows

## Migration Strategy

Existing records lack password metadata. Migrate incrementally:

### Schema Migration

Add new column with non-null default indicating "no password metadata yet":
- `password_hash` = empty string

### On-Access Enrichment

When a record is successfully decrypted with a password:
- Compute bcrypt hash and store it
- Frequently used records will be migrated automatically over time

### Offline Migration Tool (Optional)

Provide a CLI or helper that iterates over tokens, prompts for passwords, and backfills metadata.

## Benefits

- **Operational Efficiency**: Enables bulk operations on secrets sharing the same password identity
- **Key Rotation**: Simplifies regular key rotation as a security best practice
- **Breach Response**: Provides a mechanism to quickly update all affected secrets if a password is compromised
- **Password Verification**: Allows checking password correctness without full decryption overhead

## Risks and Mitigations

- **Additional Sensitive Metadata**: bcrypt hashes are one-way but still sensitive. Mitigation: treat as secrets, avoid logging, ensure least-privilege DB access.
- **Performance Impact**: bcrypt is intentionally expensive. Mitigation: tune cost parameter; allow configuration; support caching when safe.
- **Migration Complexity**: Existing records need password metadata enrichment. Mitigation: prefer on-access enrichment; provide clear documentation.
- **API Misuse**: Bulk rekey operations can be dangerous. Mitigation: document as admin/maintenance paths; integrating application must enforce access control.

## Effort Estimation

- Development: 1–2 weeks (schema changes, bcrypt integration, helpers)
- Testing: 3–5 days (unit tests, property-based tests for bcrypt and rekey flows)
- Documentation: 1–2 days (user guide updates, migration guide)
- Migration Tools: 2–3 days (optional CLI or helper functions)

## Conclusion

Adding password metadata and bulk rekeying capabilities will significantly enhance VaultStore's security and manageability, providing practical paths for password-based grouping, key rotation, and incident response.
