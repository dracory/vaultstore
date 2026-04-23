package vaultstore

import "testing"

func Test_NewRecordFromExistingData(t *testing.T) {
	data := map[string]string{
		COLUMN_ID:              "test-id",
		COLUMN_CREATED_AT:      "2024-01-01 00:00:00",
		COLUMN_UPDATED_AT:      "2024-01-02 00:00:00",
		COLUMN_SOFT_DELETED_AT: "9999-12-31 23:59:59",
		COLUMN_VAULT_TOKEN:     "test-token",
		COLUMN_VAULT_VALUE:     "test-value",
	}

	record := NewRecordFromExistingData(data)

	if record == nil {
		t.Fatal("Expected non-nil record")
	}

	if record.GetID() != "test-id" {
		t.Fatalf("Expected ID [test-id] received [%v]", record.GetID())
	}

	if record.GetCreatedAt() != "2024-01-01 00:00:00" {
		t.Fatalf("Expected CreatedAt [2024-01-01 00:00:00] received [%v]", record.GetCreatedAt())
	}

	if record.GetUpdatedAt() != "2024-01-02 00:00:00" {
		t.Fatalf("Expected UpdatedAt [2024-01-02 00:00:00] received [%v]", record.GetUpdatedAt())
	}

	if record.GetSoftDeletedAt() != "9999-12-31 23:59:59" {
		t.Fatalf("Expected SoftDeletedAt [9999-12-31 23:59:59] received [%v]", record.GetSoftDeletedAt())
	}

	if record.GetToken() != "test-token" {
		t.Fatalf("Expected Token [test-token] received [%v]", record.GetToken())
	}

	if record.GetValue() != "test-value" {
		t.Fatalf("Expected Value [test-value] received [%v]", record.GetValue())
	}

	// Ensure underlying Data map matches input
	recordData := record.Data()
	for k, v := range data {
		if recordData[k] != v {
			t.Fatalf("Expected Data[%s] [%s] received [%s]", k, v, recordData[k])
		}
	}
}

func Test_gormVaultRecord_toRecordInterface_EmptyDatetimes(t *testing.T) {
	// Test that toRecordInterface sets defaults for empty datetime fields
	gormRecord := &gormVaultRecord{
		ID:            "test-id",
		Token:         "test-token",
		Value:         "test-value",
		CreatedAt:     "", // Empty - should get default
		UpdatedAt:     "", // Empty - should get default
		ExpiresAt:     "", // Empty - should get MAX_DATETIME
		SoftDeletedAt: "", // Empty - should get MAX_DATETIME
	}

	record := gormRecord.toRecordInterface()

	if record == nil {
		t.Fatal("Expected non-nil record")
	}

	// Verify that empty datetime fields were set to defaults
	if record.GetCreatedAt() == "" {
		t.Fatal("Expected CreatedAt to be set to default, got empty string")
	}

	if record.GetUpdatedAt() == "" {
		t.Fatal("Expected UpdatedAt to be set to default, got empty string")
	}

	if record.GetExpiresAt() != MAX_DATETIME {
		t.Fatalf("Expected ExpiresAt [%s] received [%v]", MAX_DATETIME, record.GetExpiresAt())
	}

	if record.GetSoftDeletedAt() != MAX_DATETIME {
		t.Fatalf("Expected SoftDeletedAt [%s] received [%v]", MAX_DATETIME, record.GetSoftDeletedAt())
	}
}
