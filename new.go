package vaultstore

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"

	"github.com/dracory/neat"
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
func NewStore(opts NewStoreOptions) (StoreInterface, error) {
	if opts.VaultTableName == "" {
		return nil, errors.New("vault store: vaultTableName is required")
	}

	if opts.VaultMetaTableName == "" {
		return nil, errors.New("vault store: vaultMetaTableName is required")
	}

	if opts.DB == nil {
		return nil, errors.New("vault store: DB is required")
	}

	// Set crypto config with secure defaults
	cryptoConfig := opts.CryptoConfig
	if cryptoConfig == nil {
		cryptoConfig = DefaultCryptoConfig()
	}

	neatDB, err := neat.NewFromSQLDB(opts.DB)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	store := &storeImplementation{
		vaultTableName:           opts.VaultTableName,
		vaultMetaTableName:       opts.VaultMetaTableName,
		automigrateEnabled:       opts.AutomigrateEnabled,
		debugEnabled:             opts.DebugEnabled,
		cryptoConfig:             cryptoConfig,
		parallelThreshold:        opts.ParallelThreshold,
		passwordAllowEmpty:       opts.PasswordAllowEmpty,
		passwordMinLength:        opts.PasswordMinLength,
		passwordRequireLowercase: opts.PasswordRequireLowercase,
		passwordRequireUppercase: opts.PasswordRequireUppercase,
		passwordRequireNumbers:   opts.PasswordRequireNumbers,
		passwordRequireSymbols:   opts.PasswordRequireSymbols,
		db:                       neatDB,
		dbDriverName:             opts.DbDriverName,
		logger:                   logger,
	}

	if store.automigrateEnabled {
		err := store.MigrateUp(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}
