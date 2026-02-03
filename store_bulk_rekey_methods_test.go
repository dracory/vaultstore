package vaultstore

import (
	"context"
	"testing"
)

func TestBulkRekey(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	oldPassword := "old-password-123"
	newPassword := "new-password-456"

	// Create some records with the old password
	tokens := []string{}
	for i := 0; i < 3; i++ {
		token, err := store.TokenCreate(ctx, "value-"+string(rune('a'+i)), oldPassword, 32)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}
		tokens = append(tokens, token)
	}

	// Verify records are linked to old password identity
	for _, token := range tokens {
		rec, err := store.RecordFindByToken(ctx, token)
		if err != nil {
			t.Fatalf("failed to find record: %v", err)
		}
		if rec == nil {
			t.Fatal("expected non-nil record")
		}

		_, err = store.getRecordPasswordID(ctx, rec.GetID())
		if err != nil {
			t.Fatalf("record should be linked to password identity: %v", err)
		}
	}

	// Perform bulk rekey
	count, err := store.BulkRekey(ctx, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("bulk rekey failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 records rekeyed, got %d", count)
	}

	// Verify records can be read with new password
	for _, token := range tokens {
		_, err := store.TokenRead(ctx, token, newPassword)
		if err != nil {
			t.Errorf("failed to read token with new password: %v", err)
		}
	}

	// Verify old password no longer works
	for _, token := range tokens {
		_, err := store.TokenRead(ctx, token, oldPassword)
		if err == nil {
			t.Error("expected error when reading with old password after rekey")
		}
	}
}

func TestBulkRekey_EmptyPasswords(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	_, err := store.BulkRekey(ctx, "", "new-password")
	if err == nil {
		t.Error("expected error for empty old password")
	}

	_, err = store.BulkRekey(ctx, "old-password", "")
	if err == nil {
		t.Error("expected error for empty new password")
	}
}

func TestBulkRekey_NonExistentOldPassword(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	_, err := store.BulkRekey(ctx, "non-existent-password", "new-password")
	if err == nil {
		t.Error("expected error for non-existent old password")
	}
}

func TestBulkRekey_NoRecords(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	// Create an identity but no records
	oldPassword := "old-password"
	_, err := store.findOrCreateIdentity(ctx, oldPassword)
	if err != nil {
		t.Fatalf("failed to create identity: %v", err)
	}

	newPassword := "new-password"
	count, err := store.BulkRekey(ctx, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 records rekeyed, got %d", count)
	}
}

func TestMigrateRecordLinks(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "migration-password"

	// Create some records (they will be auto-linked on creation)
	// But we'll test with manually created records that aren't linked

	// First create a record directly without identity linking
	token, err := store.TokenCreate(ctx, "test-value", password, 32)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Get the record
	rec, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("failed to find record: %v", err)
	}

	// Verify it has a link (created during TokenCreate)
	_, err = store.getRecordPasswordID(ctx, rec.GetID())
	if err != nil {
		t.Fatalf("record should be linked: %v", err)
	}

	// Remove the link manually to simulate old record
	err = store.removeRecordLink(ctx, rec.GetID())
	if err != nil {
		t.Fatalf("failed to remove link: %v", err)
	}

	// Verify link is gone
	_, err = store.getRecordPasswordID(ctx, rec.GetID())
	if err == nil {
		t.Fatal("expected no link after removal")
	}

	// Now migrate
	count, err := store.MigrateRecordLinks(ctx, password)
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 record migrated, got %d", count)
	}

	// Verify link is restored
	_, err = store.getRecordPasswordID(ctx, rec.GetID())
	if err != nil {
		t.Errorf("record should be linked after migration: %v", err)
	}
}

func TestMigrateRecordLinks_EmptyPassword(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	_, err := store.MigrateRecordLinks(ctx, "")
	if err == nil {
		t.Error("expected error for empty password")
	}
}

func TestMigrateRecordLinks_AlreadyLinked(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	password := "test-password"

	// Create a record (will be auto-linked)
	_, err := store.TokenCreate(ctx, "test-value", password, 32)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Try to migrate - should skip already linked records
	count, err := store.MigrateRecordLinks(ctx, password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 records migrated (already linked), got %d", count)
	}
}

func TestMigrateRecordLinks_WrongPassword(t *testing.T) {
	store := setupTestStoreForIdentity(t)
	ctx := context.Background()

	correctPassword := "correct-password"
	wrongPassword := "wrong-password"

	// Create a record with correct password
	token, err := store.TokenCreate(ctx, "test-value", correctPassword, 32)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Get the record
	rec, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("failed to find record: %v", err)
	}

	// Remove link
	err = store.removeRecordLink(ctx, rec.GetID())
	if err != nil {
		t.Fatalf("failed to remove link: %v", err)
	}

	// Try to migrate with wrong password
	count, err := store.MigrateRecordLinks(ctx, wrongPassword)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 records migrated (wrong password), got %d", count)
	}
}
