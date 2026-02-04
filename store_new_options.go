package vaultstore

import (
	"database/sql"
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
