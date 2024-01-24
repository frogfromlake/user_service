package util

import "github.com/jackc/pgx/v5/pgtype"

// GetAccountTypes returns a copy of accountTypes slice.
func GetAccountTypeStruct() []AccountType {
	result := make([]AccountType, len(accountTypes))
	copy(result, accountTypes)
	return result
}

const (
	user_account     = 1
	artist_account   = 2
	producer_account = 3
	writer_account   = 4
	label_account    = 5
)

type AccountType struct {
	ID          int64
	Description pgtype.Text
	Permissions []byte
	IsArtist    bool
	IsProducer  bool
	IsWriter    bool
	IsLabel     bool
}

var accountTypes = []AccountType{
	{
		ID:          user_account,
		Description: ConvertToText("Default User Account Type"),
		Permissions: []byte(`{"Upload Songs": "false", "Upload Albums": "false", "Upload Playlists": "false"}`),
		IsArtist:    false,
		IsProducer:  false,
		IsWriter:    false,
		IsLabel:     false,
	},
	{
		ID:          artist_account,
		Description: ConvertToText("Artist Account Type"),
		Permissions: []byte(`{"Upload Songs": "true", "Upload Albums": "true", "Upload Playlists": "true"}`),
		IsArtist:    true,
		IsProducer:  false,
		IsWriter:    false,
		IsLabel:     false,
	},
	{
		ID:          producer_account,
		Description: ConvertToText("Producer Account Type"),
		Permissions: []byte(`{"Upload Songs": "true", "Upload Albums": "true", "Upload Playlists": "true"}`),
		IsArtist:    false,
		IsProducer:  true,
		IsWriter:    false,
		IsLabel:     false,
	},
	{
		ID:          writer_account,
		Description: ConvertToText("Writer Account Type"),
		Permissions: []byte(`{"Upload Songs": "true", "Upload Albums": "true", "Upload Playlists": "true"}`),
		IsArtist:    false,
		IsProducer:  false,
		IsWriter:    true,
		IsLabel:     false,
	},
	{
		ID:          label_account,
		Description: ConvertToText("Label Account Type"),
		Permissions: []byte(`{"Upload Songs": "true", "Upload Albums": "true", "Upload Playlists": "true"}`),
		IsArtist:    false,
		IsProducer:  false,
		IsWriter:    false,
		IsLabel:     true,
	},
}
