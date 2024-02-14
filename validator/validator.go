package validator

import (
	"fmt"
	"time"
)

// Function to validate a string value based on the minimum and maximum length
func ValidateString(value string, minLen int, maxLen int) error {
	n := len(value)
	if n < minLen || n > maxLen {
		return fmt.Errorf("must contain %d-%d characters", minLen, maxLen)
	}
	return nil
}

// Function to validate UserId
func ValidateUserId(userId int64) error {
	// Assuming the UserId should be a positive integer
	if userId <= 0 {
		return fmt.Errorf("UserId must be a positive integer")
	}
	return nil
}

// ValidateDuration checks if the given string can be parsed as a duration
func ValidateDuration(durationStr string) error {
	_, err := time.ParseDuration(durationStr)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}
	return nil
}
