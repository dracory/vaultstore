package vaultstore

import (
	"context"
	"log"
	"log/slog"

	"database/sql"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlserver"
	"github.com/dracory/database"
)

// Store defines a session store
type storeImplementation struct {
	vaultTableName     string
	db                 *sql.DB
	dbDriverName       string
	automigrateEnabled bool
	debugEnabled       bool
	logger             *slog.Logger
}

var _ StoreInterface = (*storeImplementation)(nil) // verify it extends the interface

// AutoMigrate auto migrate
func (st *storeImplementation) AutoMigrate() error {
	sql := st.SqlCreateTable()

	if st.debugEnabled {
		log.Println(sql)
	}

	_, err := st.db.Exec(sql)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// EnableDebug - enables the debug option
func (st *storeImplementation) EnableDebug(debug bool) {
	st.debugEnabled = debug
}

func (st *storeImplementation) GetDbDriverName() string {
	return st.dbDriverName
}

func (st *storeImplementation) GetVaultTableName() string {
	return st.vaultTableName
}

func (st *storeImplementation) toQuerableContext(context context.Context) database.QueryableContext {
	if database.IsQueryableContext(context) {
		return context.(database.QueryableContext)
	}

	return database.Context(context, st.db)
}
