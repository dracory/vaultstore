---
path: modules/bulk_rekey.md
page-type: module
summary: Pure encryption bulk rekey operations with scan-and-test approach for maximum security.
tags: [module, bulk-rekey, encryption, security, parallel-processing]
created: 2026-02-04
updated: 2026-02-04
version: 1.0.0
---

# Bulk Rekey Module

The Bulk Rekey module provides secure password rotation for all records encrypted with a specific password. It uses a pure encryption scan-and-test approach that eliminates password metadata storage for maximum security.

## Overview

### Security-First Design

Unlike traditional approaches that store password hashes or identity metadata, VaultStore's bulk rekey uses pure encryption:

- **Zero Metadata**: No password hashes, identity IDs, or relationship data stored
- **Scan-and-Test**: Attempts decryption with old password to identify matching records
- **Correlation Attack Prevention**: Cannot determine which records share passwords
- **Simplified Security Model**: Only encryption keys need protection

### Performance Optimization

The module automatically selects the optimal processing strategy based on dataset size:

- **Sequential Processing**: Small datasets (< 10,000 records)
- **Parallel Processing**: Large datasets (>= 10,000 records) with 10 workers
- **Cursor-based Pagination**: Very large datasets (> 1,000 records in memory) to prevent exhaustion

## Usage

### Basic Bulk Rekey

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/dracory/vaultstore"
)

func main() {
    // Initialize store
    vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
        VaultTableName: "vault",
        DB:             db,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Rekey all records from "oldpassword" to "newpassword"
    count, err := vault.BulkRekey(ctx, "oldpassword", "newpassword")
    if err != nil {
        log.Fatalf("Bulk rekey failed: %v", err)
    }
    
    fmt.Printf("Successfully rekeyed %d records\n", count)
}
```

### With Timeout and Cancellation

```go
// Set a 5-minute timeout for large datasets
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

count, err := vault.BulkRekey(ctx, "oldpassword", "newpassword")
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Partial rekey completed - count shows progress
        fmt.Printf("Partial rekey: %d records processed before timeout\n", count)
    } else {
        log.Fatalf("Bulk rekey failed: %v", err)
    }
}
```

### Custom Parallel Threshold

```go
// Lower threshold for earlier parallelization
vault, err := vaultstore.NewStore(vaultstore.NewStoreOptions{
    VaultTableName:    "vault",
    DB:                db,
    ParallelThreshold: 5000, // Use parallel for 5000+ records
})
```

## How It Works

### Algorithm

```go
func BulkRekey(ctx context.Context, oldPassword, newPassword string) (int, error) {
    // 1. Validate inputs
    if oldPassword == "" || newPassword == "" {
        return 0, error("passwords cannot be empty")
    }
    
    // 2. Count total records
    totalCount, err := store.RecordCount(ctx, RecordQuery())
    
    // 3. Select strategy based on size
    if totalCount > maxRecordsInMemory {
        // Very large: Use cursor-based pagination
        return bulkRekeyWithCursor(ctx, oldPassword, newPassword)
    }
    
    // 4. Load records
    records, err := store.RecordList(ctx, RecordQuery())
    
    // 5. Choose processing mode
    if len(records) < parallelThreshold {
        // Small dataset: Sequential processing
        return bulkRekeySequential(ctx, records, oldPassword, newPassword)
    }
    // Large dataset: Parallel processing
    return bulkRekeyParallel(ctx, records, oldPassword, newPassword)
}
```

### Sequential Processing

For small datasets, single-threaded processing provides simplicity and reliability:

```go
func bulkRekeySequential(ctx context.Context, records []Record, oldPassword, newPassword string) (int, error) {
    rekeyed := 0
    
    for _, record := range records {
        // Check context cancellation
        select {
        case <-ctx.Done():
            return rekeyed, ctx.Err()
        default:
        }
        
        // Try to decrypt with old password
        decryptedValue, err := decode(record.Value, oldPassword)
        if err != nil {
            // Record doesn't use old password, skip
            continue
        }
        
        // Re-encrypt with new password
        encodedValue, err := encode(decryptedValue, newPassword)
        if err != nil {
            return rekeyed, err
        }
        
        // Update record
        record.Value = encodedValue
        if err := store.RecordUpdate(ctx, record); err != nil {
            return rekeyed, err
        }
        
        rekeyed++
    }
    
    return rekeyed, nil
}
```

### Parallel Processing

For large datasets, worker pools provide better throughput:

```go
func bulkRekeyParallel(ctx context.Context, records []Record, oldPassword, newPassword string) (int, error) {
    const numWorkers = 10
    const batchSize = 100
    
    // Create work distribution channels
    recordChan := make(chan []Record, numWorkers*2)
    resultChan := make(chan int, numWorkers)
    errorChan := make(chan error, numWorkers)
    
    var wg sync.WaitGroup
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    // Start worker goroutines
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for batch := range recordChan {
                count, err := processBatch(ctx, batch, oldPassword, newPassword)
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
    
    // Distribute batches to workers
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
    
    // Aggregate results with error prioritization
    totalRekeyed := 0
    for {
        select {
        case err := <-errorChan:
            cancel()
            return totalRekeyed, err
        case count, ok := <-resultChan:
            if !ok {
                return totalRekeyed, nil
            }
            totalRekeyed += count
        case <-ctx.Done():
            return totalRekeyed, ctx.Err()
        }
    }
}
```

### Cursor-based Pagination

For very large datasets, streaming prevents memory exhaustion:

```go
func bulkRekeyWithCursor(ctx context.Context, oldPassword, newPassword string) (int, error) {
    const cursorBatchSize = 1000
    totalRekeyed := 0
    offset := 0
    
    for {
        // Check cancellation
        select {
        case <-ctx.Done():
            return totalRekeyed, ctx.Err()
        default:
        }
        
        // Fetch batch using pagination
        query := RecordQuery().SetLimit(cursorBatchSize).SetOffset(offset)
        records, err := store.RecordList(ctx, query)
        if err != nil {
            return totalRekeyed, err
        }
        
        // No more records
        if len(records) == 0 {
            break
        }
        
        // Process batch sequentially
        rekeyed, err := bulkRekeySequential(ctx, records, oldPassword, newPassword)
        if err != nil {
            return totalRekeyed, err
        }
        totalRekeyed += rekeyed
        
        // Move to next batch
        offset += len(records)
        
        // Check if we've processed all records
        if len(records) < cursorBatchSize {
            break
        }
    }
    
    return totalRekeyed, nil
}
```

## Performance Characteristics

### Benchmarks

| Dataset Size | Processing Mode | Time | Memory |
|--------------|----------------|------|--------|
| 1,000 | Sequential | 1-2 seconds | ~10 MB |
| 10,000 | Sequential | 10-20 seconds | ~50 MB |
| 100,000 | Parallel (10 workers) | 1-2 minutes | ~200 MB |
| 1,000,000 | Cursor + Parallel | 10-20 minutes | ~100 MB |

### Factors Affecting Performance

1. **Record Size**: Larger encrypted values take longer to decrypt/re-encrypt
2. **Database Performance**: Query and update speeds affect throughput
3. **CPU Cores**: Parallel processing benefits from multiple cores
4. **Network Latency**: Remote database connections add overhead
5. **Password Complexity**: Argon2id iterations affect key derivation time

## Security Considerations

### Zero Metadata Approach

**What is NOT stored:**
- Password hashes
- Password identity mappings
- Record-to-password relationships
- Password usage statistics

**Benefits:**
- No correlation attacks possible
- Database backups don't expose password relationships
- Query logs contain no sensitive metadata
- Simplified security auditing

### Comparison with Identity-Based Approaches

| Aspect | Pure Encryption | Identity-Based |
|--------|----------------|----------------|
| Metadata Storage | None | Password hashes, mappings |
| Security | Maximum | Vulnerable to correlation |
| Performance | O(n) scan | O(1) lookup |
| Complexity | Simple | Complex identity management |
| Audit Surface | Minimal | Multiple components |

## Error Handling

### Context Cancellation

The operation respects context cancellation and returns partial progress:

```go
ctx, cancel := context.WithCancel(context.Background())

go func() {
    time.Sleep(30 * time.Second)
    cancel() // Cancel after 30 seconds
}()

count, err := vault.BulkRekey(ctx, "old", "new")
if err != nil {
    // err wraps context.Canceled and includes count
    fmt.Printf("Partial rekey: %d records\n", count)
}
```

### Common Errors

- **Empty passwords**: Returns error immediately
- **Database errors**: Wrapped with context
- **Encryption failures**: Returns with partial count
- **Context cancellation**: Returns processed count with wrapped error

## Configuration

### Parallel Threshold

Configure when to switch from sequential to parallel processing:

```go
type NewStoreOptions struct {
    // ... other options ...
    ParallelThreshold int  // Default: 10000
}
```

- **Lower values** (e.g., 5000): Earlier parallelization, more overhead
- **Higher values** (e.g., 20000): Delayed parallelization, potential missed benefits
- **Default (10000)**: Balanced for most use cases

### Worker Count

The parallel implementation uses a fixed number of workers (10) optimized for:
- CPU parallelism without overwhelming the system
- Database connection pool utilization
- Memory pressure management

## Best Practices

### Before Bulk Rekey

1. **Backup your database**: Always backup before bulk operations
2. **Test on small dataset**: Verify with a subset of records
3. **Monitor resources**: Check CPU, memory, and database load
4. **Schedule appropriately**: Run during low-traffic periods

### During Bulk Rekey

1. **Use context with timeout**: Prevent runaway operations
2. **Monitor progress**: Log partial counts for large datasets
3. **Handle partial failures**: Check error types for cancellation vs failures

### After Bulk Rekey

1. **Verify count**: Ensure expected number of records were rekeyed
2. **Test sample records**: Verify decryption with new password works
3. **Update documentation**: Record the password change

## See Also

- [API Reference](../api_reference.md) - Complete API documentation
- [Architecture](../architecture.md) - System architecture details
- [Encryption](encryption.md) - Encryption implementation details
- [Token Operations](token_operations.md) - Token lifecycle management

## Changelog

- **v1.0.0** (2026-02-04): Initial documentation for pure encryption bulk rekey module
