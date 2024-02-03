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
	"time"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func createAccountParamsToBody(params db.CreateAccountTxParams) gin.H {
	return gin.H{
		"owner":        params.AccountParams.Owner,
		"account_type": params.AccountParams.AccountType,
		"avatar_uri":   params.AccountParams.AvatarUri,
	}
}

func createRandomAccountParamsAndReturns(owner string) (db.CreateAccountTxParams, db.CreateAccountTxResult) {
	createAccTxParams := db.CreateAccountTxParams{
		AccountParams: db.CreateAccountParams{
			Owner:       owner,
			AccountType: 1,
			AvatarUri:   util.ConvertToText("http://example.com/avatar.png"),
		},
	}

	createAccTxReturn := db.CreateAccountTxResult{
		Account: &db.CreateAccountRow{
			ID:        util.RandomInt(1, 1000),
			Owner:     createAccTxParams.AccountParams.Owner,
			AvatarUri: createAccTxParams.AccountParams.AvatarUri,
			CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
			UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		},
	}
	return createAccTxParams, createAccTxReturn
}

func TestCreateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	createAccTxParams, createAccTxReturn := createRandomAccountParamsAndReturns(user.Username)

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: createAccountParamsToBody(createAccTxParams),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
		// {
		// 	name: "NoAuthorization",
		// 	body: createAccountParamsToBody(createAccTxParams),
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 	},
		// 	buildStubs: func(store *mock_db.MockStore) {
		// 		store.EXPECT().
		// 			CreateAccountTx(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InternalError",
		// 	body: createAccountParamsToBody(createAccTxParams),
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 		addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
		// 	},
		// 	buildStubs: func(store *mock_db.MockStore) {
		// 		store.EXPECT().
		// 			CreateAccountTx(gomock.Any(), gomock.Eq(createAccTxParams)).
		// 			Times(1).
		// 			Return(db.CreateAccountTxResult{}, sql.ErrConnDone)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "BadRedquest",
		// 	body: createAccountParamsToBody(db.CreateAccountTxParams{}),
		// 	setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// 		addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
		// 	},
		// 	buildStubs: func(store *mock_db.MockStore) {
		// 		store.EXPECT().
		// 			CreateAccountTx(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock_db.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/accounts"
			request := httptest.NewRequest("POST", url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.localTokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccountByIdAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccountFromID(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorizedUser", 1, time.Minute)
			},
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
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// build request
			url := fmt.Sprintf("/accounts/id/%d", tc.accountID)
			request := httptest.NewRequest("GET", url, nil)

			tc.setupAuth(t, request, server.localTokenMaker)

			// send request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetAccountByOwnerAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccountFromID(user.Username)

	testCases := []struct {
		name          string
		accountOwner  string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:         "OK",
			accountOwner: account.Owner,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			name:         "UnauthorizedUser",
			accountOwner: account.Owner,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorizedUser", 1, time.Minute)
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByOwner(gomock.Any(), gomock.Eq(account.Owner)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return([]db.UserSvcAccountType{{ID: 1}}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:         "NoAuthorization",
			accountOwner: account.Owner,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByOwner(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:         "NotFound",
			accountOwner: account.Owner,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			name: "BadRedquest",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
			buildStubs: func(store *mock_db.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:         "ShouldBindUriError",
			accountOwner: "%20", // Percent-encoded space character
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
			buildStubs: func(store *mock_db.MockStore) {},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:         "InvalidAccountType",
			accountOwner: account.Owner,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// build request
			var url string
			if tc.name == "BadRedquest" {
				url = "/accounts/owner"
			} else {
				url = fmt.Sprintf("/accounts/owner/%s", tc.accountOwner)
			}
			request := httptest.NewRequest("GET", url, nil)

			tc.setupAuth(t, request, server.localTokenMaker)

			// send request
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	user, _ := randomUser(t)

	n := 5
	accounts := randomAccountFromList(n, user.Username)

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.ListAccountsParams{
					Owner:  user.Username,
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
			name: "NoAuthorization",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := "/accounts/list"
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// Add query parameters to request URL
			q := request.URL.Query()
			q.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			q.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			request.URL.RawQuery = q.Encode()

			tc.setupAuth(t, request, server.localTokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func UpdateAccountParamsToBody(params db.UpdateAccountParams) gin.H {
	return gin.H{
		"id":         params.ID,
		"username":   params.Owner,
		"avatar_url": params.AvatarUri,
		"plays":      params.Plays,
		"likes":      params.Likes,
		"follows":    params.Follows,
		"shares":     params.Shares,
	}
}

func TestUpdateAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	updateAccParams := db.UpdateAccountParams{
		ID:        account.ID,
		Owner:     account.Owner,
		AvatarUri: util.ConvertToText("http://example.com/avatar.png"),
		Plays:     1,
		Likes:     1,
		Follows:   1,
		Shares:    1,
	}

	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: UpdateAccountParamsToBody(updateAccParams),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
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
		{
			name: "NoAuthorization",
			body: UpdateAccountParamsToBody(updateAccParams),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					UpdateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/accounts/update/%d", tc.body["id"])
			request := httptest.NewRequest("PUT", url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.localTokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccountFromID(user.Username)

	testCases := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			name:      "UnauthorizedUser",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorizedUser", 1, time.Minute)
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)

				store.EXPECT().
					DeleteAccountTx(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NoAuthorization",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountByID(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)

				store.EXPECT().
					DeleteAccountTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, 1, time.Minute)
			},
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/delete/%d", tc.accountID)
			request := httptest.NewRequest("DELETE", url, nil)

			tc.setupAuth(t, request, server.localTokenMaker)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(owner string) db.UserSvcAccount {
	return db.UserSvcAccount{
		ID:        util.RandomInt(1, 1000),
		Owner:     owner,
		AvatarUri: util.ConvertToText("http://example.com/avatar.png"),
		Plays:     0,
		Likes:     0,
		Follows:   0,
		CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
	}
}

func randomAccountFromList(n int, username string) []db.ListAccountsRow {
	accounts := make([]db.ListAccountsRow, n)
	account := db.ListAccountsRow{
		ID:        util.RandomInt(1, 1000),
		Owner:     username,
		CreatedAt: util.ConvertToTimestamptz(util.RandomDate()),
		UpdatedAt: util.ConvertToTimestamptz(util.RandomDate()),
	}
	for i := 0; i < n; i++ {
		accounts[i] = account
	}
	return accounts
}

func randomAccountFromID(owner string) db.UserSvcAccount {
	return db.UserSvcAccount{
		ID:        util.RandomInt(1, 1000),
		Owner:     owner,
		AvatarUri: util.ConvertToText("http://example.com/avatar.png"),
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
