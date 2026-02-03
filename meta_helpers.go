package vaultstore

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// hashPassword creates a bcrypt hash of the password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BCRYPT_COST)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// verifyPassword verifies a password against a bcrypt hash
func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generatePasswordID generates a unique password identity ID
func generatePasswordID() string {
	return PASSWORD_ID_PREFIX + uuid.New().String()
}

// generateRecordMetaID generates a record ID for the meta table
func generateRecordMetaID(recordID string) string {
	return RECORD_META_ID_PREFIX + recordID
}
