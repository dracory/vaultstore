package vaultstore

// gormVaultRecord is the internal GORM model for vault records
// This struct is used internally for database operations only
type gormVaultRecord struct {
	ID            string `gorm:"primaryKey;size:40;column:id"`
	Token         string `gorm:"uniqueIndex;size:40;column:vault_token"`
	Value         string `gorm:"type:text;column:vault_value"`
	CreatedAt     string `gorm:"size:20;column:created_at"`
	UpdatedAt     string `gorm:"size:20;column:updated_at"`
	ExpiresAt     string `gorm:"size:20;column:expires_at"`
	SoftDeletedAt string `gorm:"size:20;column:soft_deleted_at"`
}

// TableName returns the table name for the GORM model
func (gormVaultRecord) TableName() string {
	return "" // Will be set dynamically via store.vaultTableName
}

// toRecordInterface converts a GORM record to a RecordInterface
func (g *gormVaultRecord) toRecordInterface() RecordInterface {
	data := map[string]string{
		COLUMN_ID:              g.ID,
		COLUMN_VAULT_TOKEN:     g.Token,
		COLUMN_VAULT_VALUE:     g.Value,
		COLUMN_CREATED_AT:      g.CreatedAt,
		COLUMN_UPDATED_AT:      g.UpdatedAt,
		COLUMN_EXPIRES_AT:      g.ExpiresAt,
		COLUMN_SOFT_DELETED_AT: g.SoftDeletedAt,
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
