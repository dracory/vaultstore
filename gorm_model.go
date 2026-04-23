package vaultstore

import "github.com/dromara/carbon/v2"

// gormVaultRecord is the internal GORM model for vault records
// This struct is used internally for database operations only
type gormVaultRecord struct {
	ID            string `gorm:"primaryKey;size:40;column:id;not null"`
	Token         string `gorm:"uniqueIndex;size:40;column:vault_token;not null"` // TOKEN_MAX_TOTAL_LENGTH
	Value         string `gorm:"type:longtext;column:vault_value;not null"`
	CreatedAt     string `gorm:"type:datetime;column:created_at;not null"`
	UpdatedAt     string `gorm:"type:datetime;column:updated_at;not null"`
	ExpiresAt     string `gorm:"type:datetime;column:expires_at;not null"`
	SoftDeletedAt string `gorm:"type:datetime;column:soft_deleted_at;not null"`
}

// TableName returns the table name for the GORM model
func (gormVaultRecord) TableName() string {
	return "" // Will be set dynamically via store.vaultTableName
}

// toRecordInterface converts a GORM record to a RecordInterface
func (g *gormVaultRecord) toRecordInterface() RecordInterface {
	// Set defaults for empty datetime fields to ensure NOT NULL constraint compliance
	createdAt := g.CreatedAt
	if createdAt == "" {
		createdAt = carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)
	}

	updatedAt := g.UpdatedAt
	if updatedAt == "" {
		updatedAt = carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)
	}

	expiresAt := g.ExpiresAt
	if expiresAt == "" {
		expiresAt = MAX_DATETIME
	}

	softDeletedAt := g.SoftDeletedAt
	if softDeletedAt == "" {
		softDeletedAt = MAX_DATETIME
	}

	data := map[string]string{
		COLUMN_ID:              g.ID,
		COLUMN_VAULT_TOKEN:     g.Token,
		COLUMN_VAULT_VALUE:     g.Value,
		COLUMN_CREATED_AT:      createdAt,
		COLUMN_UPDATED_AT:      updatedAt,
		COLUMN_EXPIRES_AT:      expiresAt,
		COLUMN_SOFT_DELETED_AT: softDeletedAt,
	}
	return NewRecordFromExistingData(data)
}

// fromRecordInterface creates a GORM record from a RecordInterface
func fromRecordInterface(r RecordInterface) *gormVaultRecord {
	return &gormVaultRecord{
		ID:            r.GetID(),
		Token:         r.GetToken(),
		Value:         r.GetValue(),
		CreatedAt:     r.GetCreatedAt(),
		UpdatedAt:     r.GetUpdatedAt(),
		ExpiresAt:     r.GetExpiresAt(),
		SoftDeletedAt: r.GetSoftDeletedAt(),
	}
}

// gormVaultMeta is the internal GORM model for vault metadata
// This struct is used internally for database operations only
type gormVaultMeta struct {
	ID         uint   `gorm:"primaryKey;column:id"`
	ObjectType string `gorm:"size:50;column:object_type"`
	ObjectID   string `gorm:"size:64;column:object_id"`
	Key        string `gorm:"size:50;column:meta_key"`
	Value      string `gorm:"type:text;column:meta_value"`
}

// TableName returns the table name for the GORM model
func (gormVaultMeta) TableName() string {
	return "" // Will be set dynamically via store.metaTableName
}
