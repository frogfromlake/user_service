package db

import (
	"context"
	"testing"
	"time"

	"github.com/Streamfair/user_service/util"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T, passwordHash string) CreateAccountRow {
	arg := CreateAccountParams{
		Username:     util.RandomUsername(),
		Email:        util.RandomEmail(),
		PasswordHash: passwordHash,
		CountryCode:  util.RandomCountryCode(),
		AvatarUrl:    util.ConvertToText("http://example.com/test-account-avatar.png"),
	}

	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, arg.Username, account.Username)
	require.Equal(t, arg.Email, account.Email)
	require.Equal(t, arg.CountryCode, account.CountryCode)
	require.NotEmpty(t, arg.AvatarUrl)
	require.NotZero(t, account.ID)
	require.WithinDuration(t, time.Now(), account.CreatedAt.Time, time.Second)
	require.WithinDuration(t, time.Now(), account.UpdatedAt.Time, time.Second)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t, util.RandomPasswordHash())
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t, util.RandomPasswordHash())

	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	deletedAccount, err := testQueries.GetAccountByID(context.Background(), account.ID)
	require.Error(t, err)
	require.Empty(t, deletedAccount)
}

func TestGetAccountByID(t *testing.T) {
	passwordHash := util.RandomPasswordHash()

	account := createRandomAccount(t, passwordHash)
	fetchedAccount, err := testQueries.GetAccountByID(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccount)

	require.Equal(t, account.ID, fetchedAccount.ID)
	require.Equal(t, account.Username, fetchedAccount.Username)
	require.Equal(t, account.Email, fetchedAccount.Email)
	require.Equal(t, account.CountryCode, fetchedAccount.CountryCode)
	require.NotEmpty(t, fetchedAccount.AvatarUrl)
	require.WithinDuration(t, account.CreatedAt.Time, fetchedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, account.UpdatedAt.Time, fetchedAccount.UpdatedAt.Time, time.Second)
}

func TestGetAccountByUsername(t *testing.T) {
	passwordHash := util.RandomPasswordHash()

	account := createRandomAccount(t, passwordHash)
	fetchedAccount, err := testQueries.GetAccountByUsername(context.Background(), account.Username)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccount)

	require.Equal(t, account.ID, fetchedAccount.ID)
	require.Equal(t, account.Username, fetchedAccount.Username)
	require.Equal(t, account.Email, fetchedAccount.Email)
	require.Equal(t, account.CountryCode, fetchedAccount.CountryCode)
	require.NotEmpty(t, fetchedAccount.AvatarUrl)
	require.WithinDuration(t, account.CreatedAt.Time, fetchedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, account.UpdatedAt.Time, fetchedAccount.UpdatedAt.Time, time.Second)
}

func TestGetAccountByAllParams(t *testing.T) {
	username := util.RandomUsername()
	email := util.RandomEmail()
	passwordHash := util.RandomPasswordHash()
	countryCode := util.RandomCountryCode()
	avatarUrl := util.ConvertToText("http://example.com/test-account-avatar.png")
	arg := CreateAccountParams{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		CountryCode:  countryCode,
		AvatarUrl:    avatarUrl,
	}
	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	argGet := GetAccountByAllParamsParams{
		Username:    username,
		Email:       email,
		CountryCode: countryCode,
		AvatarUrl:   avatarUrl,
	}
	fetchedAccount, err := testQueries.GetAccountByAllParams(context.Background(), argGet)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccount)
	require.Equal(t, account.ID, fetchedAccount.ID)
	require.Equal(t, account.Username, fetchedAccount.Username)
	require.Equal(t, account.Email, fetchedAccount.Email)
	require.Equal(t, account.CountryCode, fetchedAccount.CountryCode)
	require.NotEmpty(t, fetchedAccount.AvatarUrl)
	require.WithinDuration(t, account.CreatedAt.Time, fetchedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, account.UpdatedAt.Time, fetchedAccount.UpdatedAt.Time, time.Second)
}

func TestListAccounts(t *testing.T) {
	var ErrNegativeOffset = &pgconn.PgError{
		Code:    "2201X",
		Message: "OFFSET must not be negative",
	}

	for i := 0; i < 10; i++ {
		createRandomAccount(t, util.RandomPasswordHash())
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
				require.NotEqual(t, int64(0), account.ID)
				require.NotEmpty(t, account.Username)
				require.NotZero(t, account.CountryCode)
			}
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t, util.RandomPasswordHash())
	arg := UpdateAccountParams{
		ID:          account.ID,
		Username:    util.RandomUsername(),
		Email:       util.RandomEmail(),
		CountryCode: util.RandomCountryCode(),
		AvatarUrl:   util.ConvertToText("http://example.com/test-account-avatar.png"),
	}

	updatedAccount, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)

	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, arg.Username, updatedAccount.Username)
	require.Equal(t, arg.CountryCode, updatedAccount.CountryCode)
	require.WithinDuration(t, account.CreatedAt.Time, updatedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, time.Now(), updatedAccount.UpdatedAt.Time, time.Second)
}

func TestUpdateAccountPassword(t *testing.T) {
	passwordHash := util.RandomPasswordHash()
	newPasswordHash := util.RandomPasswordHash()
	require.NotEqual(t, passwordHash, newPasswordHash)

	account := createRandomAccount(t, passwordHash)
	arg := UpdateAccountPasswordParams{
		ID:           account.ID,
		PasswordHash: newPasswordHash,
	}

	updatedAccount, err := testQueries.UpdateAccountPassword(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccount)
	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, account.Username, updatedAccount.Username)
	require.WithinDuration(t, account.CreatedAt.Time, updatedAccount.CreatedAt.Time, time.Second)
	require.WithinDuration(t, time.Now(), updatedAccount.UpdatedAt.Time, time.Second)
}
