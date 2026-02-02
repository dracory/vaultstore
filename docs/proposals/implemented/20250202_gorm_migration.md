# Migration to GORM

## Status: Implemented

## Overview

This proposal suggests migrating VaultStore's internal database storage layer to use GORM (Go Object Relational Mapper) as an implementation detail. The existing `recordInterface` and public APIs will remain unchanged - GORM will only replace the raw SQL implementation behind the scenes.

## Constraints

- **Interface Preservation**: The `recordInterface` (private) and all public methods remain unchanged
- **Implementation Only**: GORM is strictly an internal implementation detail for database access
- **API Compatibility**: No breaking changes to existing code using VaultStore

## Current Implementation

Currently, VaultStore uses a custom database abstraction layer:

- **Raw SQL**: Direct SQL queries constructed and executed manually
- **Manual Mapping**: Hand-written code for scanning rows into structs
- **Custom Migrations**: SQL-based migration scripts
- **Dialect Handling**: Manual abstraction for SQLite and PostgreSQL differences

## Proposed Changes

1. **Adopt GORM v2**: Migrate to GORM as the primary ORM layer
2. **Model Definitions**: Define structs with GORM tags for automatic table mapping
3. **Auto-Migrations**: Use GORM's AutoMigrate for schema management
4. **Query Builder**: Replace raw SQL with GORM's fluent query API
5. **Relationship Handling**: Leverage GORM's associations for related data

## Implementation Details

The migration will replace raw SQL execution with GORM operations **inside the existing implementation**, without changing any interfaces:

```go
// gormRecord is the GORM model (internal, with tags)
type gormRecord struct {
    ID        string `gorm:"primaryKey"`
    Token     string `gorm:"index"`
    Data      []byte
    CreatedAt string
    UpdatedAt string
    ExpiresAt *string
}

// recordImplementation remains unchanged (private interface implementation)
type recordImplementation struct {
    id        string
    token     string
    data      []byte
    createdAt string
    updatedAt string
    expiresAt *string
}

// NewRecordFromGorm constructor converts GORM model to recordImplementation
func NewRecordFromGorm(gr *gormRecord) *recordImplementation {
    return &recordImplementation{
        id:        gr.ID,
        token:     gr.Token,
        data:      gr.Data,
        createdAt: gr.CreatedAt,
        updatedAt: gr.UpdatedAt,
        expiresAt: gr.ExpiresAt,
    }
}

// Internal method implementation changes only
func (s *storeImplementation) recordFindByToken(ctx context.Context, token string) (recordInterface, error) {
    // BEFORE: raw SQL query
    // row := s.db.QueryRowContext(ctx, sqlFindByToken, token)
    
    // AFTER: GORM query (same return type, same interface)
    var gr gormRecord
    result := s.gormDB.WithContext(ctx).Where("token = ?", token).First(&gr)
    if result.Error != nil {
        return nil, result.Error
    }
    return NewRecordFromGorm(&gr), nil
}
```

Key points:
- `gormRecord` is the internal GORM model with struct tags
- `recordImplementation` remains unchanged as the private interface implementation
- `NewRecordFromGorm` constructor bridges the two types
- All public methods (`TokenCreate`, `RecordGet`, etc.) keep identical signatures
- `recordInterface` methods remain unchanged; only their internal implementation uses GORM
- Migration is transparent to library consumers

## Pros (Benefits)

| Benefit | Description |
|---------|-------------|
| **Reduced Boilerplate** | Eliminates manual SQL query construction and row scanning |
| **Type Safety** | Compile-time checks for database operations |
| **Cross-Database Support** | Built-in SQLite, PostgreSQL, MySQL support with minimal code changes |
| **Migration Management** | Automated schema migrations with AutoMigrate |
| **Query Building** | Fluent, chainable API for complex queries |
| **Relationships** | Easy handling of associations (has-one, has-many, belongs-to) |
| **Hooks** | Before/after create/update/delete callbacks |
| **Community** | Large ecosystem, extensive documentation, active maintenance |
| **Testing** | Easier to mock database layer for unit tests |
| **Connection Pooling** | Built-in connection management |

## Cons (Drawbacks)

| Drawback | Description |
|----------|-------------|
| **Performance Overhead** | ORM abstraction adds latency compared to raw SQL |
| **Learning Curve** | Team needs to learn GORM-specific patterns and conventions |
| **Migration Risk** | Existing data must be preserved during transition |
| **Complex Queries** | Raw SQL may still be needed for complex queries |
| **Dependency** | Adds external dependency (GORM + database drivers) |
| **Generated SQL** | Less control over exact SQL generated |
| **Memory Usage** | Reflection-heavy operations may use more memory |
| **Magic Behavior** | Implicit behaviors (callbacks, automatic timestamps) can be confusing |

## Risks and Mitigations

- **Data Loss Risk**: Migration could corrupt existing data. Mitigation: Full backup and dry-run migrations.
- **Performance Regression**: ORM overhead may slow operations. Mitigation: Benchmark before/after, optimize critical paths.
- **Breaking Changes**: API surface may change. Mitigation: Maintain backward compatibility layer.
- **Learning Curve**: Team unfamiliarity. Mitigation: Documentation and training sessions.

## Effort Estimation

- **Research & Planning**: 3-5 days
- **Core Migration**: 2-3 weeks
- **Testing & Validation**: 1-2 weeks
- **Documentation**: 3-5 days
- **Total**: ~4-6 weeks

## Conclusion

Migrating to GORM would modernize VaultStore's database layer, reducing maintenance burden and improving developer experience. However, the team should weigh the benefits against the performance overhead and migration effort, particularly for a security-focused storage component where performance and reliability are critical.
