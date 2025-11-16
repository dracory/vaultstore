package vaultstore

import (
	"context"
	"strings"
	"testing"
)

func Test_Store_TokenCreate(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenCreate: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()
	token, err := store.TokenCreate(ctx, "test_val", "test_pass", 20)

	if err != nil {
		t.Fatalf("ValueStore Failure: [%v]", err.Error())
	}

	if token == "" {
		t.Fatal("Token expected to not be empty")
	}

	if strings.HasPrefix(token, "tk_") == false {
		t.Fatal("Token expected to start with 'tk_' received: ", token)
	}

	if len(token) != 20 {
		t.Fatal("Token length expected to be 20 received: ", len(token), " token: ", token)
	}
}

func Test_Store_TokenCreateCustom(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenCreateCustom: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()
	err = store.TokenCreateCustom(ctx, "token_custom", "test_val", "test_pass")

	if err != nil {
		t.Fatalf("vault store: Expected [err] to be nil received [%v]", err.Error())
	}

	value, err := store.TokenRead(ctx, "token_custom", "test_pass")

	if err != nil {
		t.Fatalf("vault store: Expected [err] to be nil received [%v]", err.Error())
	}

	if value != "test_val" {
		t.Fatalf("vault store: Expected [value] to be 'test_val' received [%v]", value)
	}
}

func Test_Store_TokenDelete(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_ValueDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()
	token, err := store.TokenCreate(ctx, "test_val", "test_pass", 20)
	if err != nil {
		t.Fatalf("ValueStore Failure: [%v]", err.Error())
	}

	err = store.TokenDelete(ctx, token)
	if err != nil {
		t.Fatal("Test_Store_TokenDelete: Expected [err] to be nil received " + err.Error())
	}

	record, err := store.RecordFindByToken(ctx, token)

	if err != nil {
		t.Fatalf("Test_Store_TokenDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	if record != nil {
		t.Fatalf("Test_Store_TokenDelete: Expected [record] to be nil received [%v]", record)
	}
}

func TestTokenExists(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("TestTokenExists: Expected [err] to be nil received [%v]", err.Error())
	}

	token := "token1"

	ctx := context.Background()
	exists, err := store.TokenExists(ctx, token)

	if err != nil {
		t.Fatal(err)
	}

	if exists {
		t.Fatal("token should not exist")
	}

	err = store.TokenCreateCustom(ctx, token, "value1", "password")

	if err != nil {
		t.Fatal(err)
	}

	exists, err = store.TokenExists(ctx, token)

	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("token should exist")
	}
}

func Test_Store_TokenRead(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenRead: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()
	id, err := store.TokenCreate(ctx, "test_val", "test_pass", 20)

	if err != nil {
		t.Fatal("ValueStore Failure: ", err.Error())
	}

	val, err := store.TokenRead(ctx, id, "test_pass")
	if err != nil {
		t.Fatal("ValueRead Failure: ", err.Error())
	}

	if val != "test_val" {
		t.Fatal("ValueRetrieve Incorrect val: ", val)
	}
}

func Test_Store_TokenUpdate(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenUpdate: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()
	token, err := store.TokenCreate(ctx, "test_val", "test_pass", 20)

	if err != nil {
		t.Fatal("TokenCreate Failure: ", err.Error())
	}

	val, err := store.TokenRead(ctx, token, "test_pass")
	if err != nil {
		t.Fatal("TokenRead Failure: ", err.Error())
	}

	if val != "test_val" {
		t.Fatal("TokenRead Incorrect val: ", val)
	}

	err = store.TokenUpdate(ctx, token, "test_val2", "test_pass")

	if err != nil {
		t.Fatal("TokenUpdate Failure: ", err.Error())
	}

	val, err = store.TokenRead(ctx, token, "test_pass")

	if err != nil {
		t.Fatal("TokenRead Failure: ", err.Error())
	}

	if val != "test_val2" {
		t.Fatal("TokenRead Incorrect val: ", val)
	}
}

func Test_TokensRead(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_TokensRead: Expected [err] to be nil received [%v]", err.Error())
	}

	values := []string{"value1", "value2", "value3"}
	tokens := []string{"", "", ""}

	ctx := context.Background()
	for i := 0; i < len(values); i++ {
		token, err := store.TokenCreate(ctx, values[i], "test_pass", 20)

		if err != nil {
			t.Fatal("ValueStore Failure: ", err.Error())
		}

		tokens[i] = token
	}

	vals, err := store.TokensRead(ctx, tokens, "test_pass")

	if err != nil {
		t.Fatal("ValueRead Failure: ", err.Error())
	}

	for i := 0; i < len(values); i++ {
		if vals[tokens[i]] != values[i] {
			t.Fatal("ValueRetrieve Incorrect val: ", vals[tokens[i]])
		}
	}
}

func Test_Store_TokenSoftDelete(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty token
	err = store.TokenSoftDelete(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_TokenSoftDelete: Expected error for empty token but got nil")
	}

	// Create a token
	token, err := store.TokenCreate(ctx, "test_val_soft_delete", "test_pass", 20)
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Failed to create token: [%v]", err.Error())
	}

	// Verify token exists
	exists, err := store.TokenExists(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if !exists {
		t.Fatal("Test_Store_TokenSoftDelete: Expected token to exist before soft delete")
	}

	// Soft delete the token
	err = store.TokenSoftDelete(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	// Verify token no longer exists after soft delete
	exists, err = store.TokenExists(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if exists {
		t.Fatal("Test_Store_TokenSoftDelete: Expected token to not exist after soft delete")
	}

	// Verify record is not found with default query
	record, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if record != nil {
		t.Fatal("Test_Store_TokenSoftDelete: Expected not to find soft deleted record but found it")
	}

	// Verify record can be found when including soft deleted
	query := RecordQuery().SetToken(token).SetSoftDeletedInclude(true)
	records, err := store.RecordList(ctx, query)
	if err != nil {
		t.Fatalf("Test_Store_TokenSoftDelete: Failed to list records with soft deleted: [%v]", err.Error())
	}
	if len(records) != 1 {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected to find 1 soft deleted record but found %d", len(records))
	}
	if records[0].GetToken() != token {
		t.Fatalf("Test_Store_TokenSoftDelete: Expected Token [%s] but got [%s]", token, records[0].GetToken())
	}

	// Test with non-existent token
	err = store.TokenSoftDelete(ctx, "non_existent_token")
	if err == nil {
		t.Fatal("Test_Store_TokenSoftDelete: Expected error for non-existent token but got nil")
	}
}
