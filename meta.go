package vaultstore

import (
	neatuid "github.com/dracory/neat/support/uid"
)

// == CLASS =====================================================================

type metaImplementation struct {
	IDField         string `db:"id"`
	ObjectTypeField string `db:"object_type"`
	ObjectIDField   string `db:"object_id"`
	KeyField        string `db:"meta_key"`
	ValueField      string `db:"meta_value"`
}

// == CONSTRUCTORS ==============================================================

func NewMeta() MetaInterface {
	o := &metaImplementation{}
	o.SetID(neatuid.GenerateShortID())
	o.SetObjectType("")
	o.SetObjectID("")
	o.SetKey("")
	o.SetValue("")
	return o
}

func NewMetaFromExistingData(data map[string]string) MetaInterface {
	o := &metaImplementation{}
	o.SetID(data[COLUMN_ID])
	o.SetObjectType(data[COLUMN_OBJECT_TYPE])
	o.SetObjectID(data[COLUMN_OBJECT_ID])
	o.SetKey(data[COLUMN_META_KEY])
	o.SetValue(data[COLUMN_META_VALUE])
	return o
}

// == SETTERS AND GETTERS =======================================================

func (o *metaImplementation) GetID() string {
	return o.IDField
}

func (o *metaImplementation) SetID(id string) MetaInterface {
	o.IDField = id
	return o
}

func (o *metaImplementation) GetObjectType() string {
	return o.ObjectTypeField
}

func (o *metaImplementation) SetObjectType(objectType string) MetaInterface {
	o.ObjectTypeField = objectType
	return o
}

func (o *metaImplementation) GetObjectID() string {
	return o.ObjectIDField
}

func (o *metaImplementation) SetObjectID(objectID string) MetaInterface {
	o.ObjectIDField = objectID
	return o
}

func (o *metaImplementation) GetKey() string {
	return o.KeyField
}

func (o *metaImplementation) SetKey(key string) MetaInterface {
	o.KeyField = key
	return o
}

func (o *metaImplementation) GetValue() string {
	return o.ValueField
}

func (o *metaImplementation) SetValue(value string) MetaInterface {
	o.ValueField = value
	return o
}
