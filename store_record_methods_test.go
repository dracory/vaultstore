package vaultstore

import (
	"context"
	"testing"
)

func Test_Store_RecordCount(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatalf("Test_Store_RecordCount: Expected [err] to be nil received [%v]", err.Error())
	}

	record := NewRecord().SetToken("test_token1").SetValue("test_value")

	ctx := context.Background()
	err = store.RecordCreate(ctx, record)
	if err != nil {
		t.Fatalf("Test_Store_RecordCount Failure: [%v]", err.Error())
	}

	record = NewRecord().SetToken("test_token2").SetValue("test_value")

	err = store.RecordCreate(context.Background(), record)
	if err != nil {
		t.Fatalf("Test_Store_RecordCount Failure: [%v]", err.Error())
	}

	record = NewRecord().SetToken("test_token3").SetValue("test_value")

	err = store.RecordCreate(ctx, record)
	if err != nil {
		t.Fatalf("Test_Store_RecordCount Failure: [%v]", err.Error())
	}

	count, err := store.RecordCount(ctx, RecordQuery().SetToken("test_token2"))

	if err != nil {
		t.Fatal("Test_Store_RecordCount: Expected [err] to be nil received " + err.Error())
	}

	if count != 1 {
		t.Fatalf("Test_Store_RecordCount: Expected [count] to be 1 received [%v]", count)
	}

}

func Test_Store_RecordCreate(t *testing.T) {
	store, err := initStore()

	if err != nil {
		t.Fatal("Test_Store_RecordCreate: Expected [err] to be nil received: ", err.Error())
	}

	record := NewRecord().SetToken("test_token").SetValue("test_value")

	ctx := context.Background()
	err = store.RecordCreate(ctx, record)
	if err != nil {
		t.Fatal("Test_Store_RecordCreate: Expected [err] to be nil received " + err.Error())
	}

	exists, err := store.TokenExists(ctx, "test_token")

	if err != nil {
		t.Fatal(err)
	}

	if !exists {
		t.Fatal("Test_Store_RecordCreate: token should exist")
	}
}

func Test_Store_RecordFindByID(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty ID
	record, err := store.RecordFindByID(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_RecordFindByID: Expected error for empty ID but got nil")
	}
	if record != nil {
		t.Fatal("Test_Store_RecordFindByID: Expected nil record for empty ID")
	}

	// Create a record
	newRecord := NewRecord().SetToken("test_token_find_by_id").SetValue("test_value_find_by_id")
	recordID := newRecord.GetID()

	err = store.RecordCreate(ctx, newRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Failed to create record: [%v]", err.Error())
	}

	// Find by ID
	foundRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordFindByID: Expected to find record but got nil")
	}
	if foundRecord.GetID() != recordID {
		t.Fatalf("Test_Store_RecordFindByID: Expected ID [%s] but got [%s]", recordID, foundRecord.GetID())
	}
	if foundRecord.GetToken() != "test_token_find_by_id" {
		t.Fatalf("Test_Store_RecordFindByID: Expected Token [test_token_find_by_id] but got [%s]", foundRecord.GetToken())
	}
	if foundRecord.GetValue() != "test_value_find_by_id" {
		t.Fatalf("Test_Store_RecordFindByID: Expected Value [test_value_find_by_id] but got [%s]", foundRecord.GetValue())
	}

	// Soft delete the record
	err = store.RecordSoftDeleteByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Failed to soft delete record: [%v]", err.Error())
	}

	// Verify record is not found with default query after soft delete
	softDeletedRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if softDeletedRecord != nil {
		t.Fatal("Test_Store_RecordFindByID: Expected not to find soft deleted record but found it")
	}

	// Verify record can be found when including soft deleted
	query := RecordQuery().SetID(recordID).SetSoftDeletedInclude(true)
	records, err := store.RecordList(ctx, query)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Failed to list records with soft deleted: [%v]", err.Error())
	}
	if len(records) != 1 {
		t.Fatalf("Test_Store_RecordFindByID: Expected to find 1 soft deleted record but found %d", len(records))
	}
	if records[0].GetID() != recordID {
		t.Fatalf("Test_Store_RecordFindByID: Expected ID [%s] but got [%s]", recordID, records[0].GetID())
	}

	// Test with non-existent ID
	nonExistentRecord, err := store.RecordFindByID(ctx, "non_existent_id")
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByID: Expected [err] to be nil for non-existent ID but got [%v]", err.Error())
	}
	if nonExistentRecord != nil {
		t.Fatal("Test_Store_RecordFindByID: Expected nil record for non-existent ID")
	}
}

func Test_Store_RecordFindByToken(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByToken: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty token
	record, err := store.RecordFindByToken(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_RecordFindByToken: Expected error for empty token but got nil")
	}
	if record != nil {
		t.Fatal("Test_Store_RecordFindByToken: Expected nil record for empty token")
	}

	// Create a record
	token := "test_token_find"
	newRecord := NewRecord().SetToken(token).SetValue("test_value_find")

	err = store.RecordCreate(ctx, newRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByToken: Failed to create record: [%v]", err.Error())
	}

	// Find by token
	foundRecord, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByToken: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordFindByToken: Expected to find record but got nil")
	}
	if foundRecord.GetToken() != token {
		t.Fatalf("Test_Store_RecordFindByToken: Expected Token [%s] but got [%s]", token, foundRecord.GetToken())
	}
	if foundRecord.GetValue() != "test_value_find" {
		t.Fatalf("Test_Store_RecordFindByToken: Expected Value [test_value_find] but got [%s]", foundRecord.GetValue())
	}

	// Test with non-existent token
	nonExistentRecord, err := store.RecordFindByToken(ctx, "non_existent_token")
	if err != nil {
		t.Fatalf("Test_Store_RecordFindByToken: Expected [err] to be nil for non-existent token but got [%v]", err.Error())
	}
	if nonExistentRecord != nil {
		t.Fatal("Test_Store_RecordFindByToken: Expected nil record for non-existent token")
	}
}

func Test_Store_RecordList(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordList: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create multiple records
	for i := 1; i <= 5; i++ {
		token := "test_token_list_" + string(rune('A'+i-1))
		record := NewRecord().SetToken(token).SetValue("test_value_" + string(rune('A'+i-1)))

		err = store.RecordCreate(ctx, record)
		if err != nil {
			t.Fatalf("Test_Store_RecordList: Failed to create record: [%v]", err.Error())
		}
	}

	// Test listing all records
	records, err := store.RecordList(ctx, RecordQuery())
	if err != nil {
		t.Fatalf("Test_Store_RecordList: Expected [err] to be nil received [%v]", err.Error())
	}
	if len(records) != 5 {
		t.Fatalf("Test_Store_RecordList: Expected 5 records but got %d", len(records))
	}

	// Test with token filter
	filteredRecords, err := store.RecordList(ctx, RecordQuery().SetToken("test_token_list_C"))
	if err != nil {
		t.Fatalf("Test_Store_RecordList: Expected [err] to be nil received [%v]", err.Error())
	}
	if len(filteredRecords) != 1 {
		t.Fatalf("Test_Store_RecordList: Expected 1 record with token filter but got %d", len(filteredRecords))
	}
	if filteredRecords[0].GetToken() != "test_token_list_C" {
		t.Fatalf("Test_Store_RecordList: Expected Token [test_token_list_C] but got [%s]", filteredRecords[0].GetToken())
	}

	// Test with limit
	limitedRecords, err := store.RecordList(ctx, RecordQuery().SetLimit(2))
	if err != nil {
		t.Fatalf("Test_Store_RecordList: Expected [err] to be nil received [%v]", err.Error())
	}
	if len(limitedRecords) != 2 {
		t.Fatalf("Test_Store_RecordList: Expected 2 records with limit but got %d", len(limitedRecords))
	}

	// Test with token list filter
	tokenListRecords, err := store.RecordList(ctx, RecordQuery().SetTokenIn([]string{"test_token_list_A", "test_token_list_E"}))
	if err != nil {
		t.Fatalf("Test_Store_RecordList: Expected [err] to be nil received [%v]", err.Error())
	}
	if len(tokenListRecords) != 2 {
		t.Fatalf("Test_Store_RecordList: Expected 2 records with token list filter but got %d", len(tokenListRecords))
	}
}

func Test_Store_RecordUpdate(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordUpdate: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Create a record
	record := NewRecord().SetToken("test_token_update").SetValue("test_value_original")
	recordID := record.GetID()

	err = store.RecordCreate(ctx, record)
	if err != nil {
		t.Fatalf("Test_Store_RecordUpdate: Failed to create record: [%v]", err.Error())
	}

	// Retrieve the record
	foundRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordUpdate: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordUpdate: Expected to find record but got nil")
	}

	// Update the record
	foundRecord.SetValue("test_value_updated")
	err = store.RecordUpdate(ctx, foundRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordUpdate: Failed to update record: [%v]", err.Error())
	}

	// Verify the update
	updatedRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordUpdate: Expected [err] to be nil received [%v]", err.Error())
	}
	if updatedRecord == nil {
		t.Fatal("Test_Store_RecordUpdate: Expected to find updated record but got nil")
	}
	if updatedRecord.GetValue() != "test_value_updated" {
		t.Fatalf("Test_Store_RecordUpdate: Expected updated Value [test_value_updated] but got [%s]", updatedRecord.GetValue())
	}

	// Test update with no changes
	err = store.RecordUpdate(ctx, updatedRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordUpdate: Expected [err] to be nil for no changes but got [%v]", err.Error())
	}
}

func Test_Store_RecordDeleteByID(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty ID
	err = store.RecordDeleteByID(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_RecordDeleteByID: Expected error for empty ID but got nil")
	}

	// Create a record
	record := NewRecord().SetToken("test_token_delete_id").SetValue("test_value_delete")
	recordID := record.GetID()

	err = store.RecordCreate(ctx, record)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByID: Failed to create record: [%v]", err.Error())
	}

	// Verify record exists
	foundRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordDeleteByID: Expected to find record but got nil")
	}

	// Delete the record
	err = store.RecordDeleteByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByID: Failed to delete record: [%v]", err.Error())
	}

	// Verify record is deleted
	deletedRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByID: Expected [err] to be nil after deletion but got [%v]", err.Error())
	}
	if deletedRecord != nil {
		t.Fatal("Test_Store_RecordDeleteByID: Expected nil record after deletion but got a record")
	}

	// Test deleting non-existent record
	err = store.RecordDeleteByID(ctx, "non_existent_id")
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByID: Expected [err] to be nil for non-existent ID but got [%v]", err.Error())
	}
}

func Test_Store_RecordDeleteByToken(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty token
	err = store.RecordDeleteByToken(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_RecordDeleteByToken: Expected error for empty token but got nil")
	}

	// Create a record
	token := "test_token_delete"
	record := NewRecord().SetToken(token).SetValue("test_value_delete")

	err = store.RecordCreate(ctx, record)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByToken: Failed to create record: [%v]", err.Error())
	}

	// Verify record exists
	foundRecord, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordDeleteByToken: Expected to find record but got nil")
	}

	// Delete the record
	err = store.RecordDeleteByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByToken: Failed to delete record: [%v]", err.Error())
	}

	// Verify record is deleted
	deletedRecord, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByToken: Expected [err] to be nil after deletion but got [%v]", err.Error())
	}
	if deletedRecord != nil {
		t.Fatal("Test_Store_RecordDeleteByToken: Expected nil record after deletion but got a record")
	}

	// Test deleting non-existent record
	err = store.RecordDeleteByToken(ctx, "non_existent_token")
	if err != nil {
		t.Fatalf("Test_Store_RecordDeleteByToken: Expected [err] to be nil for non-existent token but got [%v]", err.Error())
	}
}

func Test_Store_RecordSoftDelete(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with nil record
	err = store.RecordSoftDelete(ctx, nil)
	if err == nil {
		t.Fatal("Test_Store_RecordSoftDelete: Expected error for nil record but got nil")
	}

	// Create a record
	newRecord := NewRecord().SetToken("test_token_soft_delete").SetValue("test_value_soft_delete")
	recordID := newRecord.GetID()

	err = store.RecordCreate(ctx, newRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Failed to create record: [%v]", err.Error())
	}

	// Verify record exists
	foundRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordSoftDelete: Expected to find record but got nil")
	}

	// Verify token exists
	exists, err := store.TokenExists(ctx, "test_token_soft_delete")
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if !exists {
		t.Fatal("Test_Store_RecordSoftDelete: Expected token to exist before soft delete")
	}

	// Soft delete the record
	err = store.RecordSoftDelete(ctx, foundRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Failed to soft delete record: [%v]", err.Error())
	}

	// Verify record is not found with default query (which excludes soft deleted)
	softDeletedRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if softDeletedRecord != nil {
		t.Fatal("Test_Store_RecordSoftDelete: Expected not to find soft deleted record but found it")
	}

	// Verify token no longer exists after soft delete
	exists, err = store.TokenExists(ctx, "test_token_soft_delete")
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected [err] to be nil received [%v]", err.Error())
	}
	if exists {
		t.Fatal("Test_Store_RecordSoftDelete: Expected token to not exist after soft delete")
	}

	// Verify record can be found when including soft deleted
	query := RecordQuery().SetID(recordID).SetSoftDeletedInclude(true)
	records, err := store.RecordList(ctx, query)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDelete: Failed to list records with soft deleted: [%v]", err.Error())
	}
	if len(records) != 1 {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected to find 1 soft deleted record but found %d", len(records))
	}
	if records[0].GetID() != recordID {
		t.Fatalf("Test_Store_RecordSoftDelete: Expected ID [%s] but got [%s]", recordID, records[0].GetID())
	}
}

func Test_Store_RecordSoftDeleteByID(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty ID
	err = store.RecordSoftDeleteByID(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_RecordSoftDeleteByID: Expected error for empty ID but got nil")
	}

	// Create a record
	newRecord := NewRecord().SetToken("test_token_soft_delete_by_id").SetValue("test_value_soft_delete_by_id")
	recordID := newRecord.GetID()

	err = store.RecordCreate(ctx, newRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Failed to create record: [%v]", err.Error())
	}

	// Verify record exists
	foundRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordSoftDeleteByID: Expected to find record but got nil")
	}

	// Verify token exists
	exists, err := store.TokenExists(ctx, "test_token_soft_delete_by_id")
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if !exists {
		t.Fatal("Test_Store_RecordSoftDeleteByID: Expected token to exist before soft delete")
	}

	// Soft delete by ID
	err = store.RecordSoftDeleteByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Failed to soft delete record: [%v]", err.Error())
	}

	// Verify record is not found with default query
	softDeletedRecord, err := store.RecordFindByID(ctx, recordID)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if softDeletedRecord != nil {
		t.Fatal("Test_Store_RecordSoftDeleteByID: Expected not to find soft deleted record but found it")
	}

	// Verify token no longer exists after soft delete
	exists, err = store.TokenExists(ctx, "test_token_soft_delete_by_id")
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByID: Expected [err] to be nil received [%v]", err.Error())
	}
	if exists {
		t.Fatal("Test_Store_RecordSoftDeleteByID: Expected token to not exist after soft delete")
	}

	// Test with non-existent ID
	err = store.RecordSoftDeleteByID(ctx, "non_existent_id")
	if err == nil {
		t.Fatal("Test_Store_RecordSoftDeleteByID: Expected error for non-existent ID but got nil")
	}
}

func Test_Store_RecordSoftDeleteByToken(t *testing.T) {
	store, err := initStore()
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}

	ctx := context.Background()

	// Test with empty token
	err = store.RecordSoftDeleteByToken(ctx, "")
	if err == nil {
		t.Fatal("Test_Store_RecordSoftDeleteByToken: Expected error for empty token but got nil")
	}

	// Create a record
	token := "test_token_soft_delete_by_token"
	newRecord := NewRecord().SetToken(token).SetValue("test_value_soft_delete_by_token")
	recordID := newRecord.GetID()

	err = store.RecordCreate(ctx, newRecord)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Failed to create record: [%v]", err.Error())
	}

	// Verify record exists
	foundRecord, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}
	if foundRecord == nil {
		t.Fatal("Test_Store_RecordSoftDeleteByToken: Expected to find record but got nil")
	}

	// Verify token exists
	exists, err := store.TokenExists(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}
	if !exists {
		t.Fatal("Test_Store_RecordSoftDeleteByToken: Expected token to exist before soft delete")
	}

	// Soft delete by token
	err = store.RecordSoftDeleteByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Failed to soft delete record: [%v]", err.Error())
	}

	// Verify record is not found with default query
	softDeletedRecord, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}
	if softDeletedRecord != nil {
		t.Fatal("Test_Store_RecordSoftDeleteByToken: Expected not to find soft deleted record but found it")
	}

	// Verify token no longer exists after soft delete
	exists, err = store.TokenExists(ctx, token)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected [err] to be nil received [%v]", err.Error())
	}
	if exists {
		t.Fatal("Test_Store_RecordSoftDeleteByToken: Expected token to not exist after soft delete")
	}

	// Verify record can be found when including soft deleted
	query := RecordQuery().SetToken(token).SetSoftDeletedInclude(true)
	records, err := store.RecordList(ctx, query)
	if err != nil {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Failed to list records with soft deleted: [%v]", err.Error())
	}
	if len(records) != 1 {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected to find 1 soft deleted record but found %d", len(records))
	}
	if records[0].GetToken() != token {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected Token [%s] but got [%s]", token, records[0].GetToken())
	}
	if records[0].GetID() != recordID {
		t.Fatalf("Test_Store_RecordSoftDeleteByToken: Expected ID [%s] but got [%s]", recordID, records[0].GetID())
	}

	// Test with non-existent token
	err = store.RecordSoftDeleteByToken(ctx, "non_existent_token")
	if err == nil {
		t.Fatal("Test_Store_RecordSoftDeleteByToken: Expected error for non-existent token but got nil")
	}
}
