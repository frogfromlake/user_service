package api

import (
	"testing"

	db "github.com/frogfromlake/user_service/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestNormalizeParams(t *testing.T) {
	// Test with valid data
	validData := addSongTxRequest{
		AccountID:      1,
		SongParams:     db.CreateSongParams{Title: "Test Title",},
		GenreParams:    []*db.CreateGenreParams{{Name: "Test Genre"}},
		ArtistParams:   []*db.CreateArtistParams{{Name: "Test Artist"}},
		AlbumParams:    []*db.CreateAlbumParams{{Title: "Test Album"}},
		ProducerParams: []*db.CreateProducerParams{{Name: "Test Producer"}},
		WriterParams:   []*db.CreateWriterParams{{Name: "Test Writer"}},
		LabelParams:    []*db.CreateLabelParams{{Name: "Test Label"}},
	}
	result := normalizeParams(validData)
	require.Equal(t, "test title", result.SongParams.Title)
	require.Equal(t, "test genre", result.GenreParams[0].Name)
	require.Equal(t, "test artist", result.ArtistParams[0].Name)
	require.Equal(t, "test album", result.AlbumParams[0].Title)
	require.Equal(t, "test producer", result.ProducerParams[0].Name)
	require.Equal(t, "test writer", result.WriterParams[0].Name)
	require.Equal(t, "test label", result.LabelParams[0].Name)

	// Test with invalid data
	invalidData := addSongTxRequest{}
	result = normalizeParams(invalidData)
	require.Equal(t, "", result.SongParams.Title)
	require.Len(t, result.GenreParams, 0)
	require.Len(t, result.ArtistParams, 0)
	require.Len(t, result.AlbumParams, 0)
	require.Len(t, result.ProducerParams, 0)
	require.Len(t, result.WriterParams, 0)
	require.Len(t, result.LabelParams, 0)
}
