package token

import "time"

// Maker is an interface for managing tokens
type Maker interface {
	// CreateLocalToken creates a new local token for a specific username and duration
	CreateLocalToken(username string, duration time.Duration) (string, error)

	// VerifyLocalToken checks if the local token is valid or not
	VerifyLocalToken(token string) (*Payload, error)
}
