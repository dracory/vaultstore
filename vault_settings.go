package vaultstore

import (
	"context"
	"errors"
)

// GetVaultSetting retrieves a generic setting value from vault settings
func (store *storeImplementation) GetVaultSetting(ctx context.Context, key string) (string, error) {
	type metaRow struct {
		ID         string `db:"id"`
		ObjectType string `db:"object_type"`
		ObjectID   string `db:"object_id"`
		Key        string `db:"meta_key"`
		Value      string `db:"meta_value"`
	}

	var rows []metaRow
	err := store.db.Query().
		Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, key).
		Get(&rows)

	if err != nil {
		return "", err
	}

	if len(rows) == 0 {
		return "", errors.New("setting not found")
	}

	return rows[0].Value, nil
}

// SetVaultSetting sets a generic setting value in vault settings
func (store *storeImplementation) SetVaultSetting(ctx context.Context, key, value string) error {
	type metaRow struct {
		ID         string `db:"id"`
		ObjectType string `db:"object_type"`
		ObjectID   string `db:"object_id"`
		Key        string `db:"meta_key"`
		Value      string `db:"meta_value"`
	}

	var rows []metaRow
	err := store.db.Query().
		Table(store.vaultMetaTableName).
		Where("object_type = ? AND object_id = ? AND meta_key = ?", OBJECT_TYPE_VAULT_SETTINGS, VAULT_SETTINGS_ID, key).
		Get(&rows)

	if err != nil {
		return err
	}

	if len(rows) > 0 {
		// Update existing
		_, err := store.db.Query().
			Table(store.vaultMetaTableName).
			Where("id = ?", rows[0].ID).
			Update(map[string]any{
				"meta_value": value,
			})
		return err
	}

	// Create new
	meta := NewMeta().
		SetObjectType(OBJECT_TYPE_VAULT_SETTINGS).
		SetObjectID(VAULT_SETTINGS_ID).
		SetKey(key).
		SetValue(value)

	return store.db.Query().Table(store.vaultMetaTableName).Create(map[string]any{
		COLUMN_ID:          meta.GetID(),
		COLUMN_OBJECT_TYPE: meta.GetObjectType(),
		COLUMN_OBJECT_ID:   meta.GetObjectID(),
		COLUMN_META_KEY:    meta.GetKey(),
		COLUMN_META_VALUE:  meta.GetValue(),
	})
}
