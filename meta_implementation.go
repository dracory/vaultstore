package vaultstore

import (
	"strconv"

	"github.com/dracory/dataobject"
)

// metaImplementation is the internal struct for VaultMeta operations
type metaImplementation struct {
	dataobject.DataObject
}

// == CONSTRUCTORS ===========================================================

// NewMeta creates a new metadata entry
func NewMeta() MetaInterface {
	d := (&metaImplementation{}).
		SetObjectType("").
		SetObjectID("").
		SetKey("").
		SetValue("")

	return d
}

// NewMetaFromExistingData creates a metadata entry from existing data
func NewMetaFromExistingData(data map[string]string) MetaInterface {
	o := &metaImplementation{}
	o.Hydrate(data)
	return o
}

// == GETTERS ================================================================

func (m *metaImplementation) GetID() uint {
	idStr := m.Data()["id"]
	if idStr == "" {
		return 0
	}
	id, _ := strconv.ParseUint(idStr, 10, 64)
	return uint(id)
}

func (m *metaImplementation) GetObjectType() string {
	return m.Data()["object_type"]
}

func (m *metaImplementation) GetObjectID() string {
	return m.Data()["object_id"]
}

func (m *metaImplementation) GetKey() string {
	return m.Data()["meta_key"]
}

func (m *metaImplementation) GetValue() string {
	return m.Data()["meta_value"]
}

// == SETTERS ================================================================

func (m *metaImplementation) SetID(id uint) MetaInterface {
	m.Set("id", strconv.FormatUint(uint64(id), 10))
	return m
}

func (m *metaImplementation) SetObjectType(objectType string) MetaInterface {
	m.Set("object_type", objectType)
	return m
}

func (m *metaImplementation) SetObjectID(objectID string) MetaInterface {
	m.Set("object_id", objectID)
	return m
}

func (m *metaImplementation) SetKey(key string) MetaInterface {
	m.Set("meta_key", key)
	return m
}

func (m *metaImplementation) SetValue(value string) MetaInterface {
	m.Set("meta_value", value)
	return m
}
