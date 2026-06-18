package vaultstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dracory/database"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewStoreOptions define the options for creating a new session store
type NewStoreOptions struct {
	VaultTableName           string
	VaultMetaTableName       string
	DB                       *sql.DB
	DbDriverName             string
	AutomigrateEnabled       bool
	DebugEnabled             bool
	CryptoConfig             *CryptoConfig
	ParallelThreshold        int  // Threshold for parallel processing (0 = use default 10000)
	PasswordAllowEmpty       bool // Allow empty passwords (default: false)
	PasswordMinLength        int  // Minimum password length (default: 16)
	PasswordRequireLowercase bool // Require at least one lowercase letter (default: false)
	PasswordRequireUppercase bool // Require at least one uppercase letter (default: false)
	PasswordRequireNumbers   bool // Require at least one number (default: false)
	PasswordRequireSymbols   bool // Require at least one symbol (default: false)
}

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
		err := store.MigrateUp(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
