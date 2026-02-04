package vaultstore

import (
	"context"
	"errors"

	"github.com/dromara/carbon/v2"
	"gorm.io/gorm/clause"
)

func (store *storeImplementation) RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error) {
	if err := ctx.Err(); err != nil {
		return -1, err
	}

	var count int64

	db := store.gormDB.WithContext(ctx).Table(store.vaultTableName)

	// Apply filters from query
	if query.IsIDSet() && query.GetID() != "" {
		db = db.Where(COLUMN_ID+" = ?", query.GetID())
	}

	if query.IsTokenSet() && query.GetToken() != "" {
		db = db.Where(COLUMN_VAULT_TOKEN+" = ?", query.GetToken())
	}

	if query.IsIDInSet() && len(query.GetIDIn()) > 0 {
		db = db.Where(COLUMN_ID+" IN ?", query.GetIDIn())
	}

	if query.IsTokenInSet() && len(query.GetTokenIn()) > 0 {
		db = db.Where(COLUMN_VAULT_TOKEN+" IN ?", query.GetTokenIn())
	}

	// Handle soft delete filtering
	if !query.IsSoftDeletedIncludeSet() {
		db = db.Where(COLUMN_SOFT_DELETED_AT+" > ?", carbon.Now(carbon.UTC).ToDateTimeString())
	}

	err := db.Count(&count).Error
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (store *storeImplementation) RecordCreate(ctx context.Context, record RecordInterface) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	gormRecord := fromRecordInterface(record)

	err := store.gormDB.WithContext(ctx).Table(store.vaultTableName).Create(gormRecord).Error
	if err != nil {
		return err
	}

	return nil
}

func (store *storeImplementation) RecordDeleteByID(ctx context.Context, recordID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if recordID == "" {
		return errors.New("record id is empty")
	}

	err := store.gormDB.WithContext(ctx).Table(store.vaultTableName).
		Where(COLUMN_ID+" = ?", recordID).
		Delete(&gormVaultRecord{}).Error

	if err != nil {
		return err
	}

	return nil
}

func (store *storeImplementation) RecordDeleteByToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if token == "" {
		return errors.New("token is empty")
	}

	err := store.gormDB.WithContext(ctx).Table(store.vaultTableName).
		Where(COLUMN_VAULT_TOKEN+" = ?", token).
		Delete(&gormVaultRecord{}).Error

	if err != nil {
		return err
	}

	return nil
}

// RecordFindByID finds an entry by ID
func (store *storeImplementation) RecordFindByID(ctx context.Context, id string) (RecordInterface, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, errors.New("record id is empty")
	}

	// Use RecordList with a query to ensure consistent soft delete handling
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
//
// # If the supplied token is empty, an error is returned
//
// Parameters:
// - ctx: The context
// - token: The token to find
//
// Returns:
// - record: The record found
// - err: An error if something went wrong
func (store *storeImplementation) RecordFindByToken(ctx context.Context, token string) (RecordInterface, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if token == "" {
		return nil, errors.New("token is empty")
	}

	// Use the query interface to properly handle soft deletion
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

	var gormRecords []gormVaultRecord

	db := store.gormDB.WithContext(ctx).Table(store.vaultTableName)

	// Select specific columns if set
	if query.IsColumnsSet() && len(query.GetColumns()) > 0 {
		db = db.Select(query.GetColumns())
	}

	// Apply filters
	if query.IsIDSet() && query.GetID() != "" {
		db = db.Where(COLUMN_ID+" = ?", query.GetID())
	}

	if query.IsTokenSet() && query.GetToken() != "" {
		db = db.Where(COLUMN_VAULT_TOKEN+" = ?", query.GetToken())
	}

	if query.IsIDInSet() && len(query.GetIDIn()) > 0 {
		db = db.Where(COLUMN_ID+" IN ?", query.GetIDIn())
	}

	if query.IsTokenInSet() && len(query.GetTokenIn()) > 0 {
		db = db.Where(COLUMN_VAULT_TOKEN+" IN ?", query.GetTokenIn())
	}

	// Handle soft delete filtering
	if !query.IsSoftDeletedIncludeSet() {
		db = db.Where(COLUMN_SOFT_DELETED_AT+" > ?", carbon.Now(carbon.UTC).ToDateTimeString())
	}

	// Apply ordering
	if query.IsOrderBySet() && query.GetOrderBy() != "" {
		sortOrder := DESC
		if query.IsSortOrderSet() && query.GetSortOrder() != "" {
			sortOrder = query.GetSortOrder()
		}
		if sortOrder == ASC {
			db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: query.GetOrderBy()}, Desc: false})
		} else {
			db = db.Order(clause.OrderByColumn{Column: clause.Column{Name: query.GetOrderBy()}, Desc: true})
		}
	}

	// Apply limit and offset
	if query.IsLimitSet() && query.GetLimit() > 0 && !query.IsCountOnlySet() {
		db = db.Limit(query.GetLimit())
	}

	if query.IsOffsetSet() && query.GetOffset() > 0 && !query.IsCountOnlySet() {
		db = db.Offset(query.GetOffset())
	}

	err = db.Find(&gormRecords).Error
	if err != nil {
		return []RecordInterface{}, err
	}

	list := make([]RecordInterface, len(gormRecords))
	for i, gr := range gormRecords {
		list[i] = gr.toRecordInterface()
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

	// Set the soft_deleted_at field to the current time
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

	// Find the record first
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

	// Find the record first
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

	dataChanged := record.DataChanged()
	delete(dataChanged, COLUMN_ID) // ID is not updateable
	delete(dataChanged, "hash")    // Hash is not updateable

	if len(dataChanged) < 1 {
		return nil
	}

	// Convert dataChanged map to updates for GORM
	updates := make(map[string]interface{})
	for key, value := range dataChanged {
		updates[key] = value
	}

	err := store.gormDB.WithContext(ctx).Table(store.vaultTableName).
		Where(COLUMN_ID+" = ?", record.GetID()).
		Updates(updates).Error

	if err != nil {
		return err
	}

	return nil
}
