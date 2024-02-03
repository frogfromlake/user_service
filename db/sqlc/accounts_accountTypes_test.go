package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAddAccountTypeToAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)

	tests := []struct {
		name           string
		AccountsID     int64
		AccountTypesID int32
		expectedErr    bool
	}{
		{
			name:           "Normal operation",
			AccountsID:     account.ID,
			AccountTypesID: accountType.ID,
			expectedErr:    false,
		},
		{
			name:           "account doesn't exist",
			AccountsID:     account.ID,
			AccountTypesID: -1,
			expectedErr:    true,
		},
		{
			name:           "accountType doesn't exist",
			AccountsID:     -1,
			AccountTypesID: accountType.ID,
			expectedErr:    true,
		},
		{
			name:           "uniqueViolation",
			AccountsID:     account.ID,
			AccountTypesID: accountType.ID,
			expectedErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := AddAccountTypeToAccountParams{
				AccountsID:     tt.AccountsID,
				AccountTypesID: tt.AccountTypesID,
			}
			err := testQueries.AddAccountTypeToAccount(context.Background(), arg)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.AccountsID, arg.AccountsID)
				require.Equal(t, tt.AccountTypesID, arg.AccountTypesID)
			}
		})
	}
}

func TestGetAccountTypesForAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)
	err := testQueries.AddAccountTypeToAccount(context.Background(), AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	})
	require.NoError(t, err)

	accountTypes, err := testQueries.GetAccountTypesForAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountTypes)
	for i, accountType := range accountTypes {
		require.NotEmpty(t, accountType)
		require.Equal(t, accountType.ID, accountTypes[i].ID)
		require.WithinDuration(t, accountType.CreatedAt, accountTypes[i].CreatedAt, time.Second)
		require.WithinDuration(t, accountType.UpdatedAt, accountTypes[i].UpdatedAt, time.Second)
	}
}

func TestGetAccountTypeIDsForAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)
	err := testQueries.AddAccountTypeToAccount(context.Background(), AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	})
	require.NoError(t, err)

	accountTypes, err := testQueries.GetAccountTypesForAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountTypes)
	for i, accountType := range accountTypes {
		require.NotEmpty(t, accountType)
		require.Equal(t, accountType.ID, accountTypes[i].ID)
		require.WithinDuration(t, accountType.CreatedAt, accountTypes[i].CreatedAt, time.Second)
		require.WithinDuration(t, accountType.UpdatedAt, accountTypes[i].UpdatedAt, time.Second)
	}
}

func TestGetAccountsForAccountType(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)
	err := testQueries.AddAccountTypeToAccount(context.Background(), AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	})
	require.NoError(t, err)

	accounts, err := testQueries.GetAccountsForAccountType(context.Background(), accountType.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	for i, account := range accounts {
		require.NotEmpty(t, account)
		require.Equal(t, account.ID, accounts[i].ID)
		require.Equal(t, account.Owner, accounts[i].Owner)
		require.Equal(t, account.AvatarUri, accounts[i].AvatarUri)
		require.Zero(t, accounts[i].Plays)
		require.Zero(t, accounts[i].Likes)
		require.Zero(t, accounts[i].Follows)
		require.Zero(t, accounts[i].Shares)
		require.WithinDuration(t, account.CreatedAt, accounts[i].CreatedAt, time.Second)
		require.WithinDuration(t, account.UpdatedAt, accounts[i].UpdatedAt, time.Second)
	}
}

func TestRemoveAccountTypeFromAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)
	arg := AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	}
	err := testQueries.AddAccountTypeToAccount(context.Background(), arg)
	require.NoError(t, err)

	argRemove := RemoveAccountTypeFromAccountParams(arg)

	err = testQueries.RemoveAccountTypeFromAccount(context.Background(), argRemove)
	require.NoError(t, err)

	accountTypes, err := testQueries.GetAccountTypesForAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Empty(t, accountTypes)
}

func TestRemoveAllRelationshipsForAccountAccountType(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)
	arg := AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	}
	err := testQueries.AddAccountTypeToAccount(context.Background(), arg)
	require.NoError(t, err)

	err = testQueries.RemoveAllRelationshipsForAccountAccountType(context.Background(), account.ID)
	require.NoError(t, err)

	accountTypes, err := testQueries.GetAccountsForAccountType(context.Background(), accountType.ID)
	require.NoError(t, err)
	require.Empty(t, accountTypes)
}

func TestRemoveAllRelationshipsForAccountTypeAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t)
	arg := AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	}
	err := testQueries.AddAccountTypeToAccount(context.Background(), arg)
	require.NoError(t, err)

	err = testQueries.RemoveAllRelationshipsForAccountTypeAccount(context.Background(), accountType.ID)
	require.NoError(t, err)

	accounts, err := testQueries.GetAccountTypesForAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.Empty(t, accounts)
}
