# Versioned Encryption with Backward Compatibility

*Proposal Date: February 2, 2026*

## Status: Implemented

## Overview

This proposal introduces **versioned encryption** to VaultStore, enabling secure migration from the current XOR-based scheme to modern AEAD encryption (AES-GCM/ChaCha20-Poly1305) **without breaking existing data**.

## Problem Statement

The current encryption has two critical issues:

1. **XOR encryption is cryptographically broken** - easily reversible, no authentication
2. **MD5/SHA1 password derivation is insecure** - fast, broken hashes, no salt

However, changing encryption breaks all existing stored secrets. This proposal solves the migration problem.

## Design Goals

1. **Zero breaking changes** - existing tokens remain readable
2. **Transparent migration** - new data uses secure encryption automatically
3. **No re-encryption required** - legacy data stays as-is until accessed
4. **Simple implementation** - minimal code changes, easy to audit

## Proposed Solution

### Version Header Pattern

Prefix encrypted values with version identifier:

```
v1:<xor_encrypted_data>     - Legacy (current)
v2:<aes_gcm_encrypted_data> - New secure encryption
```

**Constants:**
```go
const (
    ENCRYPTION_VERSION_V1 = "v1"
    ENCRYPTION_VERSION_V2 = "v2"
    ENCRYPTION_PREFIX_V1  = ENCRYPTION_VERSION_V1 + ":"
    ENCRYPTION_PREFIX_V2  = ENCRYPTION_VERSION_V2 + ":"
)
```

### Encryption Selection Logic

```go
func encode(value, password string) string {
    // Always use new encryption for new data
    key := deriveKeyArgon2id(password, generateSalt())
    ciphertext := aesGcmEncrypt(value, key)
    return ENCRYPTION_PREFIX_V2 + base64Encode(ciphertext)
}

func decode(value, password string) (string, error) {
    if strings.HasPrefix(value, ENCRYPTION_PREFIX_V2) {
        // New AES-GCM decryption
        return aesGcmDecrypt(strings.TrimPrefix(value, ENCRYPTION_PREFIX_V2), 
               deriveKeyArgon2id(password, extractSalt(value)))
    }
    // Legacy XOR decryption (backward compatible)
    return legacyXorDecode(value, password)
}
```

## Implementation Details

### 1. Schema Changes

**None required** - version prefix stored inline in existing `vault_value` column.

### 2. Key Derivation

**New (v2): Argon2id**
```go
func deriveKeyArgon2id(password string, salt []byte) []byte {
    return argon2.IDKey([]byte(password), salt, 
        3,          // iterations
        64*1024,    // memory (64MB)
        4,          // parallelism
        32)         // key length
}
```

**Legacy (v1): Keep existing `strongifyPassword`** for backward compatibility.

### 3. AEAD Encryption (v2)

```go
func aesGcmEncrypt(plaintext string, key []byte) string {
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64Encode(ciphertext)
}
```

### 4. Salt Storage

Embed salt in ciphertext for v2:
```
v2:<base64(salt + nonce + ciphertext)>
```

Structure: `[16 byte salt][12 byte nonce][ciphertext][16 byte tag]`

## Migration Strategy

### Phase 1: Deploy Versioned Encryption (Immediate)

- New records use v2 encryption automatically
- Existing records remain v1, readable via legacy path
- No bulk migration needed

### Phase 2: Opportunistic Re-encryption (Optional)

When a record is accessed:
```go
value, err := store.TokenRead(ctx, token, password)
if err == nil && isLegacyEncryption(record) {
    // Transparently upgrade to v2
    store.TokenUpdate(ctx, token, value, password)
}
```

### Phase 3: Legacy Cleanup (Future)

When all records are v2, remove v1 code path.

## API Changes

**No public API changes.** Internal `encode()`/`decode()` functions handle versioning transparently.

## Security Benefits

| Aspect | v1 (Current) | v2 (Proposed) |
|--------|--------------|---------------|
| Algorithm | XOR | AES-256-GCM |
| Key Derivation | MD5/SHA1 | Argon2id |
| Salt | None | Per-record random 16 bytes |
| Authentication | None | 128-bit GCM tag |
| Brute-force Cost | ~1B guesses/sec | ~100 guesses/sec |

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| v1 code path remains | Necessary for compatibility; clearly marked deprecated |
| Salt storage overhead | +16 bytes per record; negligible |
| Argon2id performance | Tunable parameters; cache derived keys at app layer if needed |
| Downgrade attacks | Version prefix prevents confusion between v1/v2 |

## Effort Estimation

- Implementation: 2-3 days
- Testing: 2 days (roundtrip, migration, edge cases)
- Documentation: 1 day

## Example Migration Path

```go
// Before: Uses v1 (XOR)
token, _ := store.TokenCreate(ctx, "secret", "password", 20)
// Stored: "xk9m2pQl..." (XOR encrypted)

// After upgrade: New records use v2 (AES-GCM)
token2, _ := store.TokenCreate(ctx, "secret", "password", 20)  
// Stored: "v2:base64(salt+nonce+ciphertext+tag)"

// Both readable:
val1, _ := store.TokenRead(ctx, token, "password")   // Decrypts v1
val2, _ := store.TokenRead(ctx, token2, "password")  // Decrypts v2
```

## Conclusion

This proposal provides a **practical, non-breaking path** to modern encryption security. Legacy data remains accessible while new data enjoys cryptographic best practices.
