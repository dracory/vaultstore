package vaultstore

import (
	"context"
	"testing"
)

func setupTestStoreForRekey(t *testing.T) *storeImplementation {
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_rekey_test",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: true,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nil received [%v]", err.Error())
	}

	return store
}

func TestBulkRekey(t *testing.T) {
	store := setupTestStoreForRekey(t)
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
	store := setupTestStoreForRekey(t)
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

func TestBulkRekey_NoMatchingRecords(t *testing.T) {
	store := setupTestStoreForRekey(t)
	ctx := context.Background()

	// Don't create any records, just try to rekey
	// With pure encryption, this returns 0, nil (not an error)
	count, err := store.BulkRekey(ctx, "non-existent-password", "new-password")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 records rekeyed, got %d", count)
	}
}

func TestBulkRekey_NoRecords(t *testing.T) {
	store := setupTestStoreForRekey(t)
	ctx := context.Background()

	// Empty store - should return 0, nil
	count, err := store.BulkRekey(ctx, "old-password", "new-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 records rekeyed, got %d", count)
	}
}

func TestBulkRekey_MixedPasswords(t *testing.T) {
	store := setupTestStoreForRekey(t)
	ctx := context.Background()

	oldPassword := "old-password-123"
	otherPassword := "other-password-789"
	newPassword := "new-password-456"

	// Create some records with the old password
	tokensOld := []string{}
	for i := 0; i < 3; i++ {
		token, err := store.TokenCreate(ctx, "old-value-"+string(rune('a'+i)), oldPassword, 32)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}
		tokensOld = append(tokensOld, token)
	}

	// Create some records with a different password
	tokensOther := []string{}
	for i := 0; i < 2; i++ {
		token, err := store.TokenCreate(ctx, "other-value-"+string(rune('a'+i)), otherPassword, 32)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}
		tokensOther = append(tokensOther, token)
	}

	// Perform bulk rekey - should only rekey records with old password
	count, err := store.BulkRekey(ctx, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("bulk rekey failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 records rekeyed (only old password records), got %d", count)
	}

	// Verify old password records can be read with new password
	for _, token := range tokensOld {
		_, err := store.TokenRead(ctx, token, newPassword)
		if err != nil {
			t.Errorf("failed to read rekeyed token with new password: %v", err)
		}
	}

	// Verify other password records still work with original password
	for _, token := range tokensOther {
		_, err := store.TokenRead(ctx, token, otherPassword)
		if err != nil {
			t.Errorf("failed to read non-rekeyed token with original password: %v", err)
		}
	}
}

func TestBulkRekey_SequentialVsParallel(t *testing.T) {
	store := setupTestStoreForRekey(t)
	ctx := context.Background()

	oldPassword := "old-password-123"
	newPassword := "new-password-456"

	// Create some records (less than 10000, so sequential processing)
	tokens := []string{}
	for i := 0; i < 10; i++ {
		token, err := store.TokenCreate(ctx, "value-"+string(rune('a'+i)), oldPassword, 32)
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}
		tokens = append(tokens, token)
	}

	// Perform bulk rekey (should use sequential processing)
	count, err := store.BulkRekey(ctx, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("bulk rekey failed: %v", err)
	}

	if count != 10 {
		t.Errorf("expected 10 records rekeyed, got %d", count)
	}

	// Verify all records work with new password
	for _, token := range tokens {
		_, err := store.TokenRead(ctx, token, newPassword)
		if err != nil {
			t.Errorf("failed to read token with new password: %v", err)
		}
	}
}

func TestBulkRekey_ParallelPath(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	// Create store with low parallel threshold to trigger parallel processing
	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_rekey_parallel_test",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: true,
		ParallelThreshold:  10, // Low threshold to trigger parallel with few records
	})
	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()
	oldPassword := "old-password-123"
	newPassword := "new-password-456"

	// Create 50 records - with threshold of 10, this triggers parallel processing
	tokens := []string{}
	for i := 0; i < 50; i++ {
		token, err := store.TokenCreate(ctx, "parallel-value-"+string(rune('a'+i%26)), oldPassword, 32)
		if err != nil {
			t.Fatalf("failed to create token %d: %v", i, err)
		}
		tokens = append(tokens, token)
	}

	// Perform bulk rekey (should use parallel processing due to low threshold)
	count, err := store.BulkRekey(ctx, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("bulk rekey failed: %v", err)
	}

	if count != 50 {
		t.Errorf("expected 50 records rekeyed, got %d", count)
	}

	// Verify a sample of records work with new password
	for i := 0; i < 20; i++ {
		_, err := store.TokenRead(ctx, tokens[i], newPassword)
		if err != nil {
			t.Errorf("failed to read token %d with new password: %v", i, err)
		}
	}

	// Verify old password no longer works for sample
	for i := 0; i < 20; i++ {
		_, err := store.TokenRead(ctx, tokens[i], oldPassword)
		if err == nil {
			t.Errorf("expected error when reading token %d with old password after rekey", i)
		}
	}
}

// TestBulkRekey_ContextCancellation tests context cancellation during processing
func TestBulkRekey_ContextCancellation(t *testing.T) {
	store := setupTestStoreForRekey(t)

	oldPassword := "old-password-123"
	newPassword := "new-password-456"

	// Create records to test context cancellation
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		_, err := store.TokenCreate(ctx, "cancel-value-"+string(rune('a'+i%26)), oldPassword, 32)
		if err != nil {
			t.Fatalf("failed to create token %d: %v", i, err)
		}
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Perform bulk rekey with cancelled context
	count, err := store.BulkRekey(ctx, oldPassword, newPassword)

	// Should return partial count and context error
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	// Count should be 0 or partial
	t.Logf("Context cancellation test: rekeyed %d records, error: %v", count, err)
}
