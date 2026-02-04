package vaultstore

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// GetVaultSetting retrieves a generic setting value from vault settings
func (store *storeImplementation) GetVaultSetting(ctx context.Context, key string) (string, error) {
	var meta gormVaultMeta
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, key).
		First(&meta).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", gorm.ErrRecordNotFound
		}
		return "", err
	}

	return meta.Value, nil
}

// SetVaultSetting sets a generic setting value in vault settings
func (store *storeImplementation) SetVaultSetting(ctx context.Context, key, value string) error {
	// Check if setting already exists
	var existing gormVaultMeta
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, key).
		First(&existing).Error

	if err == nil {
		// Update existing
		existing.Value = value
		return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Save(&existing).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create new
	meta := &gormVaultMeta{
		ObjectType: OBJECT_TYPE_VAULT_SETTINGS,
		ObjectID:   VAULT_SETTINGS_ID,
		Key:        key,
		Value:      value,
	}

	return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Create(meta).Error
}
