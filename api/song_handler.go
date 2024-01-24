package api

import (
	"net/http"

	db "github.com/frogfromlake/user_service/db/sqlc"
	"github.com/gin-gonic/gin"
)

type addSongTxRequest struct {
	AccountID      int64                      `json:"accountID" binding:"required,min=1"`
	SongParams     db.CreateSongParams        `json:"songParams" binding:"required"`
	GenreParams    []*db.CreateGenreParams    `json:"genreParams" binding:"omitempty"`
	ArtistParams   []*db.CreateArtistParams   `json:"artistParams" binding:"omitempty"`
	AlbumParams    []*db.CreateAlbumParams    `json:"albumParams" binding:"omitempty"`
	ProducerParams []*db.CreateProducerParams `json:"producerParams" binding:"omitempty"`
	WriterParams   []*db.CreateWriterParams   `json:"writerParams" binding:"omitempty"`
	LabelParams    []*db.CreateLabelParams    `json:"labelParams" binding:"omitempty"`
}

var songHashes = make(map[string]bool)

func (server *Server) addSong(ctx *gin.Context) {
	var req addSongTxRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validAccount(ctx, req.AccountID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you have no permission to add a song. Must be an artist, producer, writer or label account"})
		return
	}

	// Normalize the song parameters
	normalizedParams := normalizeParams(req)

	// Compute the hash of the song data
	hash := computeSongHash(normalizedParams)
	// Check if a song with the same hash already exists
	if songHashes[hash] {
		ctx.JSON(http.StatusConflict, gin.H{"error": "A song with these parameters already exists."})
		return
	}

	// Add the hash to the map
	songHashes[hash] = true

	arg := db.AddSongTxParams{
		AccountID:      req.AccountID,
		SongParams:     req.SongParams,
		GenreParams:    req.GenreParams,
		ArtistParams:   req.ArtistParams,
		AlbumParams:    req.AlbumParams,
		ProducerParams: req.ProducerParams,
		WriterParams:   req.WriterParams,
		LabelParams:    req.LabelParams,
	}

	result, err := server.store.AddSongTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

type UpdateSongTxRequest struct {
	AccountID      int64                      `json:"accountID" binding:"required,min=1"`
	SongParams     db.UpdateSongParams        `json:"songParams" binding:"required"`
	GenreParams    []*db.CreateGenreParams    `json:"genreParams" binding:"omitempty"`
	ArtistParams   []*db.CreateArtistParams   `json:"artistParams" binding:"omitempty"`
	AlbumParams    []*db.CreateAlbumParams    `json:"albumParams" binding:"omitempty"`
	ProducerParams []*db.CreateProducerParams `json:"producerParams" binding:"omitempty"`
	WriterParams   []*db.CreateWriterParams   `json:"writerParams" binding:"omitempty"`
	LabelParams    []*db.CreateLabelParams    `json:"labelParams" binding:"omitempty"`
}

func (server *Server) updateSong(ctx *gin.Context) {
	var req UpdateSongTxRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validAccount(ctx, req.AccountID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you have no permission to update a song. Must be an artist, producer, writer or label account"})
		return
	}

	arg := db.UpdateSongTxParams{
		AccountID:      req.AccountID,
		SongParams:     req.SongParams,
		GenreParams:    req.GenreParams,
		ArtistParams:   req.ArtistParams,
		AlbumParams:    req.AlbumParams,
		ProducerParams: req.ProducerParams,
		WriterParams:   req.WriterParams,
		LabelParams:    req.LabelParams,
	}

	result, err := server.store.UpdateSongTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

type deleteSongRequest struct {
	AccountID int64 `json:"accountID" binding:"required,min=1"`
}

type deleteSongUri struct {
	SongID int64 `uri:"id" binding:"required,min=1"`
}

// TODO: only delete if account owns song
func (server *Server) deleteSong(ctx *gin.Context) {
	var uri deleteSongUri
	var req deleteSongRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validAccount(ctx, req.AccountID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you have no permission to delete a song. Must be an artist, producer, writer or label account"})
		return
	}

	arg := db.DeleteSongTxParams{
		SongID: uri.SongID,
	}
	err := server.store.DeleteSongTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "song deleted"})
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64) bool {
	// Fetch the account types of the user
	accountTypes, err := server.store.GetAccountTypesForAccount(ctx, accountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	// Check if the account is an artist
	for _, at := range accountTypes {
		if at.IsArtist {
			return true
		}
	}

	return false
}
