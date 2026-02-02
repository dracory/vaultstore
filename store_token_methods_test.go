package vaultstore

import (
	"context"
	"strings"
	"testing"
	"time"
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

func Test_Store_TokenCreateWithExpiration(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenCreateWithExpiration: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create token that expires in 1 hour
	expireTime := time.Now().UTC().Add(1 * time.Hour)
	token, err := store.TokenCreate(ctx, "test_val", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})

	if err != nil {
		t.Fatalf("TokenCreate with expiration failed: [%v]", err.Error())
	}

	if token == "" {
		t.Fatal("Token expected to not be empty")
	}

	// Verify token can be read
	val, err := store.TokenRead(ctx, token, "test_pass")
	if err != nil {
		t.Fatal("TokenRead failed: ", err.Error())
	}

	if val != "test_val" {
		t.Fatal("TokenRead incorrect value: ", val)
	}

	// Verify expiration was set
	record, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatal("Failed to find record: ", err.Error())
	}

	if record == nil {
		t.Fatal("Record not found")
	}

	expiresAt := record.GetExpiresAt()
	if expiresAt == "" {
		t.Fatal("ExpiresAt should not be empty")
	}
}

func Test_Store_TokenCreateWithExpiration_Expired(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenCreateWithExpiration_Expired: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create token that expires immediately (in the past)
	expireTime := time.Now().UTC().Add(-1 * time.Second)
	token, err := store.TokenCreate(ctx, "expired_val", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})

	if err != nil {
		t.Fatalf("TokenCreate with past expiration failed: [%v]", err.Error())
	}

	// Verify token cannot be read (returns ErrTokenExpired)
	_, err = store.TokenRead(ctx, token, "test_pass")
	if err != ErrTokenExpired {
		t.Fatalf("Expected ErrTokenExpired but got: %v", err)
	}
}

func Test_Store_TokenCreateCustomWithExpiration(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenCreateCustomWithExpiration: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create custom token that expires in 1 hour
	expireTime := time.Now().UTC().Add(1 * time.Hour)
	err = store.TokenCreateCustom(ctx, "custom_expiring_token", "test_val", "test_pass", TokenCreateOptions{
		ExpiresAt: expireTime,
	})

	if err != nil {
		t.Fatalf("TokenCreateCustom with expiration failed: [%v]", err.Error())
	}

	// Verify token can be read
	val, err := store.TokenRead(ctx, "custom_expiring_token", "test_pass")
	if err != nil {
		t.Fatal("TokenRead failed: ", err.Error())
	}

	if val != "test_val" {
		t.Fatal("TokenRead incorrect value: ", val)
	}
}

func Test_Store_TokenRead_Expired(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenRead_Expired: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create token that expired 1 second ago
	expireTime := time.Now().UTC().Add(-1 * time.Second)
	token, err := store.TokenCreate(ctx, "expired_val", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})

	if err != nil {
		t.Fatalf("TokenCreate failed: [%v]", err.Error())
	}

	// Try to read expired token
	_, err = store.TokenRead(ctx, token, "test_pass")
	if err != ErrTokenExpired {
		t.Fatalf("Expected ErrTokenExpired but got: %v", err)
	}
}

func Test_TokensRead_SkipsExpired(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_TokensRead_SkipsExpired: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create valid token
	validToken, err := store.TokenCreate(ctx, "valid_value", "test_pass", 20)
	if err != nil {
		t.Fatal("Failed to create valid token: ", err.Error())
	}

	// Create expired token
	expireTime := time.Now().UTC().Add(-1 * time.Second)
	expiredToken, err := store.TokenCreate(ctx, "expired_value", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})
	if err != nil {
		t.Fatal("Failed to create expired token: ", err.Error())
	}

	// Read both tokens
	tokens := []string{validToken, expiredToken}
	vals, err := store.TokensRead(ctx, tokens, "test_pass")

	// Function returns partial map with only valid tokens, no error
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	// Verify only valid token is in the map
	if len(vals) != 1 {
		t.Fatalf("Expected 1 value in map but got %d", len(vals))
	}

	if vals[validToken] != "valid_value" {
		t.Fatal("Valid token value incorrect")
	}

	if _, exists := vals[expiredToken]; exists {
		t.Fatal("Expired token should not be in the map")
	}
}

func Test_Store_TokenRenew(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenRenew: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create token that expires in 1 second
	expireTime := time.Now().UTC().Add(1 * time.Second)
	token, err := store.TokenCreate(ctx, "renewable_val", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})

	if err != nil {
		t.Fatalf("TokenCreate failed: [%v]", err.Error())
	}

	// Renew token to expire in 1 hour
	newExpireTime := time.Now().UTC().Add(1 * time.Hour)
	err = store.TokenRenew(ctx, token, newExpireTime)
	if err != nil {
		t.Fatalf("TokenRenew failed: [%v]", err.Error())
	}

	// Verify token can still be read
	val, err := store.TokenRead(ctx, token, "test_pass")
	if err != nil {
		t.Fatalf("TokenRead after renew failed: %v", err)
	}

	if val != "renewable_val" {
		t.Fatal("TokenRead incorrect value after renew")
	}

	// Renew to never expire (zero time)
	err = store.TokenRenew(ctx, token, time.Time{})
	if err != nil {
		t.Fatalf("TokenRenew to no-expiration failed: [%v]", err.Error())
	}

	// Verify expiration is now MAX_DATETIME
	record, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatal("Failed to find record after renew: ", err.Error())
	}

	if record.GetExpiresAt() == "" {
		t.Fatal("ExpiresAt should not be empty after renew")
	}
}

func Test_Store_TokenRenew_NonExistent(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokenRenew_NonExistent: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Try to renew non-existent token
	expireTime := time.Now().UTC().Add(1 * time.Hour)
	err = store.TokenRenew(ctx, "non_existent_token", expireTime)
	if err == nil {
		t.Fatal("Expected error for non-existent token")
	}
}

func Test_Store_TokensExpiredSoftDelete(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokensExpiredSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create expired token
	expireTime := time.Now().UTC().Add(-1 * time.Second)
	token1, err := store.TokenCreate(ctx, "expired_val1", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})
	if err != nil {
		t.Fatalf("Failed to create expired token: [%v]", err.Error())
	}

	// Create another expired token
	token2, err := store.TokenCreate(ctx, "expired_val2", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})
	if err != nil {
		t.Fatalf("Failed to create second expired token: [%v]", err.Error())
	}

	// Create valid token
	validToken, err := store.TokenCreate(ctx, "valid_val", "test_pass", 20)
	if err != nil {
		t.Fatalf("Failed to create valid token: [%v]", err.Error())
	}

	// Soft delete expired tokens
	count, err := store.TokensExpiredSoftDelete(ctx)
	if err != nil {
		t.Fatalf("TokensExpiredSoftDelete failed: [%v]", err.Error())
	}

	if count != 2 {
		t.Fatalf("Expected 2 expired tokens soft deleted, got %d", count)
	}

	// Verify expired tokens are soft deleted
	exists, _ := store.TokenExists(ctx, token1)
	if exists {
		t.Fatal("Expired token1 should not exist after soft delete")
	}

	exists, _ = store.TokenExists(ctx, token2)
	if exists {
		t.Fatal("Expired token2 should not exist after soft delete")
	}

	// Verify valid token still exists
	exists, _ = store.TokenExists(ctx, validToken)
	if !exists {
		t.Fatal("Valid token should still exist")
	}
}

func Test_Store_TokensExpiredDelete(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokensExpiredDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create expired token
	expireTime := time.Now().UTC().Add(-1 * time.Second)
	token1, err := store.TokenCreate(ctx, "expired_val1", "test_pass", 20, TokenCreateOptions{
		ExpiresAt: expireTime,
	})
	if err != nil {
		t.Fatalf("Failed to create expired token: [%v]", err.Error())
	}

	// Create valid token
	validToken, err := store.TokenCreate(ctx, "valid_val", "test_pass", 20)
	if err != nil {
		t.Fatalf("Failed to create valid token: [%v]", err.Error())
	}

	// Permanently delete expired tokens
	count, err := store.TokensExpiredDelete(ctx)
	if err != nil {
		t.Fatalf("TokensExpiredDelete failed: [%v]", err.Error())
	}

	if count != 1 {
		t.Fatalf("Expected 1 expired token deleted, got %d", count)
	}

	// Verify expired token is permanently deleted
	record, err := store.RecordFindByToken(ctx, token1)
	if err != nil {
		t.Fatalf("Error finding record: [%v]", err.Error())
	}
	if record != nil {
		t.Fatal("Expired token should be permanently deleted")
	}

	// Verify valid token still exists
	exists, _ := store.TokenExists(ctx, validToken)
	if !exists {
		t.Fatal("Valid token should still exist")
	}
}

func Test_Store_TokensExpired_NoExpiration(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_TokensExpired_NoExpiration: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create token with no expiration (default)
	token, err := store.TokenCreate(ctx, "no_expire_val", "test_pass", 20)
	if err != nil {
		t.Fatalf("Failed to create token: [%v]", err.Error())
	}

	// Soft delete expired tokens - should not delete the non-expiring token
	count, err := store.TokensExpiredSoftDelete(ctx)
	if err != nil {
		t.Fatalf("TokensExpiredSoftDelete failed: [%v]", err.Error())
	}

	if count != 0 {
		t.Fatalf("Expected 0 tokens deleted, got %d", count)
	}

	// Verify token still exists
	exists, _ := store.TokenExists(ctx, token)
	if !exists {
		t.Fatal("Non-expiring token should still exist")
	}

	// Also test hard delete
	count, err = store.TokensExpiredDelete(ctx)
	if err != nil {
		t.Fatalf("TokensExpiredDelete failed: [%v]", err.Error())
	}

	if count != 0 {
		t.Fatalf("Expected 0 tokens deleted, got %d", count)
	}

	// Verify token still exists
	exists, _ = store.TokenExists(ctx, token)
	if !exists {
		t.Fatal("Non-expiring token should still exist after hard delete attempt")
	}
}
