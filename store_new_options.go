package vaultstore

import (
	"database/sql"
)

// NewStoreOptions define the options for creating a new session store
type NewStoreOptions struct {
	VaultTableName     string
	VaultMetaTableName string
	DB                 *sql.DB
	DbDriverName       string
	AutomigrateEnabled bool
	DebugEnabled       bool
	CryptoConfig       *CryptoConfig
	ParallelThreshold  int // Threshold for parallel processing (0 = use default 10000)
}
