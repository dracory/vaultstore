package vaultstore

import (
	"testing"
)

func TestNewMeta(t *testing.T) {
	meta := NewMeta()

	if meta == nil {
		t.Fatal("expected non-nil meta")
	}

	// Check default values
	if meta.GetObjectType() != "" {
		t.Errorf("expected empty ObjectType, got: %s", meta.GetObjectType())
	}

	if meta.GetObjectID() != "" {
		t.Errorf("expected empty ObjectID, got: %s", meta.GetObjectID())
	}

	if meta.GetKey() != "" {
		t.Errorf("expected empty Key, got: %s", meta.GetKey())
	}

	if meta.GetValue() != "" {
		t.Errorf("expected empty Value, got: %s", meta.GetValue())
	}

	if meta.GetID() != 0 {
		t.Errorf("expected ID to be 0, got: %d", meta.GetID())
	}
}

func TestNewMetaFromExistingData(t *testing.T) {
	data := map[string]string{
		"id":          "42",
		"object_type": "test_type",
		"object_id":   "test_object_123",
		"meta_key":    "test_key",
		"meta_value":  "test_value",
	}

	meta := NewMetaFromExistingData(data)

	if meta == nil {
		t.Fatal("expected non-nil meta")
	}

	if meta.GetID() != 42 {
		t.Errorf("expected ID to be 42, got: %d", meta.GetID())
	}

	if meta.GetObjectType() != "test_type" {
		t.Errorf("expected ObjectType to be 'test_type', got: %s", meta.GetObjectType())
	}

	if meta.GetObjectID() != "test_object_123" {
		t.Errorf("expected ObjectID to be 'test_object_123', got: %s", meta.GetObjectID())
	}

	if meta.GetKey() != "test_key" {
		t.Errorf("expected Key to be 'test_key', got: %s", meta.GetKey())
	}

	if meta.GetValue() != "test_value" {
		t.Errorf("expected Value to be 'test_value', got: %s", meta.GetValue())
	}
}

func TestMetaImplementation_SettersAndGetters(t *testing.T) {
	meta := NewMeta()

	// Test SetID / GetID
	meta.SetID(123)
	if meta.GetID() != 123 {
		t.Errorf("expected ID to be 123, got: %d", meta.GetID())
	}

	// Test SetObjectType / GetObjectType
	meta.SetObjectType("password_identity")
	if meta.GetObjectType() != "password_identity" {
		t.Errorf("expected ObjectType to be 'password_identity', got: %s", meta.GetObjectType())
	}

	// Test SetObjectID / GetObjectID
	meta.SetObjectID("p_test-id-123")
	if meta.GetObjectID() != "p_test-id-123" {
		t.Errorf("expected ObjectID to be 'p_test-id-123', got: %s", meta.GetObjectID())
	}

	// Test SetKey / GetKey
	meta.SetKey("hash")
	if meta.GetKey() != "hash" {
		t.Errorf("expected Key to be 'hash', got: %s", meta.GetKey())
	}

	// Test SetValue / GetValue
	meta.SetValue("test-hash-value")
	if meta.GetValue() != "test-hash-value" {
		t.Errorf("expected Value to be 'test-hash-value', got: %s", meta.GetValue())
	}
}

func TestMetaImplementation_DataChanged(t *testing.T) {
	meta := NewMeta()

	// NewMeta sets empty strings via setters, which are tracked as changes
	changed := meta.DataChanged()
	// The dataobject tracks all Set calls, including empty string initializations
	if len(changed) == 0 {
		t.Logf("Note: dataobject tracks %d initial changes from NewMeta()", len(changed))
	}

	// Reset and test specific changes
	meta2 := &metaImplementation{}
	meta2.Hydrate(map[string]string{
		"id":          "0",
		"object_type": "",
		"object_id":   "",
		"meta_key":    "",
		"meta_value":  "",
	})

	// Initially after Hydrate, no changes
	changed2 := meta2.DataChanged()
	if len(changed2) != 0 {
		t.Errorf("expected no changes after Hydrate, got: %v", changed2)
	}

	// Make some changes
	meta2.SetObjectType("record")
	meta2.SetKey("password_id")
	meta2.SetValue("p_some-id")

	changed2 = meta2.DataChanged()
	if len(changed2) != 3 {
		t.Errorf("expected 3 changes, got: %d - %v", len(changed2), changed2)
	}

	if changed2["object_type"] != "record" {
		t.Errorf("expected object_type change to be 'record', got: %s", changed2["object_type"])
	}

	if changed2["meta_key"] != "password_id" {
		t.Errorf("expected meta_key change to be 'password_id', got: %s", changed2["meta_key"])
	}

	if changed2["meta_value"] != "p_some-id" {
		t.Errorf("expected meta_value change to be 'p_some-id', got: %s", changed2["meta_value"])
	}
}

func TestMetaImplementation_Data(t *testing.T) {
	meta := NewMeta()

	meta.SetID(999)
	meta.SetObjectType("vault")
	meta.SetObjectID("settings")
	meta.SetKey("version")
	meta.SetValue("1.1")

	data := meta.Data()

	if data["id"] != "999" {
		t.Errorf("expected id to be '999', got: %s", data["id"])
	}

	if data["object_type"] != "vault" {
		t.Errorf("expected object_type to be 'vault', got: %s", data["object_type"])
	}

	if data["object_id"] != "settings" {
		t.Errorf("expected object_id to be 'settings', got: %s", data["object_id"])
	}

	if data["meta_key"] != "version" {
		t.Errorf("expected meta_key to be 'version', got: %s", data["meta_key"])
	}

	if data["meta_value"] != "1.1" {
		t.Errorf("expected meta_value to be '1.1', got: %s", data["meta_value"])
	}
}

func TestMetaImplementation_SettersReturnInterface(t *testing.T) {
	meta := NewMeta()

	// All setters should return MetaInterface for chaining
	result := meta.
		SetID(1).
		SetObjectType("test").
		SetObjectID("test-id").
		SetKey("test-key").
		SetValue("test-value")

	if result == nil {
		t.Error("expected non-nil result from setter chain")
	}

	// Verify all values were set
	if meta.GetID() != 1 {
		t.Errorf("expected ID to be 1, got: %d", meta.GetID())
	}

	if meta.GetObjectType() != "test" {
		t.Errorf("expected ObjectType to be 'test', got: %s", meta.GetObjectType())
	}

	if meta.GetObjectID() != "test-id" {
		t.Errorf("expected ObjectID to be 'test-id', got: %s", meta.GetObjectID())
	}

	if meta.GetKey() != "test-key" {
		t.Errorf("expected Key to be 'test-key', got: %s", meta.GetKey())
	}

	if meta.GetValue() != "test-value" {
		t.Errorf("expected Value to be 'test-value', got: %s", meta.GetValue())
	}
}

func TestMetaImplementation_IDConversion(t *testing.T) {
	// Test uint to string conversion in SetID
	meta := NewMeta()
	meta.SetID(0)
	if meta.GetID() != 0 {
		t.Errorf("expected ID to be 0, got: %d", meta.GetID())
	}

	meta.SetID(4294967295) // max uint32
	if meta.GetID() != 4294967295 {
		t.Errorf("expected ID to be 4294967295, got: %d", meta.GetID())
	}

	// Test from existing data with string ID
	data := map[string]string{
		"id":          "12345",
		"object_type": "test",
		"object_id":   "test-id",
		"meta_key":    "test-key",
		"meta_value":  "test-value",
	}

	meta2 := NewMetaFromExistingData(data)
	if meta2.GetID() != 12345 {
		t.Errorf("expected ID to be 12345, got: %d", meta2.GetID())
	}
}

func TestMetaImplementation_EmptyID(t *testing.T) {
	// Test that empty string ID is handled as 0
	data := map[string]string{
		"id":          "",
		"object_type": "test",
		"object_id":   "test-id",
		"meta_key":    "test-key",
		"meta_value":  "test-value",
	}

	meta := NewMetaFromExistingData(data)
	if meta.GetID() != 0 {
		t.Errorf("expected ID to be 0 for empty string, got: %d", meta.GetID())
	}
}
