package vaultstore

const COLUMN_CREATED_AT = "created_at"
const COLUMN_EXPIRES_AT = "expires_at"
const COLUMN_SOFT_DELETED_AT = "soft_deleted_at"
const COLUMN_ID = "id"
const COLUMN_UPDATED_AT = "updated_at"
const COLUMN_VAULT_TOKEN = "vault_token"
const COLUMN_VAULT_VALUE = "vault_value"

// Meta table column constants
const (
	COLUMN_OBJECT_TYPE = "object_type"
	COLUMN_OBJECT_ID   = "object_id"
	COLUMN_META_KEY    = "meta_key"
	COLUMN_META_VALUE  = "meta_value"
)

const TOKEN_PREFIX = "tk_"

// Token size constraints
const (
	TOKEN_MIN_PAYLOAD_LENGTH = 12
	TOKEN_MAX_PAYLOAD_LENGTH = 37                                           // 37 chars + tk_ prefix = 40 total
	TOKEN_MIN_TOTAL_LENGTH   = len(TOKEN_PREFIX) + TOKEN_MIN_PAYLOAD_LENGTH // 15
	TOKEN_MAX_TOTAL_LENGTH   = len(TOKEN_PREFIX) + TOKEN_MAX_PAYLOAD_LENGTH // 40
)

// Object type constants for vault_meta table
const (
	OBJECT_TYPE_PASSWORD_IDENTITY = "password_identity"
	OBJECT_TYPE_RECORD            = "record"
	OBJECT_TYPE_VAULT_SETTINGS    = "vault"
)

// Meta key constants
const (
	META_KEY_HASH        = "hash"
	META_KEY_PASSWORD_ID = "password_id"
	META_KEY_VERSION     = "version"
)

// Password identity ID prefix
const PASSWORD_ID_PREFIX = "p_"

// Record ID prefix (used in meta table)
const RECORD_META_ID_PREFIX = "r_"

// bcrypt cost for password hashing (legacy - used for backward compatibility)
const BCRYPT_COST = 12

// Argon2id password hashing parameters (lightweight defaults for broad compatibility)
// Users can increase these via CryptoConfig for higher security requirements
const (
	ARGON2ID_TIME     = 1         // Minimal passes for faster verification
	ARGON2ID_MEMORY   = 16 * 1024 // 16MB - lightweight for embedded/mobile
	ARGON2ID_THREADS  = 2         // Reduced parallelism
	ARGON2ID_KEY_LEN  = 32        // Hash output length
	ARGON2ID_SALT_LEN = 16        // Salt length
)

// Vault settings constants
const (
	VAULT_SETTINGS_ID = "settings"
)

// Encryption version constants for versioned encryption
const (
	ENCRYPTION_VERSION_V1 = "v1"
	ENCRYPTION_VERSION_V2 = "v2"
	ENCRYPTION_PREFIX_V1  = ENCRYPTION_VERSION_V1 + ":"
	ENCRYPTION_PREFIX_V2  = ENCRYPTION_VERSION_V2 + ":"
)

// v2 encryption parameters (AES-GCM + Argon2id)
const (
	V2_SALT_SIZE       = 16
	V2_NONCE_SIZE      = 12
	V2_TAG_SIZE        = 16
	ARGON2_ITERATIONS  = 3
	ARGON2_MEMORY      = 64 * 1024 // 64MB
	ARGON2_PARALLELISM = 4
	ARGON2_KEY_LENGTH  = 32
)

// CryptoConfig holds configurable cryptographic parameters
type CryptoConfig struct {
	// Argon2id parameters
	Iterations  int
	Memory      int // in bytes
	Parallelism int
	KeyLength   int // in bytes

	// AES-GCM parameters
	SaltSize  int // in bytes
	NonceSize int // in bytes
	TagSize   int // in bytes
}

// DefaultCryptoConfig returns secure default cryptographic parameters
func DefaultCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		Iterations:  ARGON2_ITERATIONS,
		Memory:      ARGON2_MEMORY,
		Parallelism: ARGON2_PARALLELISM,
		KeyLength:   ARGON2_KEY_LENGTH,
		SaltSize:    V2_SALT_SIZE,
		NonceSize:   V2_NONCE_SIZE,
		TagSize:     V2_TAG_SIZE,
	}
}

// HighSecurityCryptoConfig returns parameters for high-security scenarios
func HighSecurityCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		Iterations:  4,          // Increased iterations
		Memory:      128 * 1024, // 128MB memory
		Parallelism: 4,
		KeyLength:   32,
		SaltSize:    16,
		NonceSize:   12,
		TagSize:     16,
	}
}

// LightweightCryptoConfig returns parameters for resource-constrained environments
func LightweightCryptoConfig() *CryptoConfig {
	return &CryptoConfig{
		Iterations:  2,         // Reduced iterations
		Memory:      32 * 1024, // 32MB memory
		Parallelism: 2,
		KeyLength:   32,
		SaltSize:    16,
		NonceSize:   12,
		TagSize:     16,
	}
}
