package vaultstore

import (
	"time"

	"github.com/dracory/neat/database/orm"
	"github.com/dracory/neat/database/soft_delete"
	neatuid "github.com/dracory/neat/support/uid"
	"github.com/dromara/carbon/v2"
)

// == CLASS =====================================================================

type recordImplementation struct {
	orm.ShortID

	TokenField     string    `db:"vault_token"`
	ValueField     string    `db:"vault_value"`
	ExpiresAtField time.Time `db:"expires_at"`
	CreatedAtField orm.CreatedAt
	UpdatedAtField orm.UpdatedAt
	soft_delete.SoftDeletesMaxDate
}

// == CONSTRUCTORS ==============================================================

func NewRecord() RecordInterface {
	o := &recordImplementation{}
	o.SetID(neatuid.GenerateShortID())
	o.SetToken("")
	o.SetValue("")
	o.SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	o.SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))
	o.SetExpiresAt(MAX_DATETIME)
	o.SetSoftDeletedAt(MAX_DATETIME)
	return o
}

func NewRecordFromExistingData(data map[string]string) RecordInterface {
	o := &recordImplementation{}
	o.SetID(data[COLUMN_ID])
	o.SetToken(data[COLUMN_VAULT_TOKEN])
	o.SetValue(data[COLUMN_VAULT_VALUE])
	if v, ok := data[COLUMN_EXPIRES_AT]; ok {
		o.SetExpiresAt(v)
	}
	if v, ok := data[COLUMN_CREATED_AT]; ok {
		o.SetCreatedAt(v)
	}
	if v, ok := data[COLUMN_UPDATED_AT]; ok {
		o.SetUpdatedAt(v)
	}
	if v, ok := data[COLUMN_SOFT_DELETED_AT]; ok {
		o.SetSoftDeletedAt(v)
	}
	return o
}

// == SETTERS AND GETTERS =======================================================

func (o *recordImplementation) GetID() string {
	return o.ShortID.ID
}

func (o *recordImplementation) SetID(id string) RecordInterface {
	o.ShortID.ID = id
	return o
}

func (o *recordImplementation) GetToken() string {
	return o.TokenField
}

func (o *recordImplementation) SetToken(token string) RecordInterface {
	o.TokenField = token
	return o
}

func (o *recordImplementation) GetValue() string {
	return o.ValueField
}

func (o *recordImplementation) SetValue(value string) RecordInterface {
	o.ValueField = value
	return o
}

func (o *recordImplementation) GetExpiresAt() string {
	if o.ExpiresAtField.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.ExpiresAtField).ToDateTimeString()
}

func (o *recordImplementation) GetExpiresAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.ExpiresAtField)
}

func (o *recordImplementation) SetExpiresAt(expiresAt string) RecordInterface {
	if expiresAt == "" {
		return o
	}
	o.ExpiresAtField = carbon.Parse(expiresAt, carbon.UTC).StdTime()
	return o
}

func (o *recordImplementation) GetCreatedAt() string {
	if o.CreatedAtField.CreatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.CreatedAtField.CreatedAt).ToDateTimeString()
}

func (o *recordImplementation) GetCreatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.CreatedAtField.CreatedAt)
}

func (o *recordImplementation) SetCreatedAt(createdAt string) RecordInterface {
	if createdAt == "" {
		return o
	}
	o.CreatedAtField.CreatedAt = carbon.Parse(createdAt, carbon.UTC).StdTime()
	return o
}

func (o *recordImplementation) GetUpdatedAt() string {
	if o.UpdatedAtField.UpdatedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.UpdatedAtField.UpdatedAt).ToDateTimeString()
}

func (o *recordImplementation) GetUpdatedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.UpdatedAtField.UpdatedAt)
}

func (o *recordImplementation) SetUpdatedAt(updatedAt string) RecordInterface {
	if updatedAt == "" {
		return o
	}
	o.UpdatedAtField.UpdatedAt = carbon.Parse(updatedAt, carbon.UTC).StdTime()
	return o
}

func (o *recordImplementation) GetSoftDeletedAt() string {
	if o.SoftDeletesMaxDate.SoftDeletedAt.IsZero() {
		return ""
	}
	return carbon.CreateFromStdTime(o.SoftDeletesMaxDate.SoftDeletedAt).ToDateTimeString()
}

func (o *recordImplementation) GetSoftDeletedAtCarbon() *carbon.Carbon {
	return carbon.CreateFromStdTime(o.SoftDeletesMaxDate.SoftDeletedAt)
}

func (o *recordImplementation) SetSoftDeletedAt(deletedAt string) RecordInterface {
	if deletedAt == "" {
		return o
	}
	o.SoftDeletesMaxDate.SoftDeletedAt = carbon.Parse(deletedAt, carbon.UTC).StdTime()
	return o
}
