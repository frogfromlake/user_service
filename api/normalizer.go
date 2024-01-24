package api

import (
	"strings"

	"github.com/jinzhu/copier"
)

func normalizeParams(params addSongTxRequest) addSongTxRequest {
	// Create a deep copy of the params
	var normalized addSongTxRequest
	copier.Copy(&normalized, &params)

	// Trim whitespace and convert to lowercase for string parameters
	normalized.SongParams.Title = strings.TrimSpace(strings.ToLower(params.SongParams.Title))
	normalized.SongParams.SongCoverUrl.String = strings.TrimSpace(strings.ToLower(params.SongParams.SongCoverUrl.String))
	normalized.SongParams.AudioFileUrl = strings.TrimSpace(strings.ToLower(params.SongParams.AudioFileUrl))

	for _, genreParam := range normalized.GenreParams {
		genreParam.Name = strings.TrimSpace(strings.ToLower(genreParam.Name))
	}

	for _, artistParam := range normalized.ArtistParams {
		artistParam.Name = strings.TrimSpace(strings.ToLower(artistParam.Name))
		artistParam.Bio.String = strings.TrimSpace(strings.ToLower(artistParam.Bio.String))
		artistParam.CountryCode = strings.TrimSpace(strings.ToLower(artistParam.CountryCode))
	}

	for _, albumParam := range normalized.AlbumParams {
		albumParam.Title = strings.TrimSpace(strings.ToLower(albumParam.Title))
		albumParam.AlbumCoverUrl.String = strings.TrimSpace(strings.ToLower(albumParam.AlbumCoverUrl.String))
	}

	for _, producerParam := range normalized.ProducerParams {
		producerParam.Name = strings.TrimSpace(strings.ToLower(producerParam.Name))
		producerParam.CountryCode = strings.TrimSpace(strings.ToLower(producerParam.CountryCode))
	}

	for _, writerParam := range normalized.WriterParams {
		writerParam.Name = strings.TrimSpace(strings.ToLower(writerParam.Name))
		writerParam.CountryCode = strings.TrimSpace(strings.ToLower(writerParam.CountryCode))
	}

	for _, labelParam := range normalized.LabelParams {
		labelParam.Name = strings.TrimSpace(strings.ToLower(labelParam.Name))
		labelParam.CountryCode = strings.TrimSpace(strings.ToLower(labelParam.CountryCode))
		labelParam.LabelCoverUrl.String = strings.TrimSpace(strings.ToLower(labelParam.LabelCoverUrl.String))
	}

	return normalized
}
