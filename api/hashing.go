package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
)

func computeSongHash(params addSongTxRequest) string {
	// Convert the struct to a JSON string
	jsonStr, err := json.Marshal(params)
	if err != nil {
		log.Fatal(err)
	}

	// Compute the SHA-256 hash of the JSON string
	hash := sha256.Sum256([]byte(jsonStr))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr
}