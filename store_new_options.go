package vaultstore

import (
	"database/sql"
)

// NewStoreOptions define the options for creating a new session store
type NewStoreOptions struct {
	VaultTableName          string
	VaultMetaTableName      string
	DB                      *sql.DB
	DbDriverName            string
	AutomigrateEnabled      bool
	DebugEnabled            bool
	CryptoConfig            *CryptoConfig
	PasswordIdentityEnabled bool
}
