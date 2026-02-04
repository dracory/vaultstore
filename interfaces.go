package vaultstore

import (
	"context"
	"time"
)

// RecordInterface defines the methods that a Record must implement
type RecordInterface interface {
	Data() map[string]string
	DataChanged() map[string]string

	// Getters
	GetCreatedAt() string
	GetExpiresAt() string
	GetSoftDeletedAt() string
	GetID() string
	GetToken() string
	GetUpdatedAt() string
	GetValue() string

	// Setters
	SetCreatedAt(createdAt string) RecordInterface
	SetExpiresAt(expiresAt string) RecordInterface
	SetSoftDeletedAt(softDeletedAt string) RecordInterface
	SetID(id string) RecordInterface
	SetToken(token string) RecordInterface
	SetUpdatedAt(updatedAt string) RecordInterface
	SetValue(value string) RecordInterface
}

// MetaInterface defines the methods that a VaultMeta must implement
type MetaInterface interface {
	Data() map[string]string
	DataChanged() map[string]string

	// Getters
	GetID() uint
	GetObjectType() string
	GetObjectID() string
	GetKey() string
	GetValue() string

	// Setters
	SetID(id uint) MetaInterface
	SetObjectType(objectType string) MetaInterface
	SetObjectID(objectID string) MetaInterface
	SetKey(key string) MetaInterface
	SetValue(value string) MetaInterface
}

type RecordQueryInterface interface {
	Validate() error

	GetColumns() []string
	SetColumns(columns []string) RecordQueryInterface
	IsColumnsSet() bool

	IsIDSet() bool
	GetID() string
	SetID(id string) RecordQueryInterface

	IsIDInSet() bool
	GetIDIn() []string
	SetIDIn(idIn []string) RecordQueryInterface

	IsTokenSet() bool
	GetToken() string
	SetToken(token string) RecordQueryInterface

	IsTokenInSet() bool
	GetTokenIn() []string
	SetTokenIn(tokenIn []string) RecordQueryInterface

	IsOffsetSet() bool
	GetOffset() int
	SetOffset(offset int) RecordQueryInterface

	IsOrderBySet() bool
	GetOrderBy() string
	SetOrderBy(orderBy string) RecordQueryInterface

	IsLimitSet() bool
	GetLimit() int
	SetLimit(limit int) RecordQueryInterface

	IsCountOnlySet() bool
	GetCountOnly() bool
	SetCountOnly(countOnly bool) RecordQueryInterface

	IsSortOrderSet() bool
	GetSortOrder() string
	SetSortOrder(sortOrder string) RecordQueryInterface

	IsSoftDeletedIncludeSet() bool
	GetSoftDeletedInclude() bool
	SetSoftDeletedInclude(softDeletedInclude bool) RecordQueryInterface
}

type StoreInterface interface {
	AutoMigrate() error
	EnableDebug(debug bool)

	GetDbDriverName() string
	GetVaultTableName() string
	GetMetaTableName() string

	RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error)
	RecordCreate(ctx context.Context, record RecordInterface) error
	RecordDeleteByID(ctx context.Context, recordID string) error
	RecordDeleteByToken(ctx context.Context, token string) error
	RecordFindByID(ctx context.Context, recordID string) (RecordInterface, error)
	RecordFindByToken(ctx context.Context, token string) (RecordInterface, error)
	RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error)
	RecordSoftDelete(ctx context.Context, record RecordInterface) error
	RecordSoftDeleteByID(ctx context.Context, recordID string) error
	RecordSoftDeleteByToken(ctx context.Context, token string) error
	RecordUpdate(ctx context.Context, record RecordInterface) error

	TokenCreate(ctx context.Context, value string, password string, tokenLength int, options ...TokenCreateOptions) (token string, err error)
	TokenCreateCustom(ctx context.Context, token string, value string, password string, options ...TokenCreateOptions) (err error)
	TokenDelete(ctx context.Context, token string) error
	TokenExists(ctx context.Context, token string) (bool, error)
	TokenRead(ctx context.Context, token string, password string) (string, error)
	TokenRenew(ctx context.Context, token string, expiresAt time.Time) error
	TokensExpiredSoftDelete(ctx context.Context) (count int64, err error)
	TokensExpiredDelete(ctx context.Context) (count int64, err error)
	TokenSoftDelete(ctx context.Context, token string) error
	TokenUpdate(ctx context.Context, token string, value string, password string) error
	TokensRead(ctx context.Context, tokens []string, password string) (map[string]string, error)

	// Token-based password management
	TokensChangePassword(ctx context.Context, oldPassword, newPassword string) (int, error)

	// Vault settings
	GetVaultSetting(ctx context.Context, key string) (string, error)
	SetVaultSetting(ctx context.Context, key, value string) error
}
