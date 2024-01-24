package db

import (
	"context"
	"testing"

	"github.com/frogfromlake/streamfair_backend/user_service/util"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func createRandomAccountType(t *testing.T) UserServiceAccountType {
	arg := CreateAccountTypeParams{
		Description: util.ConvertToText("This is a test account type"),
		Permissions: []byte(`{"key": "value"}`),
		IsArtist:    false,
		IsProducer:  false,
		IsWriter:    false,
		IsLabel:     false,
	}

	accountType, err := testQueries.CreateAccountType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accountType)
	require.Equal(t, arg.Description, accountType.Description)
	require.Equal(t, arg.Permissions, accountType.Permissions)
	require.Equal(t, arg.IsArtist, accountType.IsArtist)
	require.Equal(t, arg.IsProducer, accountType.IsProducer)
	require.Equal(t, arg.IsWriter, accountType.IsWriter)
	require.Equal(t, arg.IsLabel, accountType.IsLabel)
	require.NotZero(t, accountType.ID)
	require.NotZero(t, accountType.CreatedAt)
	require.NotZero(t, accountType.UpdatedAt)

	return accountType
}

func TestCreateAccountType(t *testing.T) {
	createRandomAccountType(t)
}

func TestGetAccountType(t *testing.T) {
	accountType := createRandomAccountType(t)
	fetchedAccountType, err := testQueries.GetAccountType(context.Background(), accountType.ID)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccountType)
	require.Equal(t, accountType.ID, fetchedAccountType.ID)
	require.Equal(t, accountType.Description, fetchedAccountType.Description)
	require.Equal(t, accountType.Permissions, fetchedAccountType.Permissions)
	require.Equal(t, accountType.IsArtist, fetchedAccountType.IsArtist)
	require.Equal(t, accountType.IsProducer, fetchedAccountType.IsProducer)
	require.Equal(t, accountType.IsWriter, fetchedAccountType.IsWriter)
	require.Equal(t, accountType.IsLabel, fetchedAccountType.IsLabel)
	require.Equal(t, accountType.CreatedAt, fetchedAccountType.CreatedAt)
	require.Equal(t, accountType.UpdatedAt, fetchedAccountType.UpdatedAt)
}

func TestGetAccountTypeByAllParams(t *testing.T) {
	arg := CreateAccountTypeParams{
		Description: util.ConvertToText(util.RandomString(10)),
		Permissions: []byte(`{"key": "value"}`),
	}
	accountType, err := testQueries.CreateAccountType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, accountType)

	argGet := GetAccountTypeByAllParamsParams{
		Description: arg.Description,
		Permissions: arg.Permissions,
		IsArtist:    accountType.IsArtist,
		IsProducer:  accountType.IsProducer,
		IsWriter:    accountType.IsWriter,
		IsLabel:     accountType.IsLabel,
	}

	fetchedAccountType, err := testQueries.GetAccountTypeByAllParams(context.Background(), argGet)
	require.NoError(t, err)
	require.NotEmpty(t, fetchedAccountType)
	require.Equal(t, accountType.ID, fetchedAccountType.ID)
	require.Equal(t, accountType.Description, fetchedAccountType.Description)
	require.Equal(t, accountType.Permissions, fetchedAccountType.Permissions)
	require.Equal(t, accountType.IsArtist, fetchedAccountType.IsArtist)
	require.Equal(t, accountType.IsProducer, fetchedAccountType.IsProducer)
	require.Equal(t, accountType.IsWriter, fetchedAccountType.IsWriter)
	require.Equal(t, accountType.IsLabel, fetchedAccountType.IsLabel)
	require.Equal(t, accountType.CreatedAt, fetchedAccountType.CreatedAt)
	require.Equal(t, accountType.UpdatedAt, fetchedAccountType.UpdatedAt)
}

func TestDeleteAccountType(t *testing.T) {
	accountType := createRandomAccountType(t)
	err := testQueries.DeleteAccountType(context.Background(), accountType.ID)
	require.NoError(t, err)
	fetchedAccountType, err := testQueries.GetAccountType(context.Background(), accountType.ID)
	require.Error(t, err)
	require.Empty(t, fetchedAccountType)
}

func TestListAccountTypes(t *testing.T) {
	var ErrNegativeOffset = &pgconn.PgError{
		Code:    "2201X",
		Message: "OFFSET must not be negative",
	}

	for i := 0; i < 10; i++ {
		createRandomAccountType(t)
	}

	testCases := []struct {
		Name        string
		Params      ListAccountTypesParams
		ExpectedLen int
		ExpectedErr error
	}{
		{
			Name: "ValidLimitAndOffset",
			Params: ListAccountTypesParams{
				Limit:  5,
				Offset: 5,
			},
			ExpectedLen: 5,
			ExpectedErr: nil,
		},
		{
			Name: "InvalidLimit",
			Params: ListAccountTypesParams{
				Limit:  0,
				Offset: 5,
			},
			ExpectedLen: 0,
			ExpectedErr: nil,
		},
		{
			Name: "InvalidOffset",
			Params: ListAccountTypesParams{
				Limit:  5,
				Offset: -1,
			},
			ExpectedLen: 0,
			ExpectedErr: ErrNegativeOffset,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			accountTypes, err := testQueries.ListAccountTypes(context.Background(), tc.Params)

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
			require.Len(t, accountTypes, tc.ExpectedLen)

			for _, accountType := range accountTypes {
				require.NotZero(t, accountType.ID)
				require.NotEmpty(t, accountType.Description)
				require.NotEmpty(t, accountType.Permissions)
				require.False(t, accountType.IsArtist)
				require.False(t, accountType.IsProducer)
				require.False(t, accountType.IsWriter)
				require.False(t, accountType.IsLabel)
				require.NotEmpty(t, accountType.CreatedAt)
				require.NotEmpty(t, accountType.UpdatedAt)
			}
		})
	}
}

func TestUpdateAccountType(t *testing.T) {
	accountType := createRandomAccountType(t)
	arg := UpdateAccountTypeParams{
		ID:          accountType.ID,
		Description: util.ConvertToText("This is an updated account type"),
		Permissions: []byte(`{"key": "value"}`),
		IsArtist:    true,
		IsProducer:  true,
		IsWriter:    true,
		IsLabel:     true,
	}
	updatedAccountType, err := testQueries.UpdateAccountType(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedAccountType)
	require.Equal(t, arg.ID, updatedAccountType.ID)
	require.Equal(t, arg.Description, updatedAccountType.Description)
	require.Equal(t, arg.Permissions, updatedAccountType.Permissions)
	require.Equal(t, arg.IsArtist, updatedAccountType.IsArtist)
	require.Equal(t, arg.IsProducer, updatedAccountType.IsProducer)
	require.Equal(t, arg.IsWriter, updatedAccountType.IsWriter)
	require.Equal(t, arg.IsLabel, updatedAccountType.IsLabel)
	require.Equal(t, accountType.CreatedAt, updatedAccountType.CreatedAt)
	require.NotEqual(t, accountType.UpdatedAt, updatedAccountType.UpdatedAt)
}
