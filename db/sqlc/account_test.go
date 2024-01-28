package db

import (
	"context"
	"testing"
	"time"

	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) CreateAccountRow {
	user := createRandomUser(t)
	accountType := createRandomAccountType(t)
	arg := CreateAccountParams{
		Owner:       user.Username,
		AccountType: accountType.ID,
		AvatarUri:   util.ConvertToText("http://example.com/test-account-avatar.png"),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.NotZero(t, account.ID)
	require.Equal(t, arg.Owner, account.Owner)
	require.Equal(t, arg.AvatarUri, account.AvatarUri)
	require.WithinDuration(t, time.Now(), account.CreatedAt.Time, time.Second)
	require.WithinDuration(t, time.Now(), account.UpdatedAt.Time, time.Second)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	deletedAccount, err := testQueries.GetAccountByID(context.Background(), account.ID)
	require.Error(t, err)
	require.Empty(t, deletedAccount)
}

func TestGetAccountByID(t *testing.T) {
	account := createRandomAccount(t)
	fetchedAccount, err := testQueries.GetAccountByID(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccount)

	require.Equal(t, account.ID, fetchedAccount.ID)
	require.Equal(t, account.Owner, fetchedAccount.Owner)
	require.NotEmpty(t, fetchedAccount.AvatarUri)
	require.WithinDuration(t, account.CreatedAt.Time, fetchedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, account.UpdatedAt.Time, fetchedAccount.UpdatedAt.Time, time.Second)
}

func TestGetAccountByOwner(t *testing.T) {
	account := createRandomAccount(t)
	fetchedAccount, err := testQueries.GetAccountByOwner(context.Background(), account.Owner)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccount)

	require.Equal(t, account.ID, fetchedAccount.ID)
	require.Equal(t, account.Owner, fetchedAccount.Owner)
	require.Equal(t, account.AvatarUri, fetchedAccount.AvatarUri)
	require.WithinDuration(t, account.CreatedAt.Time, fetchedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, account.UpdatedAt.Time, fetchedAccount.UpdatedAt.Time, time.Second)
}

func TestListAccounts(t *testing.T) {
	var ErrNegativeOffset = &pgconn.PgError{
		Code:    "2201X",
		Message: "OFFSET must not be negative",
	}

	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	testCases := []struct {
		Name        string
		Params      ListAccountsParams
		ExpectedLen int
		ExpectedErr error
	}{
		{
			Name: "ValidLimitAndOffset",
			Params: ListAccountsParams{
				Limit:  5,
				Offset: 5,
			},
			ExpectedLen: 5,
			ExpectedErr: nil,
		},
		{
			Name: "InvalidLimit",
			Params: ListAccountsParams{
				Limit:  0,
				Offset: 5,
			},
			ExpectedLen: 0,
			ExpectedErr: nil,
		},
		{
			Name: "InvalidOffset",
			Params: ListAccountsParams{
				Limit:  5,
				Offset: -1,
			},
			ExpectedLen: 0,
			ExpectedErr: ErrNegativeOffset,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accounts, err := testQueries.ListAccounts(context.Background(), tc.Params)

			if tc.ExpectedErr != nil {
				require.Error(t, err)
				switch e := err.(type) {
				case *pgconn.PgError:
					require.Equal(t, ErrNegativeOffset.Code, e.Code)
					require.Equal(t, ErrNegativeOffset.Message, e.Message)
				case error:
					require.Equal(t, tc.ExpectedErr, e)
				default:
					t.Errorf("unexpected error type: %T", e)
				}
				return
			}

			require.NoError(t, err)
			require.Len(t, accounts, tc.ExpectedLen)

			for _, account := range accounts {
				require.NotEmpty(t, account)
				require.NotZero(t, account.ID)
				require.NotEmpty(t, account.Owner)
				require.NotZero(t, account.CreatedAt)
				require.NotZero(t, account.UpdatedAt)
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)
	arg := UpdateAccountParams{
		ID:          account.ID,
		Owner:       account.Owner,
		AccountType: account.AccountType,
		AvatarUri:   util.ConvertToText("http://example.com/test-account-avatar.png"),
		Plays:       100,
		Likes:       100,
		Follows:     100,
		Shares:      100,
	}

	updatedAccount, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)

	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, arg.Owner, updatedAccount.Owner)
	require.Equal(t, arg.AvatarUri, updatedAccount.AvatarUri)
	require.Equal(t, arg.Plays, updatedAccount.Plays)
	require.Equal(t, arg.Likes, updatedAccount.Likes)
	require.Equal(t, arg.Follows, updatedAccount.Follows)
	require.Equal(t, arg.Shares, updatedAccount.Shares)
	require.WithinDuration(t, account.CreatedAt.Time, updatedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, time.Now(), updatedAccount.UpdatedAt.Time, time.Second)
}
