package vaultstore

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/dracory/neat"
	contractsorm "github.com/dracory/neat/contracts/database/orm"
	contractsschema "github.com/dracory/neat/contracts/database/schema"
	"github.com/samber/lo"
)

// Store defines a session store
type storeImplementation struct {
	vaultTableName           string
	vaultMetaTableName       string
	db                       *neat.Database
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
	logger                   *slog.Logger
}

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

// AutoMigrate auto migrate (deprecated - use MigrateUp)
func (store *storeImplementation) AutoMigrate() error {
	return store.MigrateUp(context.Background())
}

// MigrateUp creates the vault and meta tables
func (store *storeImplementation) MigrateUp(ctx context.Context, tx ...*sql.Tx) error {
	if store.db.Schema().HasTable(store.vaultTableName) {
		if store.debugEnabled {
			store.logger.Info("MigrateUp: table already exists", "table", store.vaultTableName)
		}
	} else {
		err := store.db.Schema().Create(store.vaultTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 40)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_VAULT_TOKEN, 40)
			table.Unique(COLUMN_VAULT_TOKEN)
			table.Text(COLUMN_VAULT_VALUE)
			table.DateTime(COLUMN_CREATED_AT)
			table.DateTime(COLUMN_UPDATED_AT)
			table.DateTime(COLUMN_EXPIRES_AT)
			table.DateTime(COLUMN_SOFT_DELETED_AT)
		})
		if err != nil {
			if store.debugEnabled {
				store.logger.Error("MigrateUp failed", "table", store.vaultTableName, "error", err)
			}
			return err
		}
	}

	if store.db.Schema().HasTable(store.vaultMetaTableName) {
		if store.debugEnabled {
			store.logger.Info("MigrateUp: table already exists", "table", store.vaultMetaTableName)
		}
	} else {
		err := store.db.Schema().Create(store.vaultMetaTableName, func(table contractsschema.Blueprint) {
			table.String(COLUMN_ID, 40)
			table.Primary(COLUMN_ID)
			table.String(COLUMN_OBJECT_TYPE, 50)
			table.String(COLUMN_OBJECT_ID, 64)
			table.String(COLUMN_META_KEY, 50)
			table.Text(COLUMN_META_VALUE)
		})
		if err != nil {
			if store.debugEnabled {
				store.logger.Error("MigrateUp failed", "table", store.vaultMetaTableName, "error", err)
			}
			return err
		}
	}

	return nil
}

// MigrateDown drops the vault and meta tables
func (store *storeImplementation) MigrateDown(ctx context.Context, tx ...*sql.Tx) error {
	if store.db.Schema().HasTable(store.vaultMetaTableName) {
		err := store.db.Schema().Drop(store.vaultMetaTableName)
		if err != nil {
			if store.debugEnabled {
				store.logger.Error("MigrateDown failed", "table", store.vaultMetaTableName, "error", err)
			}
			return err
		}
	}

	if store.db.Schema().HasTable(store.vaultTableName) {
		err := store.db.Schema().Drop(store.vaultTableName)
		if err != nil {
			if store.debugEnabled {
				store.logger.Error("MigrateDown failed", "table", store.vaultTableName, "error", err)
			}
			return err
		}
	}

	return nil
}

// EnableDebug - enables the debug option
func (store *storeImplementation) EnableDebug(debug bool) {
	store.debugEnabled = debug
	if debug {
		store.db.EnableDebug()
		store.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		store.db.DisableDebug()
		store.logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
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

func (store *storeImplementation) SetVaultTableName(tableName string) {
	store.vaultTableName = tableName
}

func (store *storeImplementation) SetMetaTableName(tableName string) {
	store.vaultMetaTableName = tableName
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

// == QUERY BUILDER ============================================================

// buildQuery builds a neat query from the record query interface.
func (store *storeImplementation) buildQuery(query RecordQueryInterface) contractsorm.Query {
	// Use Model() to enable neat's automatic soft delete handling via SoftDeletesMaxDate
	q := store.db.Query().Model(&recordImplementation{})

	if query == nil {
		return q
	}

	if query.IsIDSet() && query.GetID() != "" {
		q = q.Where(COLUMN_ID+" = ?", query.GetID())
	}

	if query.IsTokenSet() && query.GetToken() != "" {
		q = q.Where(COLUMN_VAULT_TOKEN+" = ?", query.GetToken())
	}

	if query.IsIDInSet() && len(query.GetIDIn()) > 0 {
		args := make([]any, len(query.GetIDIn()))
		for i, id := range query.GetIDIn() {
			args[i] = id
		}
		q = q.WhereIn(COLUMN_ID, args)
	}

	if query.IsTokenInSet() && len(query.GetTokenIn()) > 0 {
		args := make([]any, len(query.GetTokenIn()))
		for i, token := range query.GetTokenIn() {
			args[i] = token
		}
		q = q.WhereIn(COLUMN_VAULT_TOKEN, args)
	}

	if query.IsLimitSet() && query.GetLimit() > 0 && !query.IsCountOnlySet() {
		q = q.Limit(query.GetLimit())
	}

	if query.IsOffsetSet() && query.GetOffset() > 0 && !query.IsCountOnlySet() {
		q = q.Offset(query.GetOffset())
	}

	if query.IsOrderBySet() && query.GetOrderBy() != "" {
		sortOrder := DESC
		if query.IsSortOrderSet() && query.GetSortOrder() != "" {
			sortOrder = query.GetSortOrder()
		}
		q = q.OrderBy(query.GetOrderBy(), sortOrder)
	}

	// Handle soft delete filtering via neat's automatic handling (SoftDeletesMaxDate)
	if query.IsSoftDeletedIncludeSet() && query.GetSoftDeletedInclude() {
		q = q.WithSoftDeleted()
	}

	return q
}
