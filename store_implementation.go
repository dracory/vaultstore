package vaultstore

import (
	"context"

	"database/sql"

	"github.com/dracory/database"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// Store defines a session store
type storeImplementation struct {
	vaultTableName           string
	vaultMetaTableName       string
	db                       *sql.DB
	gormDB                   *gorm.DB
	dbDriverName             string
	automigrateEnabled       bool
	debugEnabled             bool
	cryptoConfig             *CryptoConfig
	parallelThreshold        int  // Configurable threshold for parallel processing (0 = use default)
	passwordAllowEmpty       bool // Allow empty passwords (default: false)
	passwordMinLength        int  // Minimum password length (default: 16)
	passwordRequireLowercase bool // Require at least one lowercase letter (default: false)
	passwordRequireUppercase bool // Require at least one uppercase letter (default: false)
	passwordRequireNumbers   bool // Require at least one number (default: false)
	passwordRequireSymbols   bool // Require at least one symbol (default: false)
}

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

// AutoMigrate auto migrate
func (store *storeImplementation) AutoMigrate() error {
	// Clean up existing records with empty tokens before creating unique index
	err := store.cleanupEmptyTokenRecords()
	if err != nil {
		return err
	}

	// Use GORM's AutoMigrate with dynamic table name for vault records
	err = store.gormDB.Table(store.vaultTableName).AutoMigrate(&gormVaultRecord{})
	if err != nil {
		return err
	}

	// Always migrate the meta table
	return store.gormDB.Table(store.vaultMetaTableName).AutoMigrate(&gormVaultMeta{})
}

// cleanupEmptyTokenRecords removes or updates records with empty tokens to prevent unique index violations
func (store *storeImplementation) cleanupEmptyTokenRecords() error {
	// Check if the table exists first
	hasTable := store.gormDB.Migrator().HasTable(store.vaultTableName)
	if !hasTable {
		return nil
	}

	// Find all records with empty tokens
	var records []gormVaultRecord
	err := store.gormDB.Table(store.vaultTableName).
		Where(COLUMN_VAULT_TOKEN + " = ''").
		Find(&records).Error

	if err != nil {
		return err
	}

	// If no records with empty tokens, nothing to clean up
	if len(records) == 0 {
		return nil
	}

	// Delete records with empty tokens since they violate the unique constraint
	// and are likely test data or improperly created records
	return store.gormDB.Table(store.vaultTableName).
		Where(COLUMN_VAULT_TOKEN + " = ''").
		Delete(&gormVaultRecord{}).Error
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

// TokensReadToResolvedMap accepts a map of key token pairs and returns a map of key value pairs
//
// Example:
//
//	keyTokenMap := map[string]string{
//	  "key1": "token1",
//	  "key2": "token2",
//	}
//
//	resolvedMap, err := TokensReadToResolvedMap(keyTokenMap)
//	if err != nil {
//	  return
//	}
//
//	fmt.Println(resolvedMap)
//	// map[key1:value1 key2:value2]
//
// Parameters:
// - ctx (context.Context): The context to use for the operation
// - password (string): The vault key to use for decryption
// - keyTokenMap (map[string]string): A map of key token pairs
//
// Returns:
// - resolvedMap (map[string]string): A map of key value pairs
// - err (error): An error if one occurred
func (store *storeImplementation) TokensReadToResolvedMap(ctx context.Context, keyTokenMap map[string]string, password string) (map[string]string, error) {
	// Handle empty input map
	if len(keyTokenMap) == 0 {
		return map[string]string{}, nil
	}

	tokens := lo.Values(keyTokenMap)
	values, err := store.TokensRead(ctx, tokens, password)

	if err != nil {
		return map[string]string{}, err
	}

	resolved := lo.MapValues(keyTokenMap, func(token string, key string) string {
		return values[token]
	})

	// Filter out any keys where the token was not found (expired or missing)
	filtered := lo.PickBy(resolved, func(key string, value string) bool {
		return value != ""
	})

	return filtered, nil
}
