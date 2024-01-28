package db

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) CreateUserRow {
	byteHash, err := util.HashPassword(util.RandomPassword())
	hashedPassword := base64.StdEncoding.EncodeToString(byteHash.Hash)
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:     util.RandomUsername(),
		FullName:     util.RandomString(12),
		Email:        util.RandomEmail(),
		PasswordHash: hashedPassword,
		CountryCode:  util.RandomCountryCode(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	require.NotZero(t, user.ID)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.CountryCode, user.CountryCode)
	require.WithinDuration(t, time.Now(), user.CreatedAt.Time, time.Second)

	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUserByUsername(t *testing.T) {
	user := createRandomUser(t)

	fetchedUser, err := testQueries.GetUserByUsername(context.Background(), user.Username)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedUser)
	require.Equal(t, user.Username, fetchedUser.Username)
	require.Equal(t, user.FullName, fetchedUser.FullName)
	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.CountryCode, fetchedUser.CountryCode)
	require.True(t, fetchedUser.UsernameChangedAt.Time.IsZero())
	require.True(t, fetchedUser.EmailChangedAt.Time.IsZero())
	require.True(t, fetchedUser.PasswordChangedAt.Time.IsZero())
	require.WithinDuration(t, user.CreatedAt.Time, fetchedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, time.Now(), fetchedUser.UpdatedAt.Time, time.Second)
}
