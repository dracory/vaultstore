package vaultstore

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var ErrIdentityNotFound = errors.New("password identity not found")

// findIdentityID finds a password identity ID by scanning all stored hashes
// This implements the "Try-and-Verify" approach from the proposal
func (store *storeImplementation) findIdentityID(ctx context.Context, password string) (string, error) {
	type IdentityRow struct {
		ObjectID string `gorm:"column:object_id"`
		Value    string `gorm:"column:meta_value"`
	}

	var rows []IdentityRow
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND meta_key = ?", OBJECT_TYPE_PASSWORD_IDENTITY, META_KEY_HASH).
		Scan(&rows).Error

	if err != nil {
		return "", err
	}

	// Iterate and verify each hash
	for _, row := range rows {
		if verifyPassword(password, row.Value) {
			return row.ObjectID, nil
		}
	}

	return "", ErrIdentityNotFound
}

// createIdentity creates a new password identity with a bcrypt hash
func (store *storeImplementation) createIdentity(ctx context.Context, password string) (string, error) {
	passwordID := generatePasswordID()
	hash, err := hashPassword(password)
	if err != nil {
		return "", err
	}

	meta := &gormVaultMeta{
		ObjectType: OBJECT_TYPE_PASSWORD_IDENTITY,
		ObjectID:   passwordID,
		Key:        META_KEY_HASH,
		Value:      hash,
	}

	err = store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Create(meta).Error
	if err != nil {
		return "", err
	}

	return passwordID, nil
}

// findOrCreateIdentity finds an existing identity or creates a new one
func (store *storeImplementation) findOrCreateIdentity(ctx context.Context, password string) (string, error) {
	// Try to find existing identity
	passwordID, err := store.findIdentityID(ctx, password)
	if err == nil {
		return passwordID, nil
	}

	if !errors.Is(err, ErrIdentityNotFound) {
		return "", err
	}

	// Create new identity
	return store.createIdentity(ctx, password)
}

// linkRecordToIdentity links a record to a password identity
func (store *storeImplementation) linkRecordToIdentity(ctx context.Context, recordID string, passwordID string) error {
	metaID := generateRecordMetaID(recordID)

	// Check if a link already exists
	var existing gormVaultMeta
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_RECORD, metaID, META_KEY_PASSWORD_ID).
		First(&existing).Error

	if err == nil {
		// Update existing link
		existing.Value = passwordID
		return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Save(&existing).Error
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Create new link
	meta := &gormVaultMeta{
		ObjectType: OBJECT_TYPE_RECORD,
		ObjectID:   metaID,
		Key:        META_KEY_PASSWORD_ID,
		Value:      passwordID,
	}

	return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).Create(meta).Error
}

// getRecordPasswordID gets the password ID linked to a record
func (store *storeImplementation) getRecordPasswordID(ctx context.Context, recordID string) (string, error) {
	metaID := generateRecordMetaID(recordID)

	var meta gormVaultMeta
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_RECORD, metaID, META_KEY_PASSWORD_ID).
		First(&meta).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrIdentityNotFound
		}
		return "", err
	}

	return meta.Value, nil
}

// getRecordsByPasswordID finds all record IDs linked to a password identity
func (store *storeImplementation) getRecordsByPasswordID(ctx context.Context, passwordID string) ([]string, error) {
	var recordIDs []string

	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND meta_key = ? AND meta_value = ?", OBJECT_TYPE_RECORD, META_KEY_PASSWORD_ID, passwordID).
		Pluck("object_id", &recordIDs).Error

	if err != nil {
		return nil, err
	}

	// Strip the "r_" prefix from record IDs
	for i, id := range recordIDs {
		if len(id) > len(RECORD_META_ID_PREFIX) && id[:len(RECORD_META_ID_PREFIX)] == RECORD_META_ID_PREFIX {
			recordIDs[i] = id[len(RECORD_META_ID_PREFIX):]
		}
	}

	return recordIDs, nil
}

// countRecordsByPasswordID counts how many records are linked to a password identity
func (store *storeImplementation) countRecordsByPasswordID(ctx context.Context, passwordID string) (int64, error) {
	var count int64
	err := store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND meta_key = ? AND meta_value = ?", OBJECT_TYPE_RECORD, META_KEY_PASSWORD_ID, passwordID).
		Count(&count).Error

	return count, err
}

// deleteIdentityIfUnused deletes a password identity if no records reference it
func (store *storeImplementation) deleteIdentityIfUnused(ctx context.Context, passwordID string) error {
	count, err := store.countRecordsByPasswordID(ctx, passwordID)
	if err != nil {
		return err
	}

	if count > 0 {
		// Still in use, don't delete
		return nil
	}

	// Delete the identity rows
	return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ?", OBJECT_TYPE_PASSWORD_IDENTITY, passwordID).
		Delete(&gormVaultMeta{}).Error
}

// removeRecordLink removes the link between a record and its password identity
func (store *storeImplementation) removeRecordLink(ctx context.Context, recordID string) error {
	metaID := generateRecordMetaID(recordID)

	return store.gormDB.WithContext(ctx).Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_RECORD, metaID, META_KEY_PASSWORD_ID).
		Delete(&gormVaultMeta{}).Error
}
