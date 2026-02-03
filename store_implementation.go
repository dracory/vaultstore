package vaultstore

import (
	"context"

	"database/sql"

	"github.com/dracory/database"
	"gorm.io/gorm"
)

// Store defines a session store
type storeImplementation struct {
	vaultTableName          string
	vaultMetaTableName      string
	db                      *sql.DB
	gormDB                  *gorm.DB
	dbDriverName            string
	automigrateEnabled      bool
	debugEnabled            bool
	cryptoConfig            *CryptoConfig
	passwordIdentityEnabled bool
}

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

// AutoMigrate auto migrate
func (store *storeImplementation) AutoMigrate() error {
	// Use GORM's AutoMigrate with dynamic table name for vault records
	err := store.gormDB.Table(store.vaultTableName).AutoMigrate(&gormVaultRecord{})
	if err != nil {
		return err
	}

	// Always migrate the meta table
	return store.gormDB.Table(store.vaultMetaTableName).AutoMigrate(&gormVaultMeta{})
}

// EnableDebug - enables the debug option
func (store *storeImplementation) EnableDebug(debug bool) {
	store.debugEnabled = debug
}

func (store *storeImplementation) GetDbDriverName() string {
	return store.dbDriverName
}

func (store *storeImplementation) GetVaultTableName() string {
	return store.vaultTableName
}

func (store *storeImplementation) GetMetaTableName() string {
	return store.vaultMetaTableName
}

func (store *storeImplementation) toQuerableContext(context context.Context) database.QueryableContext {
	if database.IsQueryableContext(context) {
		return context.(database.QueryableContext)
	}

	return database.Context(context, store.db)
}
