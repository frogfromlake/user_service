package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_db "github.com/frogfromlake/user_service/db/mock"
	db "github.com/frogfromlake/user_service/db/sqlc"
	"github.com/frogfromlake/user_service/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAddSongAPI(t *testing.T) {
	addSongTxParams, addSongTxReturn := addRandomSongParamsAndReturns()
	validAccountType := randomAccountType(true)
	invalidAccountType := randomAccountType(false)
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
		setup         func()
	}{
		{
			name: "OK",
			body: addSongParamsToBody(addSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(addSongTxParams.AccountID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Eq(addSongTxParams)).
					Times(1).
					Return(addSongTxReturn, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, addSongTxReturn, "db.AddSongTxResult")
			},
		},
		{
			name: "InternalError",
			body: addSongParamsToBody(addSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(addSongTxParams.AccountID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Eq(addSongTxParams)).
					Times(1).
					Return(db.AddSongTxResult{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			body: addSongParamsToBody(db.AddSongTxParams{}),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "StatusConflict",
			body: addSongParamsToBody(addSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(addSongTxParams.AccountID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, recorder.Code)
			},
			setup: func() {
				addSongTxRequest := addSongTxRequest{
					AccountID:      addSongTxParams.AccountID,
					SongParams:     addSongTxParams.SongParams,
					GenreParams:    addSongTxParams.GenreParams,
					ArtistParams:   addSongTxParams.ArtistParams,
					AlbumParams:    addSongTxParams.AlbumParams,
					ProducerParams: addSongTxParams.ProducerParams,
					WriterParams:   addSongTxParams.WriterParams,
					LabelParams:    addSongTxParams.LabelParams,
				}
				// Add the hash for the addSongTxParams to the map
				songHashes[computeSongHash(normalizeParams(addSongTxRequest))] = true
			},
		},
		{
			name: "InvalidAccount",
			body: addSongParamsToBody(addSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(addSongTxParams.AccountID)).
					Times(1).
					Return(invalidAccountType, nil)

				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			// Don't clear the songHashes map at the start of each test
			if tc.name != "StatusConflict" && tc.name != "OK" {
				songHashes = make(map[string]bool)
			}

			// Run the setup function if it exists
			if tc.setup != nil {
				tc.setup()
			}

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/songs"
			request := httptest.NewRequest("POST", url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func addRandomSongParamsAndReturns() (db.AddSongTxParams, db.AddSongTxResult) {
	addSongTxParams := db.AddSongTxParams{
		AccountID: 1,
		SongParams: db.CreateSongParams{
			Title:        util.RandomString(8),
			Duration:     util.RandomDuration(),
			ReleaseDate:  util.ConvertToDate(util.RandomDate()),
			SongCoverUrl: util.ConvertToText("http://example.com/test-song-cover.png"),
			AudioFileUrl: "http://example.com/test-song.mp3",
		},
	}
	addSongTxReturn := db.AddSongTxResult{
		AccountID: 1,
		Song: db.StreamfairSong{
			ID:           1,
			SongUuid:     util.ConvertToUUID(util.RandomUUID()),
			Title:        addSongTxParams.SongParams.Title,
			Duration:     addSongTxParams.SongParams.Duration,
			ReleaseDate:  addSongTxParams.SongParams.ReleaseDate,
			SongCoverUrl: addSongTxParams.SongParams.SongCoverUrl,
			AudioFileUrl: addSongTxParams.SongParams.AudioFileUrl,
			PlaysCount:   0,
			LikesCount:   0,
			CreatedAt:    util.ConvertToTimestamp(util.RandomDate()),
			UpdatedAt:    util.ConvertToTimestamptz(util.RandomDate()),
		},
	}
	return addSongTxParams, addSongTxReturn
}

func updateRandomSongParamsAndReturns() (db.UpdateSongTxParams, db.UpdateSongTxResults) {
	updateSongTxParams := db.UpdateSongTxParams{
		AccountID: 1,
		SongParams: db.UpdateSongParams{
			ID:           1,
			Title:        util.RandomString(8),
			Duration:     util.RandomDuration(),
			ReleaseDate:  util.ConvertToDate(util.RandomDate()),
			SongCoverUrl: util.ConvertToText("http://example.com/test-song-cover.png"),
			AudioFileUrl: "http://example.com/test-song.mp3",
			PlaysCount:   1,
			LikesCount:   1,
		},
	}
	updateSongTxReturn := db.UpdateSongTxResults{
		AccountID: 1,
		AddSongTxResult: db.AddSongTxResult{
			Song: db.StreamfairSong{
				ID:           1,
				SongUuid:     util.ConvertToUUID(util.RandomUUID()),
				Title:        updateSongTxParams.SongParams.Title,
				Duration:     updateSongTxParams.SongParams.Duration,
				ReleaseDate:  updateSongTxParams.SongParams.ReleaseDate,
				SongCoverUrl: updateSongTxParams.SongParams.SongCoverUrl,
				AudioFileUrl: updateSongTxParams.SongParams.AudioFileUrl,
				PlaysCount:   0,
				LikesCount:   0,
				CreatedAt:    util.ConvertToTimestamp(util.RandomDate()),
				UpdatedAt:    util.ConvertToTimestamptz(util.RandomDate()),
			},
		},
	}

	return updateSongTxParams, updateSongTxReturn
}

func addSongParamsToBody(params db.AddSongTxParams) gin.H {
	return gin.H{
		"accountID":  params.AccountID,
		"songParams": params.SongParams,
	}
}

func updateSongParamsToBody(params db.UpdateSongTxParams) gin.H {
	return gin.H{
		"accountID": params.AccountID,
		"songParams": gin.H{
			"id":             params.SongParams.ID,
			"title":          params.SongParams.Title,
			"duration":       params.SongParams.Duration,
			"release_date":   params.SongParams.ReleaseDate,
			"song_cover_url": params.SongParams.SongCoverUrl,
			"audio_file_url": params.SongParams.AudioFileUrl,
			"plays_count":    params.SongParams.PlaysCount,
			"likes_count":    params.SongParams.LikesCount,
		},
	}
}

func randomAccountType(validity bool) []db.StreamfairAccountType {

	if validity {
		return []db.StreamfairAccountType{
			{
				ID:          2,
				Description: util.ConvertToText(util.RandomString(10)),
				Permissions: []byte{1, 2},
				IsArtist:    true,
				IsProducer:  false,
				IsWriter:    false,
				IsLabel:     false,
			},
		}
	} else {
		return []db.StreamfairAccountType{
			{
				ID:          2,
				Description: util.ConvertToText(util.RandomString(10)),
				Permissions: []byte{1, 2},
				IsArtist:    false,
				IsProducer:  false,
				IsWriter:    false,
				IsLabel:     false,
			},
		}
	}
}

func TestUpdateSongAPI(t *testing.T) {
	updateSongTxParams, updateSongTxReturn := updateRandomSongParamsAndReturns()
	validAccountType := randomAccountType(true)
	invalidAccountType := randomAccountType(false)
	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: updateSongParamsToBody(updateSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(updateSongTxParams.AccountID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					UpdateSongTx(gomock.Any(), gomock.Eq(updateSongTxParams)).
					Times(1).
					Return(updateSongTxReturn, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatch(t, recorder.Body, updateSongTxReturn, "db.UpdateSongTxResults")
			},
		},
		{
			name: "InternalError",
			body: updateSongParamsToBody(updateSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(updateSongTxParams.AccountID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					UpdateSongTx(gomock.Any(), gomock.Eq(updateSongTxParams)).
					Times(1).
					Return(db.UpdateSongTxResults{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			body: updateSongParamsToBody(db.UpdateSongTxParams{}),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidAccount",
			body: updateSongParamsToBody(updateSongTxParams),
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(updateSongTxParams.AccountID)).
					Times(1).
					Return(invalidAccountType, nil)

				store.EXPECT().
					AddSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
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

			url := fmt.Sprintf("/songs/%d", updateSongTxParams.SongParams.ID)
			request := httptest.NewRequest("PUT", url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteSongAPI(t *testing.T) {
	account := randomAccountFromPasswordUpdate()
	validAccountType := randomAccountType(true)
	invalidAccountType := randomAccountType(false)
	deleteSongTxParams := db.DeleteSongTxParams{
		SongID: util.RandomInt(1, 1000),
	}
	testCases := []struct {
		name          string
		songID        int64
		body          gin.H
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:   "OK",
			songID: deleteSongTxParams.SongID,
			body:   gin.H{"accountID": account.ID},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					DeleteSongTx(gomock.Any(), gomock.Eq(deleteSongTxParams)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:   "BadRedquestURI",
			songID: 0,
			body:   gin.H{"accountID": db.DeleteSongTxParams{}},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					DeleteSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "BadRedquestJSON",
			songID: deleteSongTxParams.SongID,
			body:   gin.H{"accountID": db.DeleteSongTxParams{}},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					DeleteSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:   "InvalidAccount",
			songID: deleteSongTxParams.SongID,
			body:   gin.H{"accountID": account.ID},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(invalidAccountType, nil)

				store.EXPECT().
					DeleteSongTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name:   "InternalError",
			songID: deleteSongTxParams.SongID,
			body:   gin.H{"accountID": account.ID},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(validAccountType, nil)

				store.EXPECT().
					DeleteSongTx(gomock.Any(), gomock.Eq(deleteSongTxParams)).
					Times(1).
					Return(sql.ErrConnDone)
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

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/songs/%d", tc.songID)
			request := httptest.NewRequest("DELETE", url, bytes.NewReader(data))

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestValidAccount(t *testing.T) {
	account := randomAccountFromPasswordUpdate()
	testCases := []struct {
		name          string
		accountID     int64
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "InternalError",
			accountID: account.ID,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetAccountTypesForAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				var resp map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				require.Contains(t, resp, "error")
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
			
			ctx, _ := gin.CreateTestContext(recorder)
			server.validAccount(ctx, tc.accountID)

			tc.checkResponse(t, recorder)
		})
	}
}
