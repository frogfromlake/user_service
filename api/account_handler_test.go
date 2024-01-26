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
		"account_type_id": params.AccountTypeID,
		"username":        params.AccountParams.Username,
		"email":           params.AccountParams.Email,
		"password_hash":   params.AccountParams.PasswordHash,
		"country_code":    params.AccountParams.CountryCode,
		"avatar_url":      params.AccountParams.AvatarUrl,
	}
}

func createRandomAccountParamsAndReturns() (db.CreateAccountTxParams, db.CreateAccountTxResult) {
	createAccTxParams := db.CreateAccountTxParams{
		AccountParams: db.CreateAccountParams{
			Username:     util.RandomUsername(),
			Email:        util.RandomEmail(),
			PasswordHash: util.RandomPasswordHash(),
			CountryCode:  util.RandomCountryCode(),
			AvatarUrl:    util.ConvertToText("http://example.com/avatar.png"),
		},
		AccountTypeID: []int64{1, 2},
	}

	createAccTxReturn := db.CreateAccountTxResult{
		Account: &db.CreateAccountRow{
			ID:          util.RandomInt(1, 1000),
			Username:    createAccTxParams.AccountParams.Username,
			Email:       createAccTxParams.AccountParams.Email,
			CountryCode: createAccTxParams.AccountParams.CountryCode,
			CreatedAt:   util.ConvertToTimestamptz(util.RandomDate()),
			UpdatedAt:   util.ConvertToTimestamptz(util.RandomDate()),
		},
		AccountTypeID: []int64{1, 2},
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
					GetAccountTypeIDsForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return([]db.GetAccountTypeIDsForAccountRow{{ID: 1}}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				expected := getAccountByIDResponse{
					Account:        account,
					AccountTypeIDs: []db.GetAccountTypeIDsForAccountRow{{ID: 1}},
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
					Return(db.GetAccountByIDRow{}, sql.ErrNoRows)
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
					Return(db.GetAccountByIDRow{}, sql.ErrConnDone)
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
					GetAccountTypeIDsForAccount(gomock.Any(), gomock.Any()).
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

func TestGetAccountByUsernameAPI(t *testing.T) {
	params := randomAccountFromID()
	account := db.GetAccountByUsernameRow(params)
	getAccByUsername := db.GetAccountByUsernameRow(account)

	testCases := []struct {
		name            string
		accountUsername string
		buildStubs      func(store *mock_db.MockStore)
		checkResponse   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:            "OK",
			accountUsername: account.Username,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypeIDsForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return([]db.GetAccountTypeIDsForAccountRow{{ID: 1}}, nil)

				store.EXPECT().
					GetAccountByUsername(gomock.Any(), gomock.Eq(account.Username)).
					Times(1).
					Return(getAccByUsername, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				expected := getAccountByUsernameResponse{
					Account:        account,
					AccountTypeIDs: []db.GetAccountTypeIDsForAccountRow{{ID: 1}},
				}
				requireBodyMatch(t, recorder.Body, expected, "getAccountByUsernameResponse")
			},
		},
		{
			name:            "NotFound",
			accountUsername: account.Username,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByUsername(gomock.Any(), gomock.Eq(account.Username)).
					Times(1).
					Return(db.GetAccountByUsernameRow{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:            "InternalError",
			accountUsername: account.Username,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByUsername(gomock.Any(), gomock.Eq(account.Username)).
					Times(1).
					Return(db.GetAccountByUsernameRow{}, sql.ErrConnDone)
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
			name:            "ShouldBindUriError",
			accountUsername: "%20", // Percent-encoded space character
			buildStubs:      func(store *mock_db.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:            "InvalidAccountType",
			accountUsername: account.Username,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByUsername(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypeIDsForAccount(gomock.Any(), gomock.Any()).
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
				url = "/accounts/name"
			} else {
				url = fmt.Sprintf("/accounts/name/%s", tc.accountUsername)
			}
			request := httptest.NewRequest("GET", url, nil)

			// send request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccountByAllParamsAPI(t *testing.T) {
	account := randomAccount()
	getAccByAllParams := db.GetAccountByAllParamsParams{
		Username:    account.Username,
		Email:       account.Email,
		CountryCode: account.CountryCode,
		AvatarUrl:   account.AvatarUrl,
	}

	testCases := []struct {
		name          string
		Params        db.GetAccountByAllParamsParams
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			Params: getAccByAllParams,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByAllParams(gomock.Any(), gomock.Eq(getAccByAllParams)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypeIDsForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return([]db.GetAccountTypeIDsForAccountRow{{ID: 1}}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, account, "db.StreamfairAccount")
			},
		},
		{
			name:   "NotFound",
			Params: getAccByAllParams,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByAllParams(gomock.Any(), gomock.Eq(getAccByAllParams)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			Params: getAccByAllParams,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByAllParams(gomock.Any(), gomock.Eq(getAccByAllParams)).
					Times(1).
					Return(db.UserSvcAccount{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:   "BadRedquest",
			Params: db.GetAccountByAllParamsParams{},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByAllParams(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InvalidAccountType",
			Params: getAccByAllParams,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByAllParams(gomock.Any(), gomock.Any()).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypeIDsForAccount(gomock.Any(), gomock.Any()).
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
			url := fmt.Sprintf("/accounts/params?username=%s&email=%s&country_code=%s&avatar_url=%s",
				tc.Params.Username, tc.Params.Email, tc.Params.CountryCode, tc.Params.AvatarUrl.String)

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
		"id":            params.ID,
		"username":      params.Username,
		"email":         params.Email,
		"country_code":  params.CountryCode,
		"avatar_url":    params.AvatarUrl,
		"likes_count":   params.LikesCount,
		"follows_count": params.FollowsCount,
	}
}

func TestUpdateAccountAPI(t *testing.T) {
	account := randomAccountFromUpdate()
	updateAccParams := db.UpdateAccountParams{
		ID:           account.ID,
		Username:     util.RandomUsername(),
		Email:        util.RandomEmail(),
		CountryCode:  util.RandomCountryCode(),
		AvatarUrl:    util.ConvertToText("http://example.com/avatar.png"),
		LikesCount:   1,
		FollowsCount: 1,
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
				requireBodyMatch(t, recorder.Body, account, "db.UpdateAccountRow")
			},
		},
		{
			name: "InternalError",
			body: UpdateAccountParamsToBody(updateAccParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Eq(updateAccParams)).
					Times(1).
					Return(db.UpdateAccountRow{}, sql.ErrConnDone)
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
				"id":            util.RandomInt(1, 1000),
				"username":      "ab", // Invalid username length
				"email":         "invalid email",
				"country_code":  "USA",
				"avatar_url":    "http://example.com/avatar.png",
				"likes_count":   1,
				"follows_count": 1,
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

func UpdateAccountPasswordParamsToBody(params db.UpdateAccountPasswordParams) gin.H {
	return gin.H{
		"id":            params.ID,
		"password_hash": params.PasswordHash,
	}
}

func TestUpdateAccountPasswordAPI(t *testing.T) {
	account := randomAccountFromPasswordUpdate()
	updateAccPwParams := db.UpdateAccountPasswordParams{
		ID:           account.ID,
		PasswordHash: util.RandomPasswordHash(),
	}

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: UpdateAccountPasswordParamsToBody(updateAccPwParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccountPassword(gomock.Any(), gomock.Eq(updateAccPwParams)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, account, "db.UpdateAccountPasswordRow")
			},
		},
		{
			name: "InternalError",
			body: UpdateAccountPasswordParamsToBody(updateAccPwParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccountPassword(gomock.Any(), gomock.Eq(updateAccPwParams)).
					Times(1).
					Return(db.UpdateAccountPasswordRow{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRedquest",
			body: UpdateAccountPasswordParamsToBody(db.UpdateAccountPasswordParams{}),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccountPassword(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRedquestJSON",
			body: gin.H{
				"id":            util.RandomInt(1, 1000),
				"password_hash": "ab", // Invalid password hash length
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

			url := fmt.Sprintf("/accounts/password/%d", tc.body["id"])
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
					Return(db.GetAccountByIDRow{}, sql.ErrNoRows)

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
		ID:           util.RandomInt(1, 1000),
		Username:     util.RandomUsername(),
		Email:        util.RandomEmail(),
		CountryCode:  util.RandomCountryCode(),
		AvatarUrl:    util.ConvertToText("http://example.com/avatar.png"),
		LikesCount:   0,
		FollowsCount: 0,
		CreatedAt:    util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt:    util.ConvertToTimestamptz(util.RandomDate()),
	}
}

func randomAccountFromList(n int) []db.ListAccountsRow {
	accounts := make([]db.ListAccountsRow, n)
	account := db.ListAccountsRow{
		ID:          util.RandomInt(1, 1000),
		Username:    util.RandomUsername(),
		CountryCode: util.RandomCountryCode(),
		CreatedAt:   util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt:   util.ConvertToTimestamptz(util.RandomDate()),
	}
	for i := 0; i < n; i++ {
		accounts[i] = account
	}
	return accounts
}

func randomAccountFromID() db.GetAccountByIDRow {
	return db.GetAccountByIDRow{
		ID:           util.RandomInt(1, 1000),
		Username:     util.RandomUsername(),
		Email:        util.RandomEmail(),
		CountryCode:  util.RandomCountryCode(),
		AvatarUrl:    util.ConvertToText("http://example.com/avatar.png"),
		LikesCount:   0,
		FollowsCount: 0,
		CreatedAt:    util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt:    util.ConvertToTimestamptz(util.RandomDate()),
	}
}

func randomAccountFromUpdate() db.UpdateAccountRow {
	return db.UpdateAccountRow{
		ID:          util.RandomInt(1, 1000),
		Username:    util.RandomUsername(),
		CountryCode: util.RandomCountryCode(),
		CreatedAt:   util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt:   util.ConvertToTimestamptz(util.RandomDate()),
	}
}

func randomAccountFromPasswordUpdate() db.UpdateAccountPasswordRow {
	return db.UpdateAccountPasswordRow{
		ID:        util.RandomInt(1, 1000),
		Username:  util.RandomUsername(),
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
	case "getAccountByUsernameResponse":
		if !reflect.DeepEqual(expected, *gotResult.(*getAccountByUsernameResponse)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*getAccountByUsernameResponse))
		}
	case "db.StreamfairAccount":
		if !reflect.DeepEqual(expected, *gotResult.(*db.UserSvcAccount)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*db.UserSvcAccount))
		}
	case "[]db.ListAccountsRow":
		if !reflect.DeepEqual(expected, *gotResult.(*[]db.ListAccountsRow)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*[]db.ListAccountsRow))
		}
	case "db.UpdateAccountRow":
		if !reflect.DeepEqual(expected, *gotResult.(*db.UpdateAccountRow)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*db.UpdateAccountRow))
		}
	case "db.UpdateAccountPasswordRow":
		if !reflect.DeepEqual(expected, *gotResult.(*db.UpdateAccountPasswordRow)) {
			t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, *gotResult.(*db.UpdateAccountPasswordRow))
		}
	default:
		t.Errorf("Body mismatch for %s: \nEXP: %+v, \nGOT: %+v", typeName, expected, gotResult)
	}
}
