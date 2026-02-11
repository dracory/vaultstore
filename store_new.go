package vaultstore

import (
	"errors"
	"fmt"

	"github.com/dracory/database"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewStore creates a new entity store
func NewStore(opts NewStoreOptions) (*storeImplementation, error) {
	if opts.VaultTableName == "" {
		return nil, errors.New("vault store: vaultTableName is required")
	}

	if opts.VaultMetaTableName == "" {
		return nil, errors.New("vault store: vaultMetaTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("vault store: DB is required")
	}

	dbDriverName := opts.DbDriverName
	if dbDriverName == "" {
		dbDriverName = database.DatabaseType(opts.DB)
	}

	// Set crypto config with secure defaults
	cryptoConfig := opts.CryptoConfig
	if cryptoConfig == nil {
		cryptoConfig = DefaultCryptoConfig()
	}

	var dialector gorm.Dialector

	dbType := database.DatabaseType(opts.DB)
	switch dbType {
	case "sqlite":
		dialector = sqlite.New(sqlite.Config{Conn: opts.DB})
	case "mysql":
		dialector = mysql.New(mysql.Config{Conn: opts.DB})
	case "postgres", "postgresql":
		dialector = postgres.New(postgres.Config{Conn: opts.DB})
	default:
		return nil, fmt.Errorf("unsupported database connection: %s", dbType)
	}

	// Initialize GORM DB from existing *sql.DB using glebarez/sqlite (pure Go)
	// gormDB, err := gorm.Open(&sqlite.Dialector{
	// 	Conn: opts.DB,
	// }, &gorm.Config{})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	store := &storeImplementation{
		vaultTableName:           opts.VaultTableName,
		vaultMetaTableName:       opts.VaultMetaTableName,
		automigrateEnabled:       opts.AutomigrateEnabled,
		db:                       opts.DB,
		gormDB:                   gormDB,
		dbDriverName:             dbDriverName,
		debugEnabled:             opts.DebugEnabled,
		cryptoConfig:             cryptoConfig,
		parallelThreshold:        opts.ParallelThreshold,
		passwordAllowEmpty:       opts.PasswordAllowEmpty,
		passwordMinLength:        opts.PasswordMinLength,
		passwordRequireLowercase: opts.PasswordRequireLowercase,
		passwordRequireUppercase: opts.PasswordRequireUppercase,
		passwordRequireNumbers:   opts.PasswordRequireNumbers,
		passwordRequireSymbols:   opts.PasswordRequireSymbols,
	}

	if store.automigrateEnabled {
		err := store.AutoMigrate()
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
