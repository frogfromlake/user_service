package db

// import (
// 	"context"
// 	"math/rand"
// 	"testing"
// 	"time"

// 	"github.com/Streamfair/streamfair_user_svc/util"
// 	"github.com/stretchr/testify/require"
// )

// func TestCreateAccountTx(t *testing.T) {
// 	store := NewStore(testDB)

// 	n := 5

// 	errs := make(chan error)
// 	results := make(chan CreateAccountTxResult)

// 	for i := 0; i < n; i++ {
// 		go func(i int) {
// 			source := rand.NewSource(time.Now().UnixNano())
// 			r := rand.New(source)

// 			// Create a user
// 			user, err := store.CreateUser(context.Background(), CreateUserParams{
// 				Username:     util.RandomString(12, r),
// 				FullName:     "Test Full Name",
// 				Email:        util.RandomEmail(r),
// 				PasswordHash: "password",
// 				CountryCode:  "US",
// 			})
// 			require.NoError(t, err)

// 			// Create an account type
// 			accountTypeParam := CreateAccountTypeParams{
// 				Type:        util.RandomString(10),
// 				Permissions: []byte(`{"permissions": "permissions"}`),
// 				IsArtist:    true,
// 				IsProducer:  false,
// 				IsWriter:    false,
// 				IsLabel:     false,
// 			}
// 			randomAccountType, err := store.CreateAccountType(context.Background(), accountTypeParam)
// 			require.NoError(t, err)

// 			// Create an account
// 			accountParams := CreateAccountParams{
// 				Owner:       user.Username,
// 				AccountType: int32(randomAccountType.ID),
// 				AvatarUri:   util.ConvertToText("http://example.com/test-artist-avatar.png"),
// 			}
// 			params := CreateAccountTxParams{
// 				AccountParams: accountParams,
// 			}
// 			result, err := store.CreateAccountTx(context.Background(), params)
// 			require.NoError(t, err)
// 			require.NotEmpty(t, result)

// 			// Send the error and result to the channel
// 			errs <- err
// 			results <- result
// 		}(i)
// 	}

// 	// Check that the account types were associated with the account
// 	for i := 0; i < n; i++ {
// 		err := <-errs
// 		require.NoError(t, err)

// 		result := <-results
// 		require.NotEmpty(t, result)

// 		account := result.Account
// 		require.NotEmpty(t, account)
// 		require.NotZero(t, account.ID)
// 		require.NotZero(t, account.CreatedAt)
// 		require.NotZero(t, account.UpdatedAt)

// 		// account should be found in the database
// 		_, err = store.GetAccountByID(context.Background(), account.ID)
// 		require.NoError(t, err)

// 		// Verify the relationship from the account side
// 		accountTypes, err := store.GetAccountTypesForAccount(context.Background(), account.ID)
// 		require.NoError(t, err)
// 		require.Equal(t, 1, len(accountTypes))
// 		require.Equal(t, result.Account.AccountType, accountTypes[0].ID)

// 		// Verify the relationship from the account type side
// 		accounts, err := store.GetAccountsForAccountType(context.Background(), accountTypes[0].ID)
// 		require.NoError(t, err)
// 		require.Equal(t, 1, len(accounts))
// 		require.Equal(t, account.ID, accounts[0].ID)
// 	}
// }

// func TestDeleteAccountTx(t *testing.T) {
// 	store := NewStore(testDB)

// 	n := 5

// 	errs := make(chan error)

// 	for i := 0; i < n; i++ {
// 		go func(i int) {
// 			source := rand.NewSource(time.Now().UnixNano())
// 			r := rand.New(source)

// 			// Create a user
// 			user, err := store.CreateUser(context.Background(), CreateUserParams{
// 				Username:     util.RandomString(12, r),
// 				FullName:     "Test Full Name",
// 				Email:        util.RandomEmail(r),
// 				PasswordHash: "password",
// 				CountryCode:  "US",
// 			})
// 			require.NoError(t, err)

// 			// Create an account type
// 			accountTypeParam := CreateAccountTypeParams{
// 				Type:        util.RandomString(10),
// 				Permissions: []byte(`{"permissions": "permissions"}`),
// 				IsArtist:    true,
// 				IsProducer:  false,
// 				IsWriter:    false,
// 				IsLabel:     false,
// 			}
// 			randomAccountType, err := store.CreateAccountType(context.Background(), accountTypeParam)
// 			require.NoError(t, err)

// 			// Create an account
// 			accountParams := CreateAccountParams{
// 				Owner:       user.Username,
// 				AccountType: int32(randomAccountType.ID),
// 				AvatarUri:   util.ConvertToText("http://example.com/test-artist-avatar.png"),
// 			}
// 			params := CreateAccountTxParams{
// 				AccountParams: accountParams,
// 			}
// 			result, err := store.CreateAccountTx(context.Background(), params)
// 			require.NoError(t, err)
// 			require.NotEmpty(t, result)

// 			account := result.Account
// 			require.NotEmpty(t, account)
// 			require.NotZero(t, account.ID)

// 			// Delete the account
// 			err = store.DeleteAccountTx(context.Background(), account.ID)
// 			require.NoError(t, err)

// 			// Verify the account has been disassociated
// 			accountTypes, err := store.GetAccountTypesForAccount(context.Background(), account.ID)
// 			require.NoError(t, err)
// 			require.Equal(t, 0, len(accountTypes))

// 			// Send the error to the channel
// 			errs <- err
// 		}(i)
// 	}

// 	// Check that the accounts were deleted
// 	for i := 0; i < n; i++ {
// 		err := <-errs
// 		require.NoError(t, err)
// 	}
// }
