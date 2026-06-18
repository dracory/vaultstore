package vaultstore

import (
	"context"
	"errors"
	"time"

	"github.com/dromara/carbon/v2"
)

func (store *storeImplementation) RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error) {
	if err := ctx.Err(); err != nil {
		return -1, err
	}

	q := store.buildQuery(query)

	var count int64
	err := q.Table(store.vaultTableName).Count(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (store *storeImplementation) RecordCreate(ctx context.Context, record RecordInterface) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate that token is not empty to prevent unique index violations
	if record.GetToken() == "" {
		return errors.New("record token cannot be empty")
	}

	if record.GetCreatedAt() == "" {
		record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	}
	if record.GetUpdatedAt() == "" {
		record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	}
	if record.GetSoftDeletedAt() == "" {
		record.SetSoftDeletedAt(MAX_DATETIME)
	}

	row := map[string]any{
		COLUMN_ID:              record.GetID(),
		COLUMN_VAULT_TOKEN:     record.GetToken(),
		COLUMN_VAULT_VALUE:     record.GetValue(),
		COLUMN_CREATED_AT:      record.GetCreatedAtCarbon().StdTime(),
		COLUMN_UPDATED_AT:      record.GetUpdatedAtCarbon().StdTime(),
		COLUMN_EXPIRES_AT:      record.GetExpiresAtCarbon().StdTime(),
		COLUMN_SOFT_DELETED_AT: record.GetSoftDeletedAtCarbon().StdTime(),
	}

	return store.db.Query().Table(store.vaultTableName).Create(row)
}

func (store *storeImplementation) RecordDeleteByID(ctx context.Context, recordID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if recordID == "" {
		return errors.New("record id is empty")
	}

	_, err := store.db.Query().
		Table(store.vaultTableName).
		Where(COLUMN_ID+" = ?", recordID).
		Delete()

	return err
}

func (store *storeImplementation) RecordDeleteByToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if token == "" {
		return errors.New("token is empty")
	}

	_, err := store.db.Query().
		Table(store.vaultTableName).
		Where(COLUMN_VAULT_TOKEN+" = ?", token).
		Delete()

	return err
}

// RecordFindByID finds an entry by ID
func (store *storeImplementation) RecordFindByID(ctx context.Context, id string) (RecordInterface, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, errors.New("record id is empty")
	}

	query := RecordQuery().SetID(id).SetLimit(1)
	records, err := store.RecordList(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

// RecordFindByToken finds a record entity by token
func (store *storeImplementation) RecordFindByToken(ctx context.Context, token string) (RecordInterface, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if token == "" {
		return nil, errors.New("token is empty")
	}

	records, err := store.RecordList(ctx, RecordQuery().SetToken(token).SetLimit(1))
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

func (store *storeImplementation) RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error) {
	if err := ctx.Err(); err != nil {
		return []RecordInterface{}, err
	}

	err := query.Validate()
	if err != nil {
		return []RecordInterface{}, err
	}

	q := store.buildQuery(query)

	type recordRow struct {
		ID            string    `db:"id"`
		Token         string    `db:"vault_token"`
		Value         string    `db:"vault_value"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
		ExpiresAt     time.Time `db:"expires_at"`
		SoftDeletedAt time.Time `db:"soft_deleted_at"`
	}

	var rows []recordRow
	if err := q.Table(store.vaultTableName).Get(&rows); err != nil {
		return []RecordInterface{}, err
	}

	list := make([]RecordInterface, 0, len(rows))
	for _, r := range rows {
		o := &recordImplementation{}
		o.SetID(r.ID)
		o.SetToken(r.Token)
		o.SetValue(r.Value)
		o.CreatedAtField.CreatedAt = r.CreatedAt
		o.UpdatedAtField.UpdatedAt = r.UpdatedAt
		o.ExpiresAtField = r.ExpiresAt
		o.SoftDeletesMaxDate.SoftDeletedAt = r.SoftDeletedAt
		list = append(list, o)
	}

	return list, nil
}

// RecordSoftDelete soft deletes a record by setting the soft_deleted_at column to the current time
func (store *storeImplementation) RecordSoftDelete(ctx context.Context, record RecordInterface) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if record == nil {
		return errors.New("record is nil")
	}

	record.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RecordUpdate(ctx, record)
}

// RecordSoftDeleteByID soft deletes a record by ID by setting the soft_deleted_at column to the current time
func (store *storeImplementation) RecordSoftDeleteByID(ctx context.Context, recordID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if recordID == "" {
		return errors.New("record id is empty")
	}

	record, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		return err
	}

	if record == nil {
		return errors.New("record not found")
	}

	return store.RecordSoftDelete(ctx, record)
}

// RecordSoftDeleteByToken soft deletes a record by token by setting the soft_deleted_at column to the current time
func (store *storeImplementation) RecordSoftDeleteByToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if token == "" {
		return errors.New("token is empty")
	}

	record, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		return err
	}

	if record == nil {
		return errors.New("record not found")
	}

	return store.RecordSoftDelete(ctx, record)
}

func (store *storeImplementation) RecordUpdate(ctx context.Context, record RecordInterface) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if record == nil {
		return errors.New("record is nil")
	}

	if record.GetID() == "" {
		return errors.New("record id is empty")
	}

	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	row := map[string]any{
		COLUMN_VAULT_TOKEN:     record.GetToken(),
		COLUMN_VAULT_VALUE:     record.GetValue(),
		COLUMN_EXPIRES_AT:      record.GetExpiresAtCarbon().StdTime(),
		COLUMN_UPDATED_AT:      record.GetUpdatedAtCarbon().StdTime(),
		COLUMN_SOFT_DELETED_AT: record.GetSoftDeletedAtCarbon().StdTime(),
	}

	_, err := store.db.Query().
		Table(store.vaultTableName).
		Where(COLUMN_ID+" = ?", record.GetID()).
		Update(row)

	return err
}
