package vaultstore

import "github.com/dracory/sb"

// SqlCreateTable returns a SQL string for creating the setting table
func (store *storeImplementation) SqlCreateTable() string {
	sql := sb.NewBuilder(sb.DatabaseDriverName(store.db)).
		Table(store.vaultTableName).
		Column(sb.Column{
			Name:       COLUMN_ID,
			Type:       sb.COLUMN_TYPE_STRING,
			Length:     40,
			PrimaryKey: true,
		}).
		Column(sb.Column{
			Name:   COLUMN_VAULT_TOKEN,
			Type:   sb.COLUMN_TYPE_STRING,
			Length: 40,
			Unique: true,
		}).
		Column(sb.Column{
			Name: COLUMN_VAULT_VALUE,
			Type: sb.COLUMN_TYPE_LONGTEXT,
		}).
		Column(sb.Column{
			Name: COLUMN_CREATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_UPDATED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_EXPIRES_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		Column(sb.Column{
			Name: COLUMN_SOFT_DELETED_AT,
			Type: sb.COLUMN_TYPE_DATETIME,
		}).
		CreateIfNotExists()

	return sql
}
