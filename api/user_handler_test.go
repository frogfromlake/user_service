package api

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":     user.Username,
				"full_name":    user.FullName,
				"email":        user.Email,
				"password":     password,
				"country_code": user.CountryCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				arg := db.CreateUserParams{
					Username:    user.Username,
					FullName:    user.FullName,
					Email:       user.Email,
					CountryCode: user.CountryCode,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":     user.Username,
				"full_name":    user.FullName,
				"email":        user.Email,
				"password":     password,
				"country_code": user.CountryCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserSvcUser{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "DuplicateUsername",
			body: gin.H{
				"username":     user.Username,
				"full_name":    user.FullName,
				"email":        user.Email,
				"password":     password,
				"country_code": user.CountryCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.UserSvcUser{}, &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"})
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username":     "",
				"full_name":    user.FullName,
				"email":        user.Email,
				"password":     password,
				"country_code": user.CountryCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username":     user.Username,
				"full_name":    user.FullName,
				"email":        "invalid_email",
				"password":     password,
				"country_code": user.CountryCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "TooShortPassword",
			body: gin.H{
				"username":     user.Username,
				"full_name":    user.FullName,
				"email":        user.Email,
				"password":     "123",
				"country_code": user.CountryCode,
			},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.UserSvcUser, password string) {
	password = util.RandomString(8)
	byteHash, err := util.HashPassword(password)
	hashedPassword := base64.StdEncoding.EncodeToString(byteHash.Hash)

	require.NoError(t, err)

	user = db.UserSvcUser{
		Username:     util.RandomString(8),
		FullName:     util.RandomString(8),
		Email:        util.RandomEmail(),
		PasswordHash: hashedPassword,
		CountryCode:  util.RandomCountryCode(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.UserSvcUser) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.UserSvcUser
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)

	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.PasswordHash)
	require.Equal(t, user.CountryCode, gotUser.CountryCode)
}
