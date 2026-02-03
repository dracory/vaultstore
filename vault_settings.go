package vaultstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// GetVaultVersion retrieves the current vault version from settings
// Returns "1.0" if no version is set (default for legacy vaults)
func (store *storeImplementation) GetVaultVersion(ctx context.Context) (string, error) {
	var meta gormVaultMeta
	err := store.gormDB.
		WithContext(ctx).
		Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, META_KEY_VERSION).
		First(&meta).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No version set, return default
			return VAULT_VERSION_DEFAULT, nil
		}
		return "", err
	}

	return meta.Value, nil
}

// SetVaultVersion sets the vault version in settings
func (store *storeImplementation) SetVaultVersion(ctx context.Context, version string) error {
	// Check if setting already exists
	var existing gormVaultMeta
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, META_KEY_VERSION).
		First(&existing).Error

	if err == nil {
		// Update existing
		existing.Value = version
		return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Save(&existing).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create new
	meta := &gormVaultMeta{
		ObjectType: OBJECT_TYPE_VAULT_SETTINGS,
		ObjectID:   VAULT_SETTINGS_ID,
		Key:        META_KEY_VERSION,
		Value:      version,
	}

	return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Create(meta).Error
}

// IsVaultMigrated checks if the vault has been migrated to use password identities
// Returns true if version >= 1.1
func (store *storeImplementation) IsVaultMigrated(ctx context.Context) (bool, error) {
	version, err := store.GetVaultVersion(ctx)
	if err != nil {
		return false, err
	}

	// Parse versions
	currentVersion, err := parseVersion(version)
	if err != nil {
		return false, fmt.Errorf("failed to parse current version: %w", err)
	}

	targetVersion, err := parseVersion(VAULT_VERSION_WITH_IDENTITIES)
	if err != nil {
		return false, fmt.Errorf("failed to parse target version: %w", err)
	}

	return currentVersion >= targetVersion, nil
}

// MarkVaultMigrated marks the vault as fully migrated to version 1.1
func (store *storeImplementation) MarkVaultMigrated(ctx context.Context) error {
	return store.SetVaultVersion(ctx, VAULT_VERSION_WITH_IDENTITIES)
}

// parseVersion parses a version string (e.g., "1.1") into a float64 for comparison
func parseVersion(version string) (float64, error) {
	if version == "" {
		return 0, nil
	}
	v, err := strconv.ParseFloat(version, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid version format: %s", version)
	}
	return v, nil
}

// GetVaultSetting retrieves a generic setting value from vault settings
func (store *storeImplementation) GetVaultSetting(ctx context.Context, key string) (string, error) {
	var meta gormVaultMeta
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, key).
		First(&meta).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrIdentityNotFound
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
