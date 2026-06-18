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

	if meta.GetID() == "" {
		t.Errorf("expected ID to be generated, got empty string")
	}
}

func TestNewMetaFromExistingData(t *testing.T) {
	data := map[string]string{
		"id":          "test-id-42",
		"object_type": "test_type",
		"object_id":   "test_object_123",
		"meta_key":    "test_key",
		"meta_value":  "test_value",
	}

	meta := NewMetaFromExistingData(data)

	if meta == nil {
		t.Fatal("expected non-nil meta")
	}

	if meta.GetID() != "test-id-42" {
		t.Errorf("expected ID to be 'test-id-42', got: %s", meta.GetID())
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
	meta.SetID("123")
	if meta.GetID() != "123" {
		t.Errorf("expected ID to be '123', got: %s", meta.GetID())
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

func TestMetaImplementation_SettersReturnInterface(t *testing.T) {
	meta := NewMeta()

	// All setters should return MetaInterface for chaining
	result := meta.
		SetID("test-id").
		SetObjectType("test").
		SetObjectID("test-id").
		SetKey("test-key").
		SetValue("test-value")

	if result == nil {
		t.Error("expected non-nil result from setter chain")
	}

	// Verify all values were set
	if meta.GetID() != "test-id" {
		t.Errorf("expected ID to be 'test-id', got: %s", meta.GetID())
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
	meta := NewMeta()
	meta.SetID("")
	if meta.GetID() != "" {
		t.Errorf("expected ID to be empty, got: %s", meta.GetID())
	}

	meta.SetID("my-meta-id")
	if meta.GetID() != "my-meta-id" {
		t.Errorf("expected ID to be 'my-meta-id', got: %s", meta.GetID())
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
	if meta2.GetID() != "12345" {
		t.Errorf("expected ID to be '12345', got: %s", meta2.GetID())
	}
}

func TestMetaImplementation_EmptyID(t *testing.T) {
	data := map[string]string{
		"id":          "",
		"object_type": "test",
		"object_id":   "test-id",
		"meta_key":    "test-key",
		"meta_value":  "test-value",
	}

	meta := NewMetaFromExistingData(data)
	if meta.GetID() != "" {
		t.Errorf("expected ID to be empty string, got: %s", meta.GetID())
	}
}
