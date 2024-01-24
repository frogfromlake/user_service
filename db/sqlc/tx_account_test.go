package db

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/frogfromlake/streamfair_backend/user_service/util"
	"github.com/stretchr/testify/require"
)

func GenerateRandomParams(r *rand.Rand) (CreateAccountParams, CreateAccountTypeParams) {
	randomAccountParams := CreateAccountParams{
		Username:     util.RandomString(10, r),
		Email:        util.RandomEmail(r),
		PasswordHash: util.RandomString(10, r),
		CountryCode:  util.RandomCountryCode(r),
		AvatarUrl:    util.ConvertToText("http://example.com/test-artist-avatar.png"),
	}
	randomAccountTypeParams := CreateAccountTypeParams{
		Description: util.ConvertToText("Test Account Type"),
		Permissions: []byte(`{"can upload": "false"}`),
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
			accountTypeCreationParams := make([]CreateAccountTypeParams, n)
			accountTypeParams := make([]int64, n)

			for j := 0; j < n; j++ {
				var randomAccountParams, randomAccountTypeParams = GenerateRandomParams(r)
				accountParams[j] = randomAccountParams
				accountTypeCreationParams[j] = randomAccountTypeParams
				accountTypeParams[j] = util.RandomInt(1, 100, r)
			}

			// Create the account type
			accountType, err := store.CreateAccountType(context.Background(), accountTypeCreationParams[0])
			if err != nil {
				log.Fatalf("failed to create account type: %v", err)
			}

			params := CreateAccountTxParams{
				AccountParams: accountParams[i],
				AccountTypeID: []int64{accountType.ID},
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
		require.Equal(t, params.AccountParams.Username, account.Username)
		require.Equal(t, params.AccountParams.Email, account.Email)
		require.Equal(t, params.AccountParams.CountryCode, account.CountryCode)
		require.NotZero(t, account.CreatedAt)
		require.NotZero(t, account.UpdatedAt)

		// account should be found in the database
		_, err = store.GetAccountByID(context.Background(), account.ID)
		require.NoError(t, err)

		// Verify the account types have been associated with the account
		for index, accountTypeID := range result.AccountTypeID {
			fetchedAccountType, err := store.GetAccountType(context.Background(), accountTypeID)
			require.NoError(t, err)
			require.NotEmpty(t, fetchedAccountType)
			require.NotZero(t, fetchedAccountType.ID)

			// Verify the relationship from the account side
			accountTypes, err := store.GetAccountTypesForAccount(context.Background(), account.ID)
			require.NoError(t, err)
			require.Len(t, accountTypes, len(params.AccountTypeID))
			require.Equal(t, accountTypeID, accountTypes[index].ID)

			// Verify the relationship from the account type side
			accounts, err := store.GetAccountsForAccountType(context.Background(), accountTypeID)
			require.NoError(t, err)
			for i, account := range accounts {
				require.Equal(t, account.ID, accounts[i].ID)
			}
		}
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
			accountTypeParams := make([]int64, n)

			for j := 0; j < n; j++ {
				var randomAccountParams, randomAccountTypeParams = GenerateRandomParams(r)
				accountParams[j] = randomAccountParams
				accountTypeCreationParams[j] = randomAccountTypeParams
				accountTypeParams[j] = util.RandomInt(1, 100, r)
			}

			// Create the account type
			accountType, err := store.CreateAccountType(context.Background(), accountTypeCreationParams[0])
			if err != nil {
				log.Fatalf("failed to create account type: %v", err)
			}

			params := CreateAccountTxParams{
				AccountParams: accountParams[i],
				AccountTypeID: []int64{accountType.ID},
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

		// Verify the account types have been disassociated
		for _, accountTypeID := range result.AccountTypeID {
			accountTypes, err := store.GetAccountTypesForAccount(context.Background(), account.ID)
			require.NoError(t, err)
			require.NotContains(t, accountTypes, accountTypeID)
		}
	}
}
