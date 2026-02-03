package vaultstore

const COLUMN_CREATED_AT = "created_at"
const COLUMN_EXPIRES_AT = "expires_at"
const COLUMN_SOFT_DELETED_AT = "soft_deleted_at"
const COLUMN_ID = "id"
const COLUMN_UPDATED_AT = "updated_at"
const COLUMN_VAULT_TOKEN = "vault_token"
const COLUMN_VAULT_VALUE = "vault_value"

const TOKEN_PREFIX = "tk_"

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
