# Token Expiration Functionality

## Status: **Implemented**

- **Column vs Meta Table**: `expires_at` column added directly to vault table
- **Process Runner**: Client calls `CleanupExpiredTokens()` when needed

## Overview

This proposal suggests adding expiration functionality to tokens in VaultStore, allowing secrets to automatically expire after a specified time period.

## Current Implementation

Currently, VaultStore tokens do not have an expiration mechanism. Once created, tokens remain valid indefinitely unless explicitly deleted.

## Proposed Changes

1. **Add Expiration Field**: Add `COLUMN_EXPIRES_AT` constant and `expires_at` column to the database schema following the existing `COLUMN_SOFT_DELETED_AT` pattern.

2. **Extend Record Interface**: Add `GetExpiresAt()`/`SetExpiresAt()` methods to `RecordInterface` and `recordImplementation`.

3. **Expiration Parameter**: Extend the `TokenCreate` and `TokenCreateCustom` methods to accept an optional expiration duration parameter (zero duration = no expiration).

4. **Expiration Check on Read**: Modify `TokenRead` to check expiration and return an error if the token has expired.

5. **Query Support**: Add expiration filtering to `RecordQuery` interface for querying active/non-expired tokens.

6. **Cleanup Method**: Implement `CleanupExpiredTokens()` method that hard-deletes or soft-deletes expired tokens.

7. **Token Renewal**: Add `TokenRenew()` method to extend the expiration time of an existing token.

## Implementation Details

The implementation would require:

**1. Constants (`consts.go`)**
```go
const COLUMN_EXPIRES_AT = "expires_at"
```

**2. Schema Update (`sqls.go`)**
Add to `SqlCreateTable()`:
```go
Column(sb.Column{
    Name: COLUMN_EXPIRES_AT,
    Type: sb.COLUMN_TYPE_DATETIME,
})
```

**3. Record Interface Update (`interfaces.go`, `record_implementation.go`)**
```go
// In RecordInterface
GetExpiresAt() string
SetExpiresAt(expiresAt string) RecordInterface
```

**4. Token Method Signatures (`store_token_methods.go`)**
```go
// New method signatures
TokenCreate(ctx context.Context, data string, password string, tokenLength int, expiresIn time.Duration) (token string, err error)
TokenCreateCustom(ctx context.Context, token string, data string, password string, expiresIn time.Duration) (err error)
TokenRenew(ctx context.Context, token string, password string, expiresIn time.Duration) (err error)
CleanupExpiredTokens(ctx context.Context) (count int, err error)
```

**5. Expiration Check in `TokenRead`**
Compare `record.GetExpiresAt()` against current UTC time using carbon library.

**6. Cleanup Implementation**
Hard delete expired records or set `soft_deleted_at` based on configuration.

## Backward Compatibility

To maintain backward compatibility:

1. **Default Parameter**: `expiresIn` parameter with `time.Duration(0)` means no expiration
2. **New Methods**: Consider adding `TokenCreateWithExpiration` variants instead of modifying existing signatures
3. **Nullable Column**: `expires_at` should use `sb.MAX_DATETIME` (like `soft_deleted_at`) to indicate "no expiration"
4. **Grace Period**: Expired tokens return a specific error type so clients can distinguish between "not found" and "expired"

## Migration Strategy

For existing databases:

1. **Schema Migration**: Add `expires_at` column with default `"9999-12-31 23:59:59"` (same as `soft_deleted_at` pattern)
2. **Migration Method**: Provide `SqlAlterTable()` method similar to existing `SqlCreateTable()`
3. **Backward-Compatible Reads**: Tokens without expiration (null/`"9999-12-31 23:59:59"`) are treated as never expiring

## Benefits

- **Security Enhancement**: Automatically invalidates tokens after a certain period, reducing the risk of unauthorized access
- **Resource Management**: Helps keep the database clean by removing unused tokens
- **Compliance**: Supports compliance requirements that mandate credential rotation
- **Use Case Support**: Enables temporary access scenarios (e.g., one-time passwords, temporary API keys)

## Risks and Mitigations

- **Breaking Changes**: API changes might break existing code. Mitigation: Provide backward compatibility by making expiration optional.
- **Performance Impact**: Additional checks during token retrieval. Mitigation: Optimize database queries with proper indexing.
- **Clock Synchronization**: Reliance on system clock. Mitigation: Use server time consistently and document this dependency.

## Effort Estimation

- Development: 1-2 weeks
- Testing: 3-5 days
- Documentation: 1-2 days

## Conclusion

Adding token expiration functionality would significantly enhance the security and usability of VaultStore, making it more suitable for a wider range of use cases and security requirements.
