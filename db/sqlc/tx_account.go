package db

import (
	"context"
)

type CreateAccountTxParams struct {
	AccountTypeIDs []int64
	AccountParams CreateAccountParams
}

type CreateAccountTxResult struct {
	AccountTypeIDs []int64
	Account       *CreateAccountRow
}

// CreateAccountTx creates a new account and it with one ore more account types.
// Returns the created account.
// TODO: Add other necessary associations: account preferences, account playback history, account subscriptions, etc.
func (store *SQLStore) CreateAccountTx(ctx context.Context, params CreateAccountTxParams) (CreateAccountTxResult, error) {
	var result CreateAccountTxResult

	err := store.ExecTx(ctx, func(q *Queries) error {
		// Create the account
		account, err := q.CreateAccount(ctx, params.AccountParams)
		if err != nil {
			return err
		}
		result.Account = &account

		// Associate the account with the account types
		for _, accountTypeID := range params.AccountTypeIDs {
			err = q.AddAccountTypeToAccount(ctx, AddAccountTypeToAccountParams{
				AccountsID:     account.ID,
				AccountTypesID: accountTypeID,
			})
			if err != nil {
				return err
			}
		}
		result.AccountTypeIDs = params.AccountTypeIDs

		return nil
	})

	if err != nil {
		return CreateAccountTxResult{}, err
	}

	return result, nil
}

// DeleteAccountTx deletes an account and its associations with one ore more account types.
// Returns the created account.
func (store *SQLStore) DeleteAccountTx(ctx context.Context, accountID int64) error {
	err := store.ExecTx(ctx, func(q *Queries) error {
		// Delete all associations between the account and account types
		err := q.RemoveAllRelationshipsForAccountAccountType(ctx, accountID)
		if err != nil {
			return err
		}

		// Delete the account
		err = q.DeleteAccount(ctx, accountID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
