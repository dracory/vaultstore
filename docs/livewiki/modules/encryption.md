---
path: modules/encryption.md
page-type: module
summary: Encryption and decryption utilities for secure data storage.
tags: [module, encryption, security, cryptography]
created: 2026-02-03
updated: 2026-02-03
version: 1.0.0
---

# Encryption Module

The encryption module provides the cryptographic foundation for VaultStore, implementing secure encryption and decryption of stored values using industry-standard algorithms.

## Overview

The encryption module handles:
- Value encryption using AES-256-GCM
- Password-based key derivation with PBKDF2
- Secure random number generation
- Authentication tag validation
- Cryptographic error handling

## Core Algorithms

### AES-256-GCM Encryption

VaultStore uses AES-256-GCM (Galois/Counter Mode) for encryption:

```go
// AES-256-GCM provides:
// - Confidentiality (encryption)
// - Integrity (authentication)
// - Authenticated encryption
```

**Benefits:**
- **Strong Security**: 256-bit key length
- **Authentication**: Built-in integrity checking
- **Performance**: Efficient hardware acceleration
- **Standardized**: Widely supported and vetted

### PBKDF2 Key Derivation

Password-based key derivation using PBKDF2 with HMAC-SHA256:

```go
func deriveKey(password string, salt []byte) ([]byte, 32) {
    return pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
}
```

**Parameters:**
- **Iterations**: 100,000 (configurable for security/performance balance)
- **Key Length**: 32 bytes (256 bits)
- **Hash Function**: HMAC-SHA256
- **Salt**: Random salt per encryption

## Encryption Process

### Value Encryption

```go
func encrypt(value string, password string) (string, error) {
    // 1. Generate random salt
    salt := make([]byte, 16)
    _, err := rand.Read(salt)
    if err != nil {
        return "", err
    }
    
    // 2. Derive key from password and salt
    key := deriveKey(password, salt)
    
    // 3. Create AES-GCM cipher
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    // 4. Generate nonce
    nonce := make([]byte, aesGCM.NonceSize())
    _, err = rand.Read(nonce)
    if err != nil {
        return "", err
    }
    
    // 5. Encrypt and authenticate
    ciphertext := aesGCM.Seal(nil, nonce, []byte(value), nil)
    
    // 6. Combine salt + nonce + ciphertext
    result := append(salt, nonce...)
    result = append(result, ciphertext...)
    
    // 7. Encode as base64
    return base64.StdEncoding.EncodeToString(result), nil
}
```

### Data Format

Encrypted data format:

```
[16 bytes salt][12 bytes nonce][variable ciphertext]
```

- **Salt (16 bytes)**: Random salt for key derivation
- **Nonce (12 bytes)**: Random nonce for AES-GCM
- **Ciphertext**: Encrypted data with authentication tag

## Decryption Process

### Value Decryption

```go
func decrypt(encryptedValue string, password string) (string, error) {
    // 1. Decode base64
    data, err := base64.StdEncoding.DecodeString(encryptedValue)
    if err != nil {
        return "", err
    }
    
    // 2. Extract components
    if len(data) < 28 { // 16 + 12 minimum
        return "", errors.New("invalid encrypted data format")
    }
    
    salt := data[:16]
    nonce := data[16:28]
    ciphertext := data[28:]
    
    // 3. Derive key
    key := deriveKey(password, salt)
    
    // 4. Create AES-GCM cipher
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }
    
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    // 5. Decrypt and authenticate
    plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", ErrDecryptionFailed
    }
    
    return string(plaintext), nil
}
```

### Authentication Validation

The AES-GCM algorithm automatically validates the authentication tag during decryption. If the tag is invalid (indicating data tampering), the decryption fails with an error.

## Security Features

### Random Number Generation

Uses `crypto/rand` for cryptographically secure randomness:

```go
func generateRandomBytes(size int) ([]byte, error) {
    bytes := make([]byte, size)
    _, err := rand.Read(bytes)
    return bytes, err
}
```

**Applications:**
- Salt generation for key derivation
- Nonce generation for AES-GCM
- Token generation

### Key Security

#### Password Requirements

```go
func validatePassword(password string) error {
    if password == "" {
        return nil // No password is allowed (unencrypted storage)
    }
    
    if len(password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    // Additional validation can be added here
    return nil
}
```

#### Key Derivation Parameters

```go
const (
    // PBKDF2 iterations - balance between security and performance
    PBKDF2Iterations = 100000
    
    // Salt size for key derivation
    SaltSize = 16
    
    // Key size for AES-256
    KeySize = 32
    
    // Nonce size for AES-GCM
    NonceSize = 12
)
```

### Error Handling

#### Cryptographic Errors

```go
var (
    ErrDecryptionFailed = errors.New("decryption failed")
    ErrInvalidPassword  = errors.New("invalid password")
    ErrInvalidFormat    = errors.New("invalid encrypted data format")
)
```

#### Error Safety

- **No Information Leakage**: Errors don't reveal sensitive information
- **Constant Time Operations**: Prevent timing attacks where applicable
- **Secure Memory Handling**: Clear sensitive data when possible

## Usage Examples

### Basic Encryption/Decryption

```go
// Encrypt a value
value := "my secret data"
password := "my strong password"

encrypted, err := encrypt(value, password)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Encrypted: %s\n", encrypted)

// Decrypt the value
decrypted, err := decrypt(encrypted, password)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Decrypted: %s\n", decrypted)
```

### No Password Encryption

```go
// Encrypt without password (stored in plaintext)
value := "public data"
encrypted, err := encrypt(value, "")
if err != nil {
    log.Fatal(err)
}

// Decrypt without password
decrypted, err := decrypt(encrypted, "")
if err != nil {
    log.Fatal(err)
}
```

### Integration with Token Operations

```go
// Token creation with encryption
func (s *storeImplementation) TokenCreate(ctx context.Context, value string, password string, tokenLength int) (string, error) {
    // Encrypt the value
    encryptedValue, err := encrypt(value, password)
    if err != nil {
        return "", err
    }
    
    // Create record with encrypted value
    record := newRecord()
    record.SetToken(generateToken(tokenLength))
    record.SetValue(encryptedValue)
    record.SetCreatedAt(time.Now().Format(time.RFC3339))
    
    // Store in database
    err = s.RecordCreate(ctx, record)
    if err != nil {
        return "", err
    }
    
    return record.GetToken(), nil
}
```

## Performance Considerations

### PBKDF2 Iterations

The iteration count balances security and performance:

```go
// Higher iterations = more security but slower
// Lower iterations = faster but less secure

// Recommended values:
// - Development: 10,000 iterations
// - Production: 100,000+ iterations
// - High Security: 500,000+ iterations
```

### Hardware Acceleration

AES-NI hardware acceleration is automatically used when available:

```go
// Check if AES-NI is available (for debugging)
func hasAESNI() bool {
    return cpu.CPU.SupportsAES()
}
```

### Memory Usage

Encryption operations are memory efficient:

- **Stream Processing**: No need to load entire dataset
- **Fixed Overhead**: Constant memory usage regardless of data size
- **Garbage Collection**: Minimal allocation pressure

## Testing

### Unit Tests

```go
func TestEncryptionDecryption(t *testing.T) {
    value := "test value"
    password := "test password"
    
    // Encrypt
    encrypted, err := encrypt(value, password)
    require.NoError(t, err)
    assert.NotEmpty(t, encrypted)
    
    // Decrypt
    decrypted, err := decrypt(encrypted, password)
    require.NoError(t, err)
    assert.Equal(t, value, decrypted)
}

func TestEncryptionWithWrongPassword(t *testing.T) {
    value := "test value"
    password := "test password"
    wrongPassword := "wrong password"
    
    encrypted, err := encrypt(value, password)
    require.NoError(t, err)
    
    _, err = decrypt(encrypted, wrongPassword)
    assert.Error(t, err)
    assert.True(t, errors.Is(err, ErrDecryptionFailed))
}
```

### Security Tests

```go
func TestEncryptionTampering(t *testing.T) {
    value := "test value"
    password := "test password"
    
    encrypted, err := encrypt(value, password)
    require.NoError(t, err)
    
    // Tamper with encrypted data
    tampered := encrypted[:len(encrypted)-1] + "X"
    
    _, err = decrypt(tampered, password)
    assert.Error(t, err)
    assert.True(t, errors.Is(err, ErrDecryptionFailed))
}

func TestEncryptionUniqueness(t *testing.T) {
    value := "test value"
    password := "test password"
    
    // Encrypt same value multiple times
    encrypted1, err := encrypt(value, password)
    require.NoError(t, err)
    
    encrypted2, err := encrypt(value, password)
    require.NoError(t, err)
    
    // Should be different due to random salt/nonce
    assert.NotEqual(t, encrypted1, encrypted2)
}
```

### Performance Tests

```go
func BenchmarkEncryption(b *testing.B) {
    value := "test value for benchmarking"
    password := "test password"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := encrypt(value, password)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDecryption(b *testing.B) {
    value := "test value for benchmarking"
    password := "test password"
    encrypted, _ := encrypt(value, password)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := decrypt(encrypted, password)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Configuration

### Security Parameters

```go
type EncryptionConfig struct {
    PBKDF2Iterations int  // Key derivation iterations
    SaltSize        int  // Salt size in bytes
    KeySize         int  // Key size in bytes
    NonceSize       int  // Nonce size in bytes
}

func DefaultEncryptionConfig() EncryptionConfig {
    return EncryptionConfig{
        PBKDF2Iterations: 100000,
        SaltSize:        16,
        KeySize:         32,
        NonceSize:       12,
    }
}
```

### Performance Configuration

```go
// High performance configuration (less secure)
func HighPerformanceConfig() EncryptionConfig {
    return EncryptionConfig{
        PBKDF2Iterations: 10000,
        SaltSize:        16,
        KeySize:         32,
        NonceSize:       12,
    }
}

// High security configuration (slower)
func HighSecurityConfig() EncryptionConfig {
    return EncryptionConfig{
        PBKDF2Iterations: 500000,
        SaltSize:        32,
        KeySize:         32,
        NonceSize:       12,
    }
}
```

## Best Practices

### Security

1. **Use strong passwords** (8+ characters, mixed case, numbers, symbols)
2. **Never reuse passwords** across different tokens
3. **Protect encrypted data** at rest and in transit
4. **Regularly rotate passwords** for sensitive data

### Performance

1. **Choose appropriate PBKDF2 iterations** for your use case
2. **Batch operations** when processing multiple values
3. **Use connection pooling** for database operations
4. **Monitor encryption performance** in production

### Operations

1. **Validate inputs** before encryption
2. **Handle all errors** appropriately
3. **Log security events** without sensitive data
4. **Implement proper key management** practices

## Future Considerations

### Algorithm Updates

Potential future enhancements:
- **Argon2**: More modern key derivation function
- **ChaCha20-Poly1305**: Alternative to AES-GCM
- **Hardware Security Modules**: For key management
- **Quantum-Resistant Algorithms**: Future-proofing

### Key Management

Advanced key management features:
- **Key Rotation**: Automatic key updates
- **Key Escrow**: Recovery mechanisms
- **Multi-Factor Encryption**: Multiple passwords
- **Hardware-Based Keys**: TPM/HSM integration

## See Also

- [Token Operations](token_operations.md) - Token encryption usage
- [Core Store](core_store.md) - Store integration
- [Security Documentation](../security.md) - Comprehensive security guide
- [API Reference](../api_reference.md) - Complete API documentation
