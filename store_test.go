package vaultstore

import (
	"context"
	"errors"
	"testing"

	"database/sql"

	_ "modernc.org/sqlite"
)

func initDB() (*sql.DB, error) {
	dsn := ":memory:?parseTime=true"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func initStore() (StoreInterface, error) {
	db, err := initDB()
	if err != nil {
		return nil, err
	}

	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_token",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: true,
	})

	if err != nil {
		return nil, err
	}

	if store == nil {
		return nil, errors.New("unexpected nil store")
	}

	return store, nil
}

func TestWithAutoMigrateFalse(t *testing.T) {
	db, err := initDB()

	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	storeAutomigrateFalse, errAutomigrateFalse := NewStore(NewStoreOptions{
		VaultTableName:     "vault_with_automigrate_false",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: false,
	})

	if errAutomigrateFalse != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", errAutomigrateFalse.Error())
	}

	sImplFalse := storeAutomigrateFalse.(*storeImplementation)
	if sImplFalse.automigrateEnabled != false {
		t.Fatalf("automigrateEnabled: Expected [false] received [%v]", sImplFalse.automigrateEnabled)
	}

	storeAutomigrateTrue, errAutomigrateTrue := NewStore(NewStoreOptions{
		VaultTableName:     "vault_with_automigrate_true",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: true,
	})

	if errAutomigrateTrue != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", errAutomigrateTrue.Error())
	}

	sImplTrue := storeAutomigrateTrue.(*storeImplementation)
	if sImplTrue.automigrateEnabled != true {
		t.Fatalf("automigrateEnabled: Expected [true] received [%v]", sImplTrue.automigrateEnabled)
	}
}

func Test_Store_AutoMigrate(t *testing.T) {
	db, err := initDB()

	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_automigrate",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: false,
	})

	if err != nil {
		t.Fatalf("automigrateEnabled: Expected [err] to be nill received [%v]", err.Error())
	}

	sImpl := store.(*storeImplementation)
	if sImpl.automigrateEnabled != false {
		t.Fatalf("automigrateEnabled: Expected [false] received [%v]", sImpl.automigrateEnabled)
	}

	err = store.MigrateUp(context.Background())

	if err != nil {
		t.Fatalf("AutoMigrate Failure [%v]", err.Error())
	}

	if store.GetVaultTableName() != "vault_automigrate" {
		t.Fatalf("Expected vaultTableName [vault_automigrate] received [%v]", store.GetVaultTableName())
	}
	if sImpl.automigrateEnabled != false {
		t.Fatalf("Failure:  AutoMigrate")
	}
}

func Test_createRandomBlock(t *testing.T) {
	s := createRandomBlock(10)
	if len(s) != 10 {
		t.Fatalf("createRandomBlock Error")
	}

	s = createRandomBlock(50)
	if len(s) != 50 {
		t.Fatalf("createRandomBlock Error")
	}
}

func Test_calculateRequiredBlockLength(t *testing.T) {
	i := calculateRequiredBlockLength(1000)
	if i != 1024 {
		t.Fatalf("calculateRequiredBlockLength Error")
	}
}

func Test_base64Encode(t *testing.T) {
	str := "testing"
	s := base64Encode([]byte(str))
	if len(s) == 0 {
		t.Fatalf("base64Encode Failure")
	}
}

func Test_base64Decode(t *testing.T) {
	str := "testing"
	s := base64Encode([]byte(str))
	data, err := base64Decode(s)
	if err != nil {
		t.Fatalf("base64Decode Failure: err[%v]", err.Error())
	}
	if str != string(data) {
		t.Fatalf("base64Decode Failure")
	}
}

func Test_strToMD5Hash(t *testing.T) {
	ret := strToMD5Hash("testing")
	if len(ret) == 0 {
		t.Fatalf("strToMD5Hash Failure")
	}
}

func Test_strToSHA1Hash(t *testing.T) {
	ret := strToSHA1Hash("testing")
	if len(ret) == 0 {
		t.Fatalf("strToSHA1Hash Failure")
	}
}

func Test_strToSHA256Hash(t *testing.T) {
	ret := strToSHA256Hash("testing")
	if len(ret) == 0 {
		t.Fatalf("strToSHA256Hash Failure")
	}
}

func Test_isBase64(t *testing.T) {
	// Base64 of Hello -> SGVsbG8=
	ret := isBase64("SGVsbG8=")
	if !ret {
		t.Fatalf("isBase64 should ret TRUE, Failure")
	}

	ret = isBase64("Hello")
	if ret {
		t.Fatalf("isBase64 should ret FALSE, Failure")
	}
}

func Test_NewStore_Errors(t *testing.T) {
	// Test with empty table name
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: false,
	})

	if err == nil {
		t.Fatal("Expected error for empty table name but got nil")
	}
	if store != nil {
		t.Fatal("Expected nil store for empty table name")
	}

	// Test with empty meta table name
	store, err = NewStore(NewStoreOptions{
		VaultTableName:     "vault_test",
		VaultMetaTableName: "",
		DB:                 db,
		AutomigrateEnabled: false,
	})

	if err == nil {
		t.Fatal("Expected error for empty meta table name but got nil")
	}
	if store != nil {
		t.Fatal("Expected nil store for empty meta table name")
	}

	// Test with nil DB
	store, err = NewStore(NewStoreOptions{
		VaultTableName:     "vault_test",
		VaultMetaTableName: "vault_meta",
		DB:                 nil,
		AutomigrateEnabled: false,
	})

	if err == nil {
		t.Fatal("Expected error for nil DB but got nil")
	}
	if store != nil {
		t.Fatal("Expected nil store for nil DB")
	}
}

func Test_Store_EnableDebug(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_debug_test",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		AutomigrateEnabled: false,
		DebugEnabled:       false,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nil received [%v]", err.Error())
	}

	sImpl := store.(*storeImplementation)

	// Verify initial debug state
	if sImpl.debugEnabled != false {
		t.Fatalf("Expected debugEnabled to be false initially, got %v", sImpl.debugEnabled)
	}

	// Enable debug
	store.EnableDebug(true)
	if sImpl.debugEnabled != true {
		t.Fatalf("Expected debugEnabled to be true after enabling, got %v", sImpl.debugEnabled)
	}

	// Disable debug
	store.EnableDebug(false)
	if sImpl.debugEnabled != false {
		t.Fatalf("Expected debugEnabled to be false after disabling, got %v", sImpl.debugEnabled)
	}
}

func Test_Store_DbDriverName(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("initDB: Expected [err] to be nil received [%v]", err.Error())
	}

	// Test with explicit driver name
	store, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_driver_test",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		DbDriverName:       "sqlite",
		AutomigrateEnabled: false,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nil received [%v]", err.Error())
	}

	driverName := store.GetDbDriverName()
	if driverName != "sqlite" {
		t.Fatalf("Expected dbDriverName to be 'sqlite', got '%s'", driverName)
	}

	// Test with empty driver name
	store2, err := NewStore(NewStoreOptions{
		VaultTableName:     "vault_driver_test2",
		VaultMetaTableName: "vault_meta",
		DB:                 db,
		DbDriverName:       "",
		AutomigrateEnabled: false,
	})

	if err != nil {
		t.Fatalf("NewStore: Expected [err] to be nil received [%v]", err.Error())
	}

	driverName2 := store2.GetDbDriverName()
	if driverName2 != "" {
		t.Fatalf("Expected dbDriverName to be empty, got '%s'", driverName2)
	}
}
