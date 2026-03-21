package vaultstore

import (
	"context"
	"time"
)

// RecordInterface defines the methods that a Record must implement.
// It provides access to record data, timestamps, and metadata.
type RecordInterface interface {
	// Data returns the record data as a map
	Data() map[string]string
	// DataChanged returns the changed data as a map
	DataChanged() map[string]string

	// Getters
	// GetCreatedAt returns the created at timestamp
	GetCreatedAt() string
	// GetExpiresAt returns the expires at timestamp
	GetExpiresAt() string
	// GetSoftDeletedAt returns the soft deleted at timestamp
	GetSoftDeletedAt() string
	// GetID returns the record ID
	GetID() string
	// GetToken returns the record token
	GetToken() string
	// GetUpdatedAt returns the updated at timestamp
	GetUpdatedAt() string
	// GetValue returns the record value
	GetValue() string

	// Setters
	// SetCreatedAt sets the created at timestamp
	SetCreatedAt(createdAt string) RecordInterface
	// SetExpiresAt sets the expires at timestamp
	SetExpiresAt(expiresAt string) RecordInterface
	// SetSoftDeletedAt sets the soft deleted at timestamp
	SetSoftDeletedAt(softDeletedAt string) RecordInterface
	// SetID sets the record ID
	SetID(id string) RecordInterface
	// SetToken sets the record token
	SetToken(token string) RecordInterface
	// SetUpdatedAt sets the updated at timestamp
	SetUpdatedAt(updatedAt string) RecordInterface
	// SetValue sets the record value
	SetValue(value string) RecordInterface
}

// MetaInterface defines the methods that a VaultMeta must implement.
// It provides access to metadata for vault objects including keys and values.
type MetaInterface interface {
	// Data returns the meta data as a map
	Data() map[string]string
	// DataChanged returns the changed meta data as a map
	DataChanged() map[string]string

	// Getters
	// GetID returns the meta ID
	GetID() uint
	// GetObjectType returns the object type
	GetObjectType() string
	// GetObjectID returns the object ID
	GetObjectID() string
	// GetKey returns the meta key
	GetKey() string
	// GetValue returns the meta value
	GetValue() string

	// Setters
	// SetID sets the meta ID
	SetID(id uint) MetaInterface
	// SetObjectType sets the object type
	SetObjectType(objectType string) MetaInterface
	// SetObjectID sets the object ID
	SetObjectID(objectID string) MetaInterface
	// SetKey sets the meta key
	SetKey(key string) MetaInterface
	// SetValue sets the meta value
	SetValue(value string) MetaInterface
}

// RecordQueryInterface defines methods for building and executing record queries.
// It provides a fluent interface for setting query parameters and filters.
type RecordQueryInterface interface {
	// Validate validates the query parameters
	Validate() error

	// GetColumns returns the columns to select
	GetColumns() []string
	// SetColumns sets the columns to select
	SetColumns(columns []string) RecordQueryInterface
	// IsColumnsSet returns true if columns are set
	IsColumnsSet() bool

	// IsIDSet returns true if ID is set
	IsIDSet() bool
	// GetID returns the ID filter
	GetID() string
	// SetID sets the ID filter
	SetID(id string) RecordQueryInterface

	// IsIDInSet returns true if ID In filter is set
	IsIDInSet() bool
	// GetIDIn returns the ID In filter
	GetIDIn() []string
	// SetIDIn sets the ID In filter
	SetIDIn(idIn []string) RecordQueryInterface

	// IsTokenSet returns true if token is set
	IsTokenSet() bool
	// GetToken returns the token filter
	GetToken() string
	// SetToken sets the token filter
	SetToken(token string) RecordQueryInterface

	// IsTokenInSet returns true if token In filter is set
	IsTokenInSet() bool
	// GetTokenIn returns the token In filter
	GetTokenIn() []string
	// SetTokenIn sets the token In filter
	SetTokenIn(tokenIn []string) RecordQueryInterface

	// IsOffsetSet returns true if offset is set
	IsOffsetSet() bool
	// GetOffset returns the offset for pagination
	GetOffset() int
	// SetOffset sets the offset for pagination
	SetOffset(offset int) RecordQueryInterface

	// IsOrderBySet returns true if order by is set
	IsOrderBySet() bool
	// GetOrderBy returns the order by clause
	GetOrderBy() string
	// SetOrderBy sets the order by clause
	SetOrderBy(orderBy string) RecordQueryInterface

	// IsLimitSet returns true if limit is set
	IsLimitSet() bool
	// GetLimit returns the limit for pagination
	GetLimit() int
	// SetLimit sets the limit for pagination
	SetLimit(limit int) RecordQueryInterface

	// IsCountOnlySet returns true if count only is set
	IsCountOnlySet() bool
	// GetCountOnly returns the count only flag
	GetCountOnly() bool
	// SetCountOnly sets the count only flag
	SetCountOnly(countOnly bool) RecordQueryInterface

	// IsSortOrderSet returns true if sort order is set
	IsSortOrderSet() bool
	// GetSortOrder returns the sort order
	GetSortOrder() string
	// SetSortOrder sets the sort order
	SetSortOrder(sortOrder string) RecordQueryInterface

	// IsSoftDeletedIncludeSet returns true if soft deleted include is set
	IsSoftDeletedIncludeSet() bool
	// GetSoftDeletedInclude returns the soft deleted include flag
	GetSoftDeletedInclude() bool
	// SetSoftDeletedInclude sets the soft deleted include flag
	SetSoftDeletedInclude(softDeletedInclude bool) RecordQueryInterface
}

// StoreInterface defines the main interface for vault store operations.
// It provides methods for record management, token operations, and vault configuration.
//
// The store supports:
// - Record CRUD operations with soft delete support
// - Token-based encrypted storage with expiration
// - Bulk token operations for improved performance
// - Vault settings and metadata management
type StoreInterface interface {
	// AutoMigrate automatically migrates the database schema
	AutoMigrate() error
	// EnableDebug enables or disables debug mode
	EnableDebug(debug bool)

	// GetDbDriverName returns the database driver name
	GetDbDriverName() string
	// GetVaultTableName returns the vault table name
	GetVaultTableName() string
	// GetMetaTableName returns the meta table name
	GetMetaTableName() string

	// RecordCount returns the count of records matching the query
	RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error)
	// RecordCreate creates a new record
	RecordCreate(ctx context.Context, record RecordInterface) error
	// RecordDeleteByID deletes a record by its ID
	RecordDeleteByID(ctx context.Context, recordID string) error
	// RecordDeleteByToken deletes a record by its token
	RecordDeleteByToken(ctx context.Context, token string) error
	// RecordFindByID finds a record by its ID
	RecordFindByID(ctx context.Context, recordID string) (RecordInterface, error)
	// RecordFindByToken finds a record by its token
	RecordFindByToken(ctx context.Context, token string) (RecordInterface, error)
	// RecordList returns a list of records matching the query
	RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error)
	// RecordSoftDelete soft deletes a record
	RecordSoftDelete(ctx context.Context, record RecordInterface) error
	// RecordSoftDeleteByID soft deletes a record by its ID
	RecordSoftDeleteByID(ctx context.Context, recordID string) error
	// RecordSoftDeleteByToken soft deletes a record by its token
	RecordSoftDeleteByToken(ctx context.Context, token string) error
	// RecordUpdate updates an existing record
	RecordUpdate(ctx context.Context, record RecordInterface) error

	// TokenCreate creates a new token and returns the token string
	TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (token string, err error)
	// TokenCreateCustom creates a new token with a custom token string
	TokenCreateCustom(ctx context.Context, token string, value string, password string, options ...TokenCreateOptions) (err error)
	// TokenDelete deletes a token
	TokenDelete(ctx context.Context, token string) error
	// TokenExists checks if a token exists
	TokenExists(ctx context.Context, token string) (bool, error)
	// TokenRead reads the value of a token
	TokenRead(ctx context.Context, token string, password string) (string, error)
	// TokenRenew renews a token with a new expiration time
	TokenRenew(ctx context.Context, token string, expiresAt time.Time) error
	// TokensExpiredSoftDelete soft deletes all expired tokens
	TokensExpiredSoftDelete(ctx context.Context) (count int64, err error)
	// TokensExpiredDelete permanently deletes all expired tokens
	TokensExpiredDelete(ctx context.Context) (count int64, err error)
	// TokenSoftDelete soft deletes a token
	TokenSoftDelete(ctx context.Context, token string) error
	// TokenUpdate updates the value of a token
	TokenUpdate(ctx context.Context, token string, value string, password string) error

	// TokensRead reads multiple tokens at once with a single database query
	// This is more efficient than calling TokenRead multiple times
	TokensRead(ctx context.Context, tokens []string, password string) (map[string]string, error)

	// Token-based password management
	// TokensChangePassword changes the password for all tokens
	TokensChangePassword(ctx context.Context, oldPassword, newPassword string) (int, error)

	// TokensReadToResolvedMap accepts a map of key token pairs and returns a map of key value pairs
	// This is a convenience method that combines TokensRead and MapValues
	TokensReadToResolvedMap(ctx context.Context, keyTokenMap map[string]string, password string) (map[string]string, error)

	// Vault settings
	// GetVaultSetting gets a vault setting value
	GetVaultSetting(ctx context.Context, key string) (string, error)
	// SetVaultSetting sets a vault setting value
	SetVaultSetting(ctx context.Context, key, value string) error
}
