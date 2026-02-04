package vaultstore

import (
	"context"
	"fmt"
	"sync"
)

// maxRecordsInMemory is the maximum number of records to load into memory at once
// for bulk rekey operations. This prevents memory exhaustion on very large datasets.
// Be conservative, some records can be large
const maxRecordsInMemory = 1000

// getParallelThreshold returns the configured threshold for parallel processing
// Returns 10000 if not configured (default)
func (store *storeImplementation) getParallelThreshold() int {
	if store.parallelThreshold > 0 {
		return store.parallelThreshold
	}
	return 10000
}

// BulkRekey performs pure encryption bulk rekey without password metadata
// It re-encrypts all records that can be decrypted with the old password using the new password
// Returns the number of records rekeyed
//
// This method uses a scan-and-test approach for maximum security:
//   - No password metadata is stored, preventing correlation attacks
//   - Each record is tested against the old password to determine if rekeying is needed
//   - Small datasets (< parallelThreshold records) use sequential processing
//   - Large datasets use parallel processing with 10 workers for better performance
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - oldPassword: The current password used to decrypt records
//   - newPassword: The new password to re-encrypt records with
//
// Returns:
//   - int: Number of records successfully rekeyed
//   - error: Error if the operation fails (nil on success)
//
// Example usage:
//
//	count, err := store.BulkRekey(ctx, "oldPassword123", "newSecurePassword456")
//	if err != nil {
//	    log.Fatalf("Bulk rekey failed: %v", err)
//	}
//	fmt.Printf("Successfully rekeyed %d records\n", count)
//
// Edge cases:
//   - Empty passwords: Returns error immediately
//   - No records in store: Returns 0, nil
//   - No records match old password: Returns 0, nil
//   - Context cancellation: Returns number processed so far, context error
//   - Mixed password records: Only rekeys records matching old password
func (store *storeImplementation) BulkRekey(ctx context.Context, oldPassword, newPassword string) (int, error) {
	if oldPassword == "" || newPassword == "" {
		return 0, fmt.Errorf("passwords cannot be empty")
	}

	// Get total count first to determine strategy
	totalCount, err := store.RecordCount(ctx, RecordQuery())
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	if totalCount == 0 {
		return 0, nil
	}

	// For large datasets, use cursor-based pagination to avoid memory exhaustion
	if totalCount > maxRecordsInMemory {
		return store.bulkRekeyWithCursor(ctx, oldPassword, newPassword)
	}

	// Get all records - safe for small datasets
	records, err := store.RecordList(ctx, RecordQuery())
	if err != nil {
		return 0, fmt.Errorf("failed to list records: %w", err)
	}

	// Choose processing strategy based on dataset size
	threshold := store.getParallelThreshold()
	if len(records) < threshold {
		return store.bulkRekeySequential(ctx, records, oldPassword, newPassword)
	}
	return store.bulkRekeyParallel(ctx, records, oldPassword, newPassword)
}

// bulkRekeySequential processes records sequentially for small datasets
// Returns partial count on context cancellation - caller must check error to determine if complete
func (store *storeImplementation) bulkRekeySequential(ctx context.Context, records []RecordInterface, oldPassword, newPassword string) (int, error) {
	rekeyed := 0

	for _, rec := range records {
		select {
		case <-ctx.Done():
			return rekeyed, fmt.Errorf("partial rekey completed %d records: %w", rekeyed, ctx.Err())
		default:
		}

		// Try to decrypt with old password
		decryptedValue, err := decode(rec.GetValue(), oldPassword)
		if err != nil {
			// Record doesn't use old password, skip it
			continue
		}

		// Re-encrypt with new password
		encodedValue, err := encode(decryptedValue, newPassword)
		if err != nil {
			return rekeyed, fmt.Errorf("failed to encode value for record %s: %w", rec.GetID(), err)
		}

		// Update record
		rec.SetValue(encodedValue)
		if err := store.RecordUpdate(ctx, rec); err != nil {
			return rekeyed, fmt.Errorf("failed to update record %s: %w", rec.GetID(), err)
		}

		rekeyed++
	}

	return rekeyed, nil
}

// bulkRekeyParallel processes records in parallel for large datasets
// Uses worker pool pattern with configurable number of workers and batch size
func (store *storeImplementation) bulkRekeyParallel(ctx context.Context, records []RecordInterface, oldPassword, newPassword string) (int, error) {
	// 10 workers chosen as balance between CPU parallelism and memory pressure
	// Each worker holds one batch (100 records) in memory
	// This provides good throughput without overwhelming system resources
	const numWorkers = 10
	const batchSize = 100

	// Create channels for work distribution
	recordChan := make(chan []RecordInterface, numWorkers*2)
	resultChan := make(chan int, numWorkers)
	errorChan := make(chan error, numWorkers)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range recordChan {
				count, err := store.processBatch(ctx, batch, oldPassword, newPassword)
				if err != nil {
					select {
					case errorChan <- err:
					case <-ctx.Done():
					}
					return
				}

				select {
				case resultChan <- count:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Send batches to workers
	go func() {
		defer close(recordChan)
		for i := 0; i < len(records); i += batchSize {
			end := i + batchSize
			if end > len(records) {
				end = len(records)
			}

			select {
			case recordChan <- records[i:end]:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Aggregate results with error priority to avoid race conditions
	totalRekeyed := 0
	for {
		// Check error channel first with non-blocking select to prioritize errors
		select {
		case err := <-errorChan:
			cancel()
			return totalRekeyed, err
		default:
		}

		select {
		case count, ok := <-resultChan:
			if !ok {
				return totalRekeyed, nil
			}
			totalRekeyed += count
		case err := <-errorChan:
			cancel()
			return totalRekeyed, err
		case <-ctx.Done():
			return totalRekeyed, fmt.Errorf("partial rekey completed %d records: %w", totalRekeyed, ctx.Err())
		}
	}
}

// processBatch processes a batch of records
// It tries to decrypt each record with the old password and re-encrypts with the new password
// Returns partial count on context cancellation - caller must check error to determine if complete
func (store *storeImplementation) processBatch(ctx context.Context, records []RecordInterface, oldPassword, newPassword string) (int, error) {
	rekeyed := 0

	for _, rec := range records {
		select {
		case <-ctx.Done():
			return rekeyed, fmt.Errorf("partial rekey completed %d records: %w", rekeyed, ctx.Err())
		default:
		}

		// Try to decrypt with old password
		decryptedValue, err := decode(rec.GetValue(), oldPassword)
		if err != nil {
			// Record doesn't use old password, skip it
			continue
		}

		// Re-encrypt with new password
		encodedValue, err := encode(decryptedValue, newPassword)
		if err != nil {
			return rekeyed, fmt.Errorf("failed to encode value for record %s: %w", rec.GetID(), err)
		}

		// Update record value
		rec.SetValue(encodedValue)
		if err := store.RecordUpdate(ctx, rec); err != nil {
			return rekeyed, fmt.Errorf("failed to update record %s: %w", rec.GetID(), err)
		}

		rekeyed++
	}

	return rekeyed, nil
}

// bulkRekeyWithCursor processes large datasets using cursor-based pagination
// to avoid loading all records into memory at once
// Returns partial count on context cancellation - caller must check error to determine if complete
func (store *storeImplementation) bulkRekeyWithCursor(ctx context.Context, oldPassword, newPassword string) (int, error) {
	const cursorBatchSize = 1000
	totalRekeyed := 0
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return totalRekeyed, fmt.Errorf("partial rekey completed %d records: %w", totalRekeyed, ctx.Err())
		default:
		}

		// Fetch batch of records using pagination
		query := RecordQuery().SetLimit(cursorBatchSize).SetOffset(offset)
		records, err := store.RecordList(ctx, query)
		if err != nil {
			return totalRekeyed, fmt.Errorf("failed to list records at offset %d: %w", offset, err)
		}

		// No more records to process
		if len(records) == 0 {
			break
		}

		// Process this batch
		rekeyed, err := store.bulkRekeySequential(ctx, records, oldPassword, newPassword)
		if err != nil {
			return totalRekeyed, err
		}
		totalRekeyed += rekeyed

		// Move to next batch
		offset += len(records)

		// If we got fewer records than batch size, we've processed all records
		if len(records) < cursorBatchSize {
			break
		}
	}

	return totalRekeyed, nil
}
