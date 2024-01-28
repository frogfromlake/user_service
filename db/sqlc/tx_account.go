package db

import (
	"context"
)

type CreateAccountTxParams struct {
	AccountParams CreateAccountParams
}

type CreateAccountTxResult struct {
	Account *CreateAccountRow
}

// CreateAccountTx creates a new account and associates it with an account types.
// Returns the created account.
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
		err = q.AddAccountTypeToAccount(ctx, AddAccountTypeToAccountParams{
			AccountsID:     account.ID,
			AccountTypesID: account.AccountType,
		})
		if err != nil {
			return err
		}
		result.Account.AccountType = params.AccountParams.AccountType

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
		// Remove all associations between the account and account types
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
