package db

import (
	"context"
	"testing"
	"time"

	"github.com/frogfromlake/user_service/util"
	"github.com/stretchr/testify/require"
)

func TestAddAccountTypeToAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t, util.RandomPasswordHash())

	tests := []struct {
		name           string
		AccountsID     int64
		AccountTypesID int64
		expectedErr    bool
	}{
		{
			name:           "Normal operation",
			AccountTypesID: accountType.ID,
			AccountsID:     account.ID,
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
			name:           "account is already in the accountType's discography",
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
	account := createRandomAccount(t, util.RandomPasswordHash())
	err := testQueries.AddAccountTypeToAccount(context.Background(), AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	})
	require.NoError(t, err)

	accountTypes, err := testQueries.GetAccountTypesForAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountTypes)
	require.Equal(t, accountType.ID, accountTypes[0].ID)
	require.Equal(t, accountType.Description, accountTypes[0].Description)
	require.WithinDuration(t, accountType.CreatedAt.Time, accountTypes[0].CreatedAt.Time, time.Second)
	require.WithinDuration(t, accountType.UpdatedAt.Time, accountTypes[0].UpdatedAt.Time, time.Second)
}

func TestGetAccountTypeIDsForAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t, util.RandomPasswordHash())
	err := testQueries.AddAccountTypeToAccount(context.Background(), AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	})
	require.NoError(t, err)

	accountTypes, err := testQueries.GetAccountTypeIDsForAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accountTypes)
	require.Equal(t, accountType.ID, accountTypes[0].ID)
	require.WithinDuration(t, accountType.CreatedAt.Time, accountTypes[0].CreatedAt.Time, time.Second)
	require.WithinDuration(t, accountType.UpdatedAt.Time, accountTypes[0].UpdatedAt.Time, time.Second)
}

func TestGetAccountsForAccountType(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t, util.RandomPasswordHash())
	err := testQueries.AddAccountTypeToAccount(context.Background(), AddAccountTypeToAccountParams{
		AccountsID:     account.ID,
		AccountTypesID: accountType.ID,
	})
	require.NoError(t, err)

	accounts, err := testQueries.GetAccountsForAccountType(context.Background(), accountType.ID)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	require.Equal(t, account.ID, accounts[0].ID)
	require.Equal(t, account.Username, accounts[0].Username)
	require.Equal(t, account.Email, accounts[0].Email)
	require.Equal(t, account.CountryCode, accounts[0].CountryCode)
	require.WithinDuration(t, account.CreatedAt.Time, accounts[0].CreatedAt.Time, time.Second)
	require.WithinDuration(t, account.UpdatedAt.Time, accounts[0].UpdatedAt.Time, time.Second)
}

func TestRemoveAccountTypeFromAccount(t *testing.T) {
	accountType := createRandomAccountType(t)
	account := createRandomAccount(t, util.RandomPasswordHash())
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
	account := createRandomAccount(t, util.RandomPasswordHash())
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
	account := createRandomAccount(t, util.RandomPasswordHash())
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
