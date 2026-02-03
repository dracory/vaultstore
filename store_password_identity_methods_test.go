package vaultstore

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestStoreForIdentity(t *testing.T) *storeImplementation {
	// Create an in-memory SQLite database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Initialize GORM
	gormDB, err := gorm.Open(&sqlite.Dialector{Conn: db}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to initialize GORM: %v", err)
	}

	store := &storeImplementation{
		vaultTableName:          "test_vault",
		vaultMetaTableName:      "test_vault_meta",
		db:                      db,
		gormDB:                  gormDB,
		dbDriverName:            "sqlite",
		passwordIdentityEnabled: true,
		cryptoConfig:            DefaultCryptoConfig(),
	}

	// Migrate tables
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return store
}

func TestFindIdentityID(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()
	password := "test-password-123"

	// Test finding non-existent identity
	_, err := store.findIdentityID(ctx, password)
	if !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("expected ErrIdentityNotFound, got: %v", err)
	}

	// Create identity
	passwordID, err := store.createIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	if passwordID == "" {
		t.Error("expected non-empty passwordID")
	}

	if passwordID[:2] != "p_" {
		t.Errorf("expected passwordID to start with 'p_', got: %s", passwordID)
	}

	// Now find the identity
	foundID, err := store.findIdentityID(ctx, password)
	if err != nil {
		t.Errorf("unexpected error finding identity: %v", err)
	}

	if foundID != passwordID {
		t.Errorf("expected %s, got %s", passwordID, foundID)
	}

	// Test with wrong password
	_, err = store.findIdentityID(ctx, "wrong-password")
	if !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("expected ErrIdentityNotFound for wrong password, got: %v", err)
	}
}

func TestFindOrCreateIdentity(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()
	password := "test-password-456"

	// First call should create
	id1, err := store.findOrCreateIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to find or create identity: %v", err)
	}

	// Second call should find existing
	id2, err := store.findOrCreateIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to find or create identity on second call: %v", err)
	}

	if id1 != id2 {
		t.Errorf("expected same ID on second call, got %s vs %s", id1, id2)
	}
}

func TestLinkRecordToIdentity(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "test-password-789"
	recordID := "test-record-123"

	// Create identity first
	passwordID, err := store.createIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	// Link record to identity
	err = store.linkRecordToIdentity(ctx, recordID, passwordID)
	if err != nil {
		t.Fatalf("failed to link record to identity: %v", err)
	}

	// Verify link exists
	foundPasswordID, err := store.getRecordPasswordID(ctx, recordID)
	if err != nil {
		t.Fatalf("failed to get record password ID: %v", err)
	}

	if foundPasswordID != passwordID {
		t.Errorf("expected %s, got %s", passwordID, foundPasswordID)
	}
}

func TestGetRecordPasswordID_NotFound(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	_, err := store.getRecordPasswordID(ctx, "non-existent-record")
	if !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("expected ErrIdentityNotFound, got: %v", err)
	}
}

func TestGetRecordsByPasswordID(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "shared-password"
	passwordID, err := store.createIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	// Create multiple records linked to same password
	recordIDs := []string{"record-1", "record-2", "record-3"}
	for _, recordID := range recordIDs {
		err := store.linkRecordToIdentity(ctx, recordID, passwordID)
		if err != nil {
			t.Fatalf("failed to link record %s: %v", recordID, err)
		}
	}

	// Get all records by password ID
	foundRecordIDs, err := store.getRecordsByPasswordID(ctx, passwordID)
	if err != nil {
		t.Fatalf("failed to get records by password ID: %v", err)
	}

	if len(foundRecordIDs) != len(recordIDs) {
		t.Errorf("expected %d records, got %d", len(recordIDs), len(foundRecordIDs))
	}

	// Verify all records found
	recordMap := make(map[string]bool)
	for _, id := range foundRecordIDs {
		recordMap[id] = true
	}

	for _, expectedID := range recordIDs {
		if !recordMap[expectedID] {
			t.Errorf("expected to find record %s", expectedID)
		}
	}
}

func TestCountRecordsByPasswordID(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "count-test-password"
	passwordID, err := store.createIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	// Initially should be 0
	count, err := store.countRecordsByPasswordID(ctx, passwordID)
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 records, got %d", count)
	}

	// Add some records
	for i := 0; i < 3; i++ {
		err := store.linkRecordToIdentity(ctx, "record-"+string(rune('a'+i)), passwordID)
		if err != nil {
			t.Fatalf("failed to link record: %v", err)
		}
	}

	// Should now be 3
	count, err = store.countRecordsByPasswordID(ctx, passwordID)
	if err != nil {
		t.Fatalf("failed to count records: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 records, got %d", count)
	}
}

func TestDeleteIdentityIfUnused(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "delete-test-password"
	passwordID, err := store.createIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	// Link a record
	recordID := "record-to-delete"
	err = store.linkRecordToIdentity(ctx, recordID, passwordID)
	if err != nil {
		t.Fatalf("failed to link record: %v", err)
	}

	// Should not delete because record is linked
	err = store.deleteIdentityIfUnused(ctx, passwordID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Identity should still exist
	_, err = store.findIdentityID(ctx, password)
	if err != nil {
		t.Errorf("identity should still exist: %v", err)
	}

	// Remove record link
	err = store.removeRecordLink(ctx, recordID)
	if err != nil {
		t.Fatalf("failed to remove record link: %v", err)
	}

	// Now should delete
	err = store.deleteIdentityIfUnused(ctx, passwordID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Identity should be gone
	_, err = store.findIdentityID(ctx, password)
	if !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("expected identity to be deleted: %v", err)
	}
}

func TestRemoveRecordLink(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "remove-test-password"
	passwordID, err := store.createIdentity(ctx, password)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	recordID := "record-to-unlink"
	err = store.linkRecordToIdentity(ctx, recordID, passwordID)
	if err != nil {
		t.Fatalf("failed to link record: %v", err)
	}

	// Verify link exists
	_, err = store.getRecordPasswordID(ctx, recordID)
	if err != nil {
		t.Fatalf("link should exist: %v", err)
	}

	// Remove link
	err = store.removeRecordLink(ctx, recordID)
	if err != nil {
		t.Fatalf("failed to remove link: %v", err)
	}

	// Verify link is gone
	_, err = store.getRecordPasswordID(ctx, recordID)
	if !errors.Is(err, ErrIdentityNotFound) {
		t.Errorf("expected link to be removed: %v", err)
	}
}
