package vaultstore

import (
	"context"
	"errors"
	"strings"

	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
)

// TokenCreate creates a new record and returns the token
func (st *storeImplementation) TokenCreate(ctx context.Context, data string, password string, tokenLength int) (token string, err error) {
	token, err = generateToken(tokenLength)

	if err != nil {
		return "", err
	}

	encodedData := encode(data, password)

	var newEntry = NewRecord().
		SetToken(token).
		SetValue(encodedData).
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	err = st.RecordCreate(ctx, newEntry)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (store *storeImplementation) TokenCreateCustom(ctx context.Context, token string, data string, password string) (err error) {
	encodedData := encode(data, password)

	var newEntry = NewRecord().
		SetToken(token).
		SetValue(encodedData).
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	err = store.RecordCreate(ctx, newEntry)

	if err != nil {
		return err
	}

	return nil
}

// TokenDelete deletes a token from the store
//
// # If the supplied token is empty, an error is returned
//
// Parameters:
// - ctx: The context
// - token: The token to delete
//
// Returns:
// - err: An error if something went wrong
func (st *storeImplementation) TokenDelete(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	return st.RecordDeleteByToken(ctx, token)
}

// TokenExists checks if a token exists
//
// # If the supplied token is empty, an error is returned
//
// Parameters:
// - ctx: The context
// - token: The token to check
//
// Returns:
// - exists: A boolean indicating if the token exists
// - err: An error if something went wrong
func (store *storeImplementation) TokenExists(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, errors.New("token is empty")
	}

	count, err := store.RecordCount(ctx, RecordQuery().SetToken(token))

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// TokenRead retrieves the value of a token
//
// # If the token does not exist, an error is returned
//
// Parameters:
// - ctx: The context
// - token: The token to retrieve
// - password: The password to use for decryption
//
// Returns:
// - value: The value of the token
// - err: An error if something went wrong
func (st *storeImplementation) TokenRead(ctx context.Context, token string, password string) (value string, err error) {
	entry, err := st.RecordFindByToken(ctx, token)

	if err != nil {
		return "", err
	}

	if entry == nil {
		return "", errors.New("token does not exist")
	}

	decoded, err := decode(entry.GetValue(), password)

	if err != nil {
		return "", err
	}

	return decoded, nil
}

// TokenSoftDelete soft deletes a token from the store
//
// Soft deleting keeps the record in the database but marks it
// as soft deleted and soft deleted records are not returned by default
//
// # If the supplied token is empty, an error is returned
//
// Parameters:
// - ctx: The context
// - token: The token to soft delete
//
// Returns:
// - err: An error if something went wrong
func (st *storeImplementation) TokenSoftDelete(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	return st.RecordSoftDeleteByToken(ctx, token)
}

// TokenUpdate updates the value of a token
//
// # If the token does not exist, an error is returned
//
// Parameters:
// - ctx: The context
// - token: The token to update
// - value: The new value
// - password: The password to use for encryption
//
// Returns:
// - err: An error if something went wrong
func (st *storeImplementation) TokenUpdate(ctx context.Context, token string, value string, password string) (err error) {
	entry, errFind := st.RecordFindByToken(ctx, token)

	if errFind != nil {
		return err
	}

	if entry == nil {
		return errors.New("token does not exist")
	}

	encodedValue := encode(value, password)

	entry.SetValue(encodedValue)

	return st.RecordUpdate(ctx, entry)
}

// TokensRead reads a list of tokens, returns a map of token to value
//
// # If a token is not found, it is not included in the map
//
// Parameters:
// - ctx: The context
// - tokens: The list of tokens to read
// - password: The password to use for decryption
//
// Returns:
// - values: A map of token to value
// - err: An error if something went wrong
func (st *storeImplementation) TokensRead(ctx context.Context, tokens []string, password string) (values map[string]string, err error) {
	values = map[string]string{}

	entries, err := st.RecordList(ctx, RecordQuery().SetTokenIn(tokens))

	if err != nil {
		return values, err
	}

	if len(entries) != len(tokens) {
		var entryTokens = lo.Map(entries, func(entry RecordInterface, _ int) string {
			return entry.GetToken()
		})

		_, missingTokens := lo.Difference(tokens, entryTokens)

		return values, errors.New("missing tokens: " + strings.Join(missingTokens, ", "))
	}

	for _, entry := range entries {
		decoded, err := decode(entry.GetValue(), password)

		if err != nil {
			return map[string]string{}, errors.New("decode error for token: " + entry.GetToken() + " : " + err.Error())
		}

		values[entry.GetToken()] = decoded
	}

	return values, nil
}
