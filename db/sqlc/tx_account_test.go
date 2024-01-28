package db

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/stretchr/testify/require"
)

func GenerateRandomParams(t *testing.T, r *rand.Rand) (CreateAccountParams, CreateAccountTypeParams) {
	user := createRandomUser(t)
	accountType := createRandomAccountType(t)

	randomAccountParams := CreateAccountParams{
		Owner:       user.Username,
		AccountType: accountType.ID,
		AvatarUri:   util.ConvertToText("http://example.com/test-artist-avatar.png"),
	}
	randomAccountTypeParams := CreateAccountTypeParams{
		Type:        util.RandomString(10, r),
		Permissions: []byte(`{"key": "value"}`),
		IsArtist:    false,
		IsProducer:  false,
		IsWriter:    false,
		IsLabel:     false,
	}
	return randomAccountParams, randomAccountTypeParams
}

func TestCreateAccountTx(t *testing.T) {
	store := NewStore(testDB)

	n := 5

	errs := make(chan error)
	results := make(chan CreateAccountTxResult)
	paramsChan := make(chan CreateAccountTxParams)

	for i := 0; i < n; i++ {
		go func(i int) {
			source := rand.NewSource(time.Now().UnixNano())
			r := rand.New(source)

			accountParams := make([]CreateAccountParams, n)

			for j := 0; j < n; j++ {
				var randomAccountParams, _ = GenerateRandomParams(t, r)
				accountParams[j] = randomAccountParams
			}

			params := CreateAccountTxParams{
				AccountParams: accountParams[i],
			}

			// Send the params to the channel
			paramsChan <- params

			result, err := store.CreateAccountTx(context.Background(), params)

			// Send the error and result to the channel
			errs <- err
			results <- result
		}(i)
	}

	// Check that the account types were associated with the account
	for i := 0; i < n; i++ {
		params := <-paramsChan
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		account := result.Account
		require.NotEmpty(t, account)
		require.NotZero(t, account.ID)
		require.Equal(t, params.AccountParams.Owner, account.Owner)
		require.Equal(t, params.AccountParams.AvatarUri, account.AvatarUri)
		require.NotZero(t, account.CreatedAt)
		require.NotZero(t, account.UpdatedAt)

		// account should be found in the database
		_, err = store.GetAccountByID(context.Background(), account.ID)
		require.NoError(t, err)

		// Verify the relationship from the account side
		accountTypes, err := store.GetAccountTypesForAccount(context.Background(), account.ID)
		require.NoError(t, err)
		require.Equal(t, 1, len(accountTypes))
		require.Equal(t, result.Account.AccountType, accountTypes[0].ID)

		// Verify the relationship from the account type side
		accounts, err := store.GetAccountsForAccountType(context.Background(), accountTypes[0].ID)
		require.NoError(t, err)
		require.Equal(t, 1, len(accounts))
		require.Equal(t, account.ID, accounts[0].ID)
	}
}

func TestDeleteAccountTx(t *testing.T) {
	store := NewStore(testDB)

	n := 5

	errs := make(chan error)
	results := make(chan CreateAccountTxResult)

	for i := 0; i < n; i++ {
		go func(i int) {
			source := rand.NewSource(time.Now().UnixNano())
			r := rand.New(source)

			accountParams := make([]CreateAccountParams, n)
			accountTypeCreationParams := make([]CreateAccountTypeParams, n)

			for j := 0; j < n; j++ {
				var randomAccountParams, createAccountTypeParams = GenerateRandomParams(t, r)
				accountParams[j] = randomAccountParams
				accountTypeCreationParams[j] = createAccountTypeParams
			}

			params := CreateAccountTxParams{
				AccountParams: accountParams[i],
			}

			result, err := store.CreateAccountTx(context.Background(), params)
			require.NoError(t, err)
			require.NotEmpty(t, result)

			account := result.Account
			require.NotEmpty(t, account)
			require.NotZero(t, account.ID)

			// Delete the account
			err = store.DeleteAccountTx(context.Background(), account.ID)
			require.NoError(t, err)

			// Send the error and result to the channel
			errs <- err
			results <- result
		}(i)
	}

	// Check that the account types were disassociated from the account
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		account := result.Account
		require.NotEmpty(t, account)
		require.NotZero(t, account.ID)

		// Verify the account has been deleted
		_, err = store.GetAccountByID(context.Background(), account.ID)
		require.Error(t, err)

		// Verify the accounttype has been disassociated
		accountTypes, err := store.GetAccountTypesForAccount(context.Background(), account.ID)
		require.NoError(t, err)
		require.NotContains(t, accountTypes, result.Account.ID)
	}
}
