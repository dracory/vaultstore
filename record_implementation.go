package vaultstore

import (
	"github.com/dracory/dataobject"
	"github.com/dracory/sb"
	"github.com/dracory/uid"
	"github.com/dromara/carbon/v2"
)

// == CLASS ==================================================================

type recordImplementation struct {
	dataobject.DataObject
}

// == CONSTRUCTORS ===========================================================

func NewRecord() RecordInterface {
	d := (&recordImplementation{}).
		SetID(uid.HumanUid()).
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetSoftDeletedAt(sb.MAX_DATETIME)

	return d
}

func NewRecordFromExistingData(data map[string]string) RecordInterface {
	o := &recordImplementation{}
	o.Hydrate(data)
	return o
}

// == METHODS ================================================================

// == SETTERS AND GETTERS ====================================================

func (v *recordImplementation) GetCreatedAt() string {
	return v.Get(COLUMN_CREATED_AT)
}

func (v *recordImplementation) SetCreatedAt(createdAt string) RecordInterface {
	v.Set(COLUMN_CREATED_AT, createdAt)
	return v
}

func (v *recordImplementation) GetSoftDeletedAt() string {
	return v.Get(COLUMN_SOFT_DELETED_AT)
}

func (v *recordImplementation) SetSoftDeletedAt(softDeletedAt string) RecordInterface {
	v.Set(COLUMN_SOFT_DELETED_AT, softDeletedAt)
	return v
}

func (v *recordImplementation) GetID() string {
	return v.Get(COLUMN_ID)
}

func (v *recordImplementation) SetID(id string) RecordInterface {
	v.Set(COLUMN_ID, id)
	return v
}

func (v *recordImplementation) GetToken() string {
	return v.Get(COLUMN_VAULT_TOKEN)
}

func (v *recordImplementation) SetToken(token string) RecordInterface {
	v.Set(COLUMN_VAULT_TOKEN, token)
	return v
}

func (v *recordImplementation) GetUpdatedAt() string {
	return v.Get(COLUMN_UPDATED_AT)
}

func (v *recordImplementation) SetUpdatedAt(updatedAt string) RecordInterface {
	v.Set(COLUMN_UPDATED_AT, updatedAt)
	return v
}

func (v *recordImplementation) GetValue() string {
	return v.Get(COLUMN_VAULT_VALUE)
}

func (v *recordImplementation) SetValue(value string) RecordInterface {
	v.Set(COLUMN_VAULT_VALUE, value)
	return v
}
