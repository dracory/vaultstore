package vaultstore

import (
	"github.com/dracory/dataobject"
	"github.com/dracory/sb"
	"github.com/dracory/uid"
	"github.com/dromara/carbon/v2"
)

// == CLASS ==================================================================

type record struct {
	dataobject.DataObject
}

// == CONSTRUCTORS ===========================================================

func NewRecord() RecordInterface {
	d := (&record{}).
		SetID(uid.HumanUid()).
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetSoftDeletedAt(sb.MAX_DATETIME)

	return d
}

func NewRecordFromExistingData(data map[string]string) RecordInterface {
	o := &record{}
	o.Hydrate(data)
	return o
}

// == METHODS ================================================================

// == SETTERS AND GETTERS ====================================================

func (v *record) GetCreatedAt() string {
	return v.Get(COLUMN_CREATED_AT)
}

func (v *record) SetCreatedAt(createdAt string) RecordInterface {
	v.Set(COLUMN_CREATED_AT, createdAt)
	return v
}

func (v *record) GetSoftDeletedAt() string {
	return v.Get(COLUMN_SOFT_DELETED_AT)
}

func (v *record) SetSoftDeletedAt(softDeletedAt string) RecordInterface {
	v.Set(COLUMN_SOFT_DELETED_AT, softDeletedAt)
	return v
}

func (v *record) GetID() string {
	return v.Get(COLUMN_ID)
}

func (v *record) SetID(id string) RecordInterface {
	v.Set(COLUMN_ID, id)
	return v
}

func (v *record) GetToken() string {
	return v.Get(COLUMN_VAULT_TOKEN)
}

func (v *record) SetToken(token string) RecordInterface {
	v.Set(COLUMN_VAULT_TOKEN, token)
	return v
}

func (v *record) GetUpdatedAt() string {
	return v.Get(COLUMN_UPDATED_AT)
}

func (v *record) SetUpdatedAt(updatedAt string) RecordInterface {
	v.Set(COLUMN_UPDATED_AT, updatedAt)
	return v
}

func (v *record) GetValue() string {
	return v.Get(COLUMN_VAULT_VALUE)
}

func (v *record) SetValue(value string) RecordInterface {
	v.Set(COLUMN_VAULT_VALUE, value)
	return v
}
