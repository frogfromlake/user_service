package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func createAccountParamsToBody(params db.CreateAccountTxParams) gin.H {
	return gin.H{
		"account_type_ids": params.AccountTypeIDs,
		"owner":            params.AccountParams.Owner,
		"avatar_uri":       params.AccountParams.AvatarUrl,
	}
}

func createRandomAccountParamsAndReturns() (db.CreateAccountTxParams, db.CreateAccountTxResult) {
	createAccTxParams := db.CreateAccountTxParams{
		AccountTypeIDs: []int64{1, 2},
		AccountParams: db.CreateAccountParams{
			Owner:     util.RandomUsername(),
			AvatarUrl: util.ConvertToText("http://example.com/avatar.png"),
		},
	}

	createAccTxReturn := db.CreateAccountTxResult{
		Account: &db.CreateAccountRow{
			ID:        util.RandomInt(1, 1000),
			Owner:     createAccTxParams.AccountParams.Owner,
			AvatarUrl: createAccTxParams.AccountParams.AvatarUrl,
			CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
			UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		},
		AccountTypeIDs: []int64{1, 2},
	}
	return createAccTxParams, createAccTxReturn
}

func TestCreateAccountAPI(t *testing.T) {
	createAccTxParams, createAccTxReturn := createRandomAccountParamsAndReturns()

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: createAccountParamsToBody(createAccTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateAccountTx(gomock.Any(), gomock.Eq(createAccTxParams)).
					Times(1).
					Return(createAccTxReturn, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, createAccTxReturn, "db.CreateAccountTxResult")
			},
		},
		{
			name: "InternalError",
			body: createAccountParamsToBody(createAccTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateAccountTx(gomock.Any(), gomock.Eq(createAccTxParams)).
					Times(1).
					Return(db.CreateAccountTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRedquest",
			body: createAccountParamsToBody(db.CreateAccountTxParams{}),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateAccountTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/accounts"
			request := httptest.NewRequest("POST", url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccountByIdAPI(t *testing.T) {
	account := randomAccountFromID()
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return([]db.UserSvcAccountType{{ID: 1}}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				expected := getAccountByIDResponse{
					Account:      account,
					AccountTypes: []db.UserSvcAccountType{{ID: 1}},
				}
				requireBodyMatch(t, recorder.Body, expected, "getAccountByIDResponse")
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRedquest",
			accountID: 0,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InvalidAccountType",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)
			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// build request
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request := httptest.NewRequest("GET", url, nil)

			// send request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccountByOwnerAPI(t *testing.T) {
	account := randomAccountFromID()

	testCases := []struct {
		name          string
		accountOwner  string
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "OK",
			accountOwner: account.Owner,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return([]db.UserSvcAccountType{{ID: 1}}, nil)

				store.EXPECT().
					GetAccountByOwner(gomock.Any(), gomock.Eq(account.Owner)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				expected := getAccountByOwnerResponse{
					Account:      account,
					AccountTypes: []db.UserSvcAccountType{{ID: 1}},
				}
				requireBodyMatch(t, recorder.Body, expected, "getAccountByOwnerResponse")
			},
		},
		{
			name:         "NotFound",
			accountOwner: account.Owner,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByOwner(gomock.Any(), gomock.Eq(account.Owner)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:         "InternalError",
			accountOwner: account.Owner,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByOwner(gomock.Any(), gomock.Eq(account.Owner)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "BadRedquest",
			buildStubs: func(store *mock_db.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:         "ShouldBindUriError",
			accountOwner: "%20", // Percent-encoded space character
			buildStubs:   func(store *mock_db.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:         "InvalidAccountType",
			accountOwner: account.Owner,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByOwner(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)
			// start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// build request
			var url string
			if tc.name == "BadRedquest" {
				url = "/accounts/owner"
			} else {
				url = fmt.Sprintf("/accounts/owner/%s", tc.accountOwner)
			}
			request := httptest.NewRequest("GET", url, nil)

			// send request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	n := 5
	accounts := randomAccountFromList(n)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.ListAccountsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, accounts, "[]db.ListAccountsRow")
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.ListAccountsRow{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPageID",
			query: Query{
				pageID:   -1,
				pageSize: n,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidPageSize",
			query: Query{
				pageID:   1,
				pageSize: 100000,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func UpdateAccountParamsToBody(params db.UpdateAccountParams) gin.H {
	return gin.H{
		"id":         params.ID,
		"username":   params.Owner,
		"avatar_url": params.AvatarUrl,
		"plays":      params.Plays,
		"likes":      params.Likes,
		"follows":    params.Follows,
		"shares":     params.Shares,
	}
}

func TestUpdateAccountAPI(t *testing.T) {
	account := randomAccount()
	updateAccParams := db.UpdateAccountParams{
		ID:        account.ID,
		Owner:     util.RandomUsername(),
		AvatarUrl: util.ConvertToText("http://example.com/avatar.png"),
		Plays:     1,
		Likes:     1,
		Follows:   1,
		Shares:    1,
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: UpdateAccountParamsToBody(updateAccParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(updateAccParams)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, account, "*db.UserSvcAccount")
			},
		},
		{
			name: "InternalError",
			body: UpdateAccountParamsToBody(updateAccParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(updateAccParams)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRedquestURI",
			body: UpdateAccountParamsToBody(db.UpdateAccountParams{}),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRedquestJSON",
			body: gin.H{
				"id":           util.RandomInt(1, 1000),
				"username":     "ab", // Invalid username length
				"email":        "invalid email",
				"country_code": "USA",
				"avatar_url":   "http://example.com/avatar.png",
				"likes":        1,
				"follows":      1,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/accounts/%d", tc.body["id"])
			request := httptest.NewRequest("PUT", url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteAccountAPI(t *testing.T) {
	account := randomAccountFromID()

	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					DeleteAccountTx(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrNoRows)

				store.EXPECT().
					DeleteAccountTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					DeleteAccountTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "BadRedquest",
			accountID: 0,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Any()).
					Times(0)

				store.EXPECT().
					DeleteAccountTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			request := httptest.NewRequest("DELETE", url, nil)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.UserSvcAccount {
	return db.UserSvcAccount{
		ID:        util.RandomInt(1, 1000),
		Owner:     util.RandomUsername(),
		AvatarUrl: util.ConvertToText("http://example.com/avatar.png"),
		Plays:     0,
		Likes:     0,
		Follows:   0,
		CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
	}
}

func randomAccountFromList(n int) []db.ListAccountsRow {
	accounts := make([]db.ListAccountsRow, n)
	account := db.ListAccountsRow{
		ID:        util.RandomInt(1, 1000),
		Owner:     util.RandomUsername(),
		CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
	}
	for i := 0; i < n; i++ {
		accounts[i] = account
	}
	return accounts
}

func randomAccountFromID() db.UserSvcAccount {
	return db.UserSvcAccount{
		ID:        util.RandomInt(1, 1000),
		Owner:     util.RandomUsername(),
		AvatarUrl: util.ConvertToText("http://example.com/avatar.png"),
		Plays:     0,
		Likes:     0,
		Follows:   0,
		Shares:    0,
		CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
	}
}

func requireBodyMatch(t *testing.T, body *bytes.Buffer, expected interface{}, typeName string) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	// Create a new variable of the expected type and assign it the value of gotResult
	var gotResult = reflect.New(reflect.TypeOf(expected)).Interface()
	err = json.Unmarshal(data, &gotResult)
	if err != nil {
		t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, gotResult)
		return
	}

	switch typeName {
	case "db.CreateAccountTxResult":
		if !reflect.DeepEqual(expected, *gotResult.(*db.CreateAccountTxResult)) {
			t.Errorf("Body mismatch for %s: \nexpected: %+v, \ngot: %+v", typeName, expected, *gotResult.(*db.CreateAccountTxResult))
		}
	case "getAccountByIDResponse":
		if !reflect.DeepEqual(expected, *gotResult.(*getAccountByIDResponse)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*getAccountByIDResponse))
		}
	case "getAccountByOwnerResponse":
		if !reflect.DeepEqual(expected, *gotResult.(*getAccountByOwnerResponse)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*getAccountByOwnerResponse))
		}
	case "[]db.ListAccountsRow":
		if !reflect.DeepEqual(expected, *gotResult.(*[]db.ListAccountsRow)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*[]db.ListAccountsRow))
		}
	case "*db.UserSvcAccount":
		if !reflect.DeepEqual(expected, *gotResult.(*db.UserSvcAccount)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*db.UserSvcAccount))
		}
	default:
		t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, gotResult)
	}
}
