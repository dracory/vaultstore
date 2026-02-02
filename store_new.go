package vaultstore

import (
	"errors"

	"github.com/dracory/database"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// NewStore creates a new entity store
func NewStore(opts NewStoreOptions) (*storeImplementation, error) {
	if opts.VaultTableName == "" {
		return nil, errors.New("vault store: vaultTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("vault store: DB is required")
	}

	dbDriverName := opts.DbDriverName
	if dbDriverName == "" {
		dbDriverName = database.DatabaseType(opts.DB)
	}

	// Initialize GORM DB from existing *sql.DB using glebarez/sqlite (pure Go)
	gormDB, err := gorm.Open(&sqlite.Dialector{
		Conn: opts.DB,
	}, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	store := &storeImplementation{
		vaultTableName:     opts.VaultTableName,
		automigrateEnabled: opts.AutomigrateEnabled,
		db:                 opts.DB,
		gormDB:             gormDB,
		dbDriverName:       dbDriverName,
		debugEnabled:       opts.DebugEnabled,
	}

	if store.automigrateEnabled {
		err := store.AutoMigrate()
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
