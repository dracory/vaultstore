package vaultstore

import (
	"context"
	"fmt"
)

// BulkRekey re-encrypts all records using one password with a new password
// This implements the bulk rekeying workflow from the proposal
// Returns the number of records rekeyed
//
// If passwordIdentityEnabled is true, uses fast metadata-based lookup
// If passwordIdentityEnabled is false, uses scan-and-test approach (slower)
func (store *storeImplementation) BulkRekey(ctx context.Context, oldPassword, newPassword string) (int, error) {
	if oldPassword == "" || newPassword == "" {
		return 0, fmt.Errorf("passwords cannot be empty")
	}

	// Use fast path if identity feature is enabled
	if store.passwordIdentityEnabled {
		return store.bulkRekeyFast(ctx, oldPassword, newPassword)
	}

	// Fallback: Scan-and-Test (slower)
	// If metadata is disabled, we must try to decrypt every record
	return store.bulkRekeyScan(ctx, oldPassword, newPassword)
}

// bulkRekeyFast uses metadata lookup for efficient rekeying
func (store *storeImplementation) bulkRekeyFast(ctx context.Context, oldPassword, newPassword string) (int, error) {
	// 1. Find Identity for oldPass
	oldPassID, err := store.findIdentityID(ctx, oldPassword)
	if err != nil {
		return 0, fmt.Errorf("no records found with the old password: %w", err)
	}

	// 2. Find/Create Identity for newPass
	newPassID, err := store.findOrCreateIdentity(ctx, newPassword)
	if err != nil {
		return 0, fmt.Errorf("failed to create identity for new password: %w", err)
	}

	// 3. Find all records linked to oldPassID
	recordIDs, err := store.getRecordsByPasswordID(ctx, oldPassID)
	if err != nil {
		return 0, fmt.Errorf("failed to get records by password ID: %w", err)
	}

	if len(recordIDs) == 0 {
		return 0, nil
	}

	// 4. Re-encrypt loop
	count := 0
	for _, recordID := range recordIDs {
		// Get the record
		rec, err := store.RecordFindByID(ctx, recordID)
		if err != nil {
			return count, fmt.Errorf("failed to find record %s: %w", recordID, err)
		}
		if rec == nil {
			// Record might have been deleted, skip
			continue
		}

		// Decrypt with old password
		decryptedValue, err := decode(rec.GetValue(), oldPassword)
		if err != nil {
			// If decryption fails, skip this record
			continue
		}

		// Re-encrypt with new password
		encodedValue, err := encode(decryptedValue, newPassword)
		if err != nil {
			return count, fmt.Errorf("failed to encode value for record %s: %w", recordID, err)
		}

		// Update record value
		rec.SetValue(encodedValue)
		err = store.RecordUpdate(ctx, rec)
		if err != nil {
			return count, fmt.Errorf("failed to update record %s: %w", recordID, err)
		}

		// Update the link to use newPassID
		err = store.linkRecordToIdentity(ctx, recordID, newPassID)
		if err != nil {
			return count, fmt.Errorf("failed to link record %s to new identity: %w", recordID, err)
		}

		count++
	}

	// 5. Cleanup - delete old identity if no longer used
	err = store.deleteIdentityIfUnused(ctx, oldPassID)
	if err != nil {
		// Log but don't fail - the rekeying itself succeeded
		_ = err
	}

	return count, nil
}

// bulkRekeyScan scans all records and tries to decrypt with old password
// This is slower but works without identity metadata
func (store *storeImplementation) bulkRekeyScan(ctx context.Context, oldPassword, newPassword string) (int, error) {
	// Get all records
	records, err := store.RecordList(ctx, RecordQuery())
	if err != nil {
		return 0, fmt.Errorf("failed to list records: %w", err)
	}

	count := 0
	for _, rec := range records {
		// Try to decrypt with old password
		decryptedValue, err := decode(rec.GetValue(), oldPassword)
		if err != nil {
			// Decryption failed, this record doesn't use the old password
			continue
		}

		// Re-encrypt with new password
		encodedValue, err := encode(decryptedValue, newPassword)
		if err != nil {
			return count, fmt.Errorf("failed to encode value for record %s: %w", rec.GetID(), err)
		}

		// Update record value
		rec.SetValue(encodedValue)
		err = store.RecordUpdate(ctx, rec)
		if err != nil {
			return count, fmt.Errorf("failed to update record %s: %w", rec.GetID(), err)
		}

		count++
	}

	return count, nil
}

// MigrateRecordLinks migrates existing records to use password identities
// This should be called to link records created before identity-based management
// password is the password to try for decryption
// Returns the number of records migrated
// Returns error if password identity feature is not enabled
func (store *storeImplementation) MigrateRecordLinks(ctx context.Context, password string) (int, error) {
	if !store.passwordIdentityEnabled {
		return 0, fmt.Errorf("password identity feature is not enabled")
	}

	if password == "" {
		return 0, fmt.Errorf("password cannot be empty")
	}

	// Get or create identity for this password
	passwordID, err := store.findOrCreateIdentity(ctx, password)
	if err != nil {
		return 0, fmt.Errorf("failed to find or create identity: %w", err)
	}

	// Get all records
	records, err := store.RecordList(ctx, RecordQuery())
	if err != nil {
		return 0, fmt.Errorf("failed to list records: %w", err)
	}

	count := 0
	for _, rec := range records {
		// Check if already linked
		existingPassID, _ := store.getRecordPasswordID(ctx, rec.GetID())
		if existingPassID != "" {
			// Already linked, skip
			continue
		}

		// Try to decrypt with the password
		_, err := decode(rec.GetValue(), password)
		if err != nil {
			// Decryption failed, this record doesn't use this password
			continue
		}

		// Link the record
		err = store.linkRecordToIdentity(ctx, rec.GetID(), passwordID)
		if err != nil {
			return count, fmt.Errorf("failed to link record %s: %w", rec.GetID(), err)
		}

		count++
	}

	return count, nil
}
