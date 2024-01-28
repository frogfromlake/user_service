package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateHash(t *testing.T) {
	password := RandomString(8)
	hashedPassword, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)

	// Test with matching password
	err = ComparePassword(hashedPassword.Hash, hashedPassword.Salt, password)
	require.NoError(t, err)

	// Test with wrong password
	wrongPassword := RandomString(8)
	err = ComparePassword(hashedPassword.Hash, hashedPassword.Salt, wrongPassword)
	require.Error(t, err)
	require.Equal(t, "invalid password", err.Error())

	// Test with empty password
	emptyPassword := ""
	err = ComparePassword(hashedPassword.Hash, hashedPassword.Salt, emptyPassword)
	require.Error(t, err)

	// Test with nil hash and salt
	err = ComparePassword(nil, nil, password)
	require.Error(t, err)

	// Test with same password but different hash
	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	require.NotEqual(t, hashedPassword.Hash, hashedPassword2.Hash)
}
