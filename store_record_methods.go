package vaultstore

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	"github.com/dracory/database"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

func (store *Store) RecordCount(ctx context.Context, query RecordQueryInterface) (int64, error) {
	query = query.SetCountOnly(true)
	dataset, _, err := query.toSelectDataset(store)

	if err != nil {
		return -1, err
	}

	sqlStr, sqlParams, errSql := dataset.Limit(1).
		Select(goqu.COUNT(goqu.Star()).As("count")).
		Prepared(true).
		ToSQL()

	if errSql != nil {
		return -1, nil
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	mapped, err := database.SelectToMapString(store.toQuerableContext(ctx), sqlStr, sqlParams...)

	if err != nil {
		return -1, err
	}

	if len(mapped) < 1 {
		return -1, nil
	}

	countStr := mapped[0]["count"]

	i, err := strconv.ParseInt(countStr, 10, 64)

	if err != nil {
		return -1, err

	}

	return i, nil
}

func (store *Store) RecordCreate(ctx context.Context, record RecordInterface) error {
	record.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	record.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	data := record.Data()

	sqlStr, sqlParams, errSql := goqu.Dialect(store.dbDriverName).
		Insert(store.vaultTableName).
		Prepared(true).
		Rows(data).
		ToSQL()

	if errSql != nil {
		if store.debugEnabled {
			log.Println(errSql.Error())
		}

		return errSql
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err := database.Execute(store.toQuerableContext(ctx), sqlStr, sqlParams...)

	if err != nil {
		if store.debugEnabled {
			log.Println(err)
		}

		return err
	}

	return nil
}

func (store *Store) RecordDeleteByID(ctx context.Context, recordID string) error {
	if recordID == "" {
		return errors.New("record id is empty")
	}

	q := goqu.Dialect(store.dbDriverName).
		Delete(store.vaultTableName).
		Prepared(true).
		Where(goqu.C(COLUMN_ID).Eq(recordID))

	sqlStr, sqlParams, err := q.ToSQL()

	if err != nil {
		return err
	}

	_, err = database.Execute(store.toQuerableContext(ctx), sqlStr, sqlParams...)

	if err != nil {
		return err
	}

	return nil
}

func (store *Store) RecordDeleteByToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	q := goqu.Dialect(store.dbDriverName).
		Delete(store.vaultTableName).
		Prepared(true).
		Where(goqu.C(COLUMN_VAULT_TOKEN).Eq(token))

	sqlStr, sqlParams, err := q.ToSQL()

	if err != nil {
		return err
	}

	_, err = database.Execute(store.toQuerableContext(ctx), sqlStr, sqlParams...)

	if err != nil {
		return err
	}

	return nil
}

// FindByID finds an entry by ID
func (st *Store) RecordFindByID(ctx context.Context, id string) (RecordInterface, error) {
	if id == "" {
		return nil, errors.New("record id is empty")
	}

	// Use RecordList with a query to ensure consistent soft delete handling
	query := RecordQuery().SetID(id).SetLimit(1)
	records, err := st.RecordList(ctx, query)
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
func (st *Store) RecordFindByToken(ctx context.Context, token string) (RecordInterface, error) {
	if token == "" {
		return nil, errors.New("token is empty")
	}

	// Use the query interface to properly handle soft deletion
	records, err := st.RecordList(ctx, RecordQuery().SetToken(token).SetLimit(1))
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

func (store *Store) RecordList(ctx context.Context, query RecordQueryInterface) ([]RecordInterface, error) {
	err := query.Validate()

	if err != nil {
		return []RecordInterface{}, err
	}

	dataset, columns, err := query.toSelectDataset(store)

	if err != nil {
		return []RecordInterface{}, err
	}

	sqlStr, sqlParams, errSql := dataset.Select(columns...).Prepared(true).ToSQL()

	if errSql != nil {
		return []RecordInterface{}, nil
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	modelMaps, err := database.SelectToMapString(store.toQuerableContext(ctx), sqlStr, sqlParams...)

	if err != nil {
		return []RecordInterface{}, err
	}

	list := []RecordInterface{}

	lo.ForEach(modelMaps, func(modelMap map[string]string, index int) {
		model := NewRecordFromExistingData(modelMap)
		list = append(list, model)
	})

	return list, nil
}

// RecordSoftDelete soft deletes a record by setting the soft_deleted_at column to the current time
func (store *Store) RecordSoftDelete(ctx context.Context, record RecordInterface) error {
	if record == nil {
		return errors.New("record is nil")
	}

	// Set the soft_deleted_at field to the current time
	record.SetSoftDeletedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	return store.RecordUpdate(ctx, record)
}

// RecordSoftDeleteByID soft deletes a record by ID by setting the soft_deleted_at column to the current time
func (store *Store) RecordSoftDeleteByID(ctx context.Context, recordID string) error {
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
func (store *Store) RecordSoftDeleteByToken(ctx context.Context, token string) error {
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

func (store *Store) RecordUpdate(ctx context.Context, record RecordInterface) error {
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

	sqlStr, sqlParams, err := goqu.Dialect(store.dbDriverName).
		Update(store.vaultTableName).
		Prepared(true).
		Set(dataChanged).
		Where(goqu.C(COLUMN_ID).Eq(record.GetID())).
		ToSQL()

	if err != nil {
		return err
	}

	if store.debugEnabled {
		log.Println(sqlStr)
	}

	_, err = database.Execute(store.toQuerableContext(ctx), sqlStr, sqlParams...)

	if err != nil {
		return err
	}

	return nil
}
