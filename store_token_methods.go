package vaultstore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dracory/sb"
	"github.com/dromara/carbon/v2"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// ErrTokenExpired is returned when a token has expired
var ErrTokenExpired = errors.New("token has expired")

// TokenCreateOptions contains optional parameters for token creation
type TokenCreateOptions struct {
	// ExpiresAt is the expiration time for the token
	// If zero value, token never expires
	ExpiresAt time.Time
}

// TokenCreate creates a new record and returns the token
func (store *storeImplementation) TokenCreate(ctx context.Context, data string, password string, tokenLength int, options ...TokenCreateOptions) (token string, err error) {
	maxAttempts := 3

	for attempt := 0; attempt < maxAttempts; attempt++ {
		token, err = generateToken(tokenLength)
		if err != nil {
			return "", err
		}

		// Check if token already exists
		existing, err := store.RecordFindByToken(ctx, token)
		if err != nil {
			return "", err
		}
		if existing != nil {
			continue // Try again with a new token
		}

		encodedData, err := encode(data, password)
		if err != nil {
			return "", fmt.Errorf("failed to encode data: %w", err)
		}

		var newEntry = NewRecord().
			SetToken(token).
			SetValue(encodedData).
			SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
			SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

		// Apply options if provided
		if len(options) > 0 && !options[0].ExpiresAt.IsZero() {
			newEntry.SetExpiresAt(carbon.CreateFromStdTime(options[0].ExpiresAt).ToDateTimeString(carbon.UTC))
		}

		err = store.RecordCreate(ctx, newEntry)
		if err != nil {
			continue // Try again
		}

		// Link record to password identity only if the feature is enabled
		if store.passwordIdentityEnabled {
			passwordID, err := store.findOrCreateIdentity(ctx, password)
			if err != nil {
				return "", fmt.Errorf("failed to find or create identity: %w", err)
			}

			err = store.linkRecordToIdentity(ctx, newEntry.GetID(), passwordID)
			if err != nil {
				return "", fmt.Errorf("failed to link record to identity: %w", err)
			}
		}

		return token, nil
	}

	return "", errors.New("failed to create token")
}

func (store *storeImplementation) TokenCreateCustom(ctx context.Context, token string, data string, password string, options ...TokenCreateOptions) (err error) {
	// Validate token is not empty (custom tokens can have any format)
	if token == "" {
		return errors.New("token is empty")
	}

	// Check if token already exists
	existing, err := store.RecordFindByToken(ctx, token)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("token already exists")
	}

	encodedData, err := encode(data, password)
	if err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	var newEntry = NewRecord().
		SetToken(token).
		SetValue(encodedData).
		SetCreatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC)).
		SetUpdatedAt(carbon.Now(carbon.UTC).ToDateTimeString(carbon.UTC))

	// Apply options if provided
	if len(options) > 0 && !options[0].ExpiresAt.IsZero() {
		newEntry.SetExpiresAt(carbon.CreateFromStdTime(options[0].ExpiresAt).ToDateTimeString(carbon.UTC))
	}

	err = store.RecordCreate(ctx, newEntry)
	if err != nil {
		return err
	}

	// Link record to password identity only if the feature is enabled
	if store.passwordIdentityEnabled {
		passwordID, err := store.findOrCreateIdentity(ctx, password)
		if err != nil {
			return fmt.Errorf("failed to find or create identity: %w", err)
		}

		err = store.linkRecordToIdentity(ctx, newEntry.GetID(), passwordID)
		if err != nil {
			return fmt.Errorf("failed to link record to identity: %w", err)
		}
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
func (store *storeImplementation) TokenDelete(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	return store.RecordDeleteByToken(ctx, token)
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
func (store *storeImplementation) TokenRead(ctx context.Context, token string, password string) (value string, err error) {
	if token == "" {
		return "", errors.New("token is empty")
	}

	entry, err := store.RecordFindByToken(ctx, token)

	if err != nil {
		return "", err
	}

	if entry == nil {
		return "", errors.New("token does not exist")
	}

	// Check if token has expired
	expiresAt := entry.GetExpiresAt()
	if expiresAt != "" && expiresAt != sb.MAX_DATETIME {
		expiryTime := carbon.Parse(expiresAt, carbon.UTC)
		if !expiryTime.IsZero() && carbon.Now(carbon.UTC).Gt(expiryTime) {
			return "", ErrTokenExpired
		}
	}

	decoded, err := decode(entry.GetValue(), password)

	if err != nil {
		return "", err
	}

	// On-access migration: Check if record is linked to a password identity
	// Only if password identity feature is enabled
	// If not, link it now (this handles records created before identity-based management)
	if store.passwordIdentityEnabled {
		existingPassID, _ := store.getRecordPasswordID(ctx, entry.GetID())
		if existingPassID == "" {
			// Record not linked yet, create the link within a transaction
			err = store.gormDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				// Create identity within transaction
				passwordID, identityErr := store.findOrCreateIdentity(ctx, password)
				if identityErr != nil {
					return fmt.Errorf("unable to create identity for record %s: %w", entry.GetID(), identityErr)
				}

				// Link record to identity within same transaction
				linkErr := store.linkRecordToIdentity(ctx, entry.GetID(), passwordID)
				if linkErr != nil {
					return fmt.Errorf("unable to link record %s to identity %s: %w", entry.GetID(), passwordID, linkErr)
				}

				return nil
			})

			if err != nil {
				// Transaction failed - this is a data consistency issue
				// Return error to signal the problem to the caller
				return "", fmt.Errorf("migration transaction failed for record %s: %w", entry.GetID(), err)
			}
		}
	}

	return decoded, nil
}

// TokenRenew extends the expiration time of an existing token
func (store *storeImplementation) TokenRenew(ctx context.Context, token string, expiresAt time.Time) error {
	if token == "" {
		return errors.New("token is empty")
	}

	entry, err := store.RecordFindByToken(ctx, token)

	if err != nil {
		return err
	}

	if entry == nil {
		return errors.New("token does not exist")
	}

	if expiresAt.IsZero() {
		entry.SetExpiresAt(sb.MAX_DATETIME)
	} else {
		entry.SetExpiresAt(carbon.CreateFromStdTime(expiresAt).ToDateTimeString(carbon.UTC))
	}

	return store.RecordUpdate(ctx, entry)
}

// TokensExpiredSoftDelete soft-deletes all expired tokens
func (store *storeImplementation) TokensExpiredSoftDelete(ctx context.Context) (count int64, err error) {
	records, err := store.RecordList(ctx, RecordQuery())
	if err != nil {
		return 0, err
	}

	for _, record := range records {
		expiresAt := record.GetExpiresAt()
		if expiresAt == "" || expiresAt == sb.MAX_DATETIME {
			continue
		}

		expiryTime := carbon.Parse(expiresAt, carbon.UTC)
		if expiryTime.IsZero() || carbon.Now(carbon.UTC).Lte(expiryTime) {
			continue
		}

		err = store.RecordSoftDelete(ctx, record)
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
}

// TokensExpiredDelete permanently deletes all expired tokens
func (store *storeImplementation) TokensExpiredDelete(ctx context.Context) (count int64, err error) {
	records, err := store.RecordList(ctx, RecordQuery())
	if err != nil {
		return 0, err
	}

	for _, record := range records {
		expiresAt := record.GetExpiresAt()
		if expiresAt == "" || expiresAt == sb.MAX_DATETIME {
			continue
		}

		expiryTime := carbon.Parse(expiresAt, carbon.UTC)
		if expiryTime.IsZero() || carbon.Now(carbon.UTC).Lte(expiryTime) {
			continue
		}

		err = store.RecordDeleteByID(ctx, record.GetID())
		if err != nil {
			return count, err
		}
		count++
	}

	return count, nil
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
func (store *storeImplementation) TokenSoftDelete(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is empty")
	}

	return store.RecordSoftDeleteByToken(ctx, token)
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
func (store *storeImplementation) TokenUpdate(ctx context.Context, token string, value string, password string) (err error) {
	if token == "" {
		return errors.New("token is empty")
	}

	entry, errFind := store.RecordFindByToken(ctx, token)

	if errFind != nil {
		return err
	}

	if entry == nil {
		return errors.New("token does not exist")
	}

	encodedValue, err := encode(value, password)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}

	entry.SetValue(encodedValue)

	err = store.RecordUpdate(ctx, entry)
	if err != nil {
		return err
	}

	// Link record to password identity only if the feature is enabled
	if store.passwordIdentityEnabled {
		passwordID, err := store.findOrCreateIdentity(ctx, password)
		if err != nil {
			return fmt.Errorf("failed to find or create identity: %w", err)
		}

		err = store.linkRecordToIdentity(ctx, entry.GetID(), passwordID)
		if err != nil {
			return fmt.Errorf("failed to link record to identity: %w", err)
		}
	}

	return nil
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
func (store *storeImplementation) TokensRead(ctx context.Context, tokens []string, password string) (values map[string]string, err error) {
	values = map[string]string{}

	// Validate all tokens are not empty
	for _, token := range tokens {
		if token == "" {
			return values, errors.New("token cannot be empty")
		}
	}

	entries, err := store.RecordList(ctx, RecordQuery().SetTokenIn(tokens))

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
		// Check if token has expired
		expiresAt := entry.GetExpiresAt()
		if expiresAt != "" && expiresAt != sb.MAX_DATETIME {
			expiryTime := carbon.Parse(expiresAt, carbon.UTC)
			if !expiryTime.IsZero() && carbon.Now(carbon.UTC).Gt(expiryTime) {
				continue // Skip expired tokens
			}
		}

		decoded, err := decode(entry.GetValue(), password)

		if err != nil {
			return map[string]string{}, errors.New("decryption failed for one or more tokens")
		}

		values[entry.GetToken()] = decoded
	}

	return values, nil
}
