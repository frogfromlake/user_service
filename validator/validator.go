package validator

import (
	"fmt"
	"net/mail"
	"strings"
)

const (
	// Define Limit and Offset for the API
	MaxLimit  = 100
	MaxOffset = 100
)

// Function to validate UserId
func ValidateId(userId int64) error {
	// Assuming the UserId should be a positive integer
	if userId <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

// Function to validate a string value based on the minimum and maximum length
func ValidateString(value string, minLen int, maxLen int) error {
	n := len(value)
	if n < minLen || n > maxLen {
		return fmt.Errorf("must contain %d-%d characters", minLen, maxLen)
	}
	return nil
}

func ValidateUsername(username string) error {
	// Should be at least 3 and maximum 24 characters long
	return ValidateString(username, 3, 24)
}

func ValidateFullName(fullName string) error {
	// Should contain at least one surname and lastname separated by a space
	// Should be at least 3 and maximum 24 characters long
	data := strings.Split(fullName, " ")
	if len(data) < 2 {
		return fmt.Errorf("must contain at least one surname and lastname separated by a space")
	}
	for _, name := range data {
		if err := ValidateString(name, 3, 24); err != nil {
			return fmt.Errorf("'%s': %s", name, err.Error())
		}
	}
	return nil
}

func ValidateEmail(email string) error {
	// Should be a valid email address
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("must be a valid email address")
	}
	return nil
}

func ValidatePassword(password string) error {
	// Should be at least 8 characters long
	return ValidateString(password, 8, 64)
}

func ValidateCountryCode(countryCode string) error {
	// Should be a valid country code
	if len(countryCode) != 2 {
		return fmt.Errorf("must be a valid iso3166_1 alpha2 country code")
	}
	return nil
}

// Function to validate UserId
func ValidateRoleId(roleId int64) error {
	// Assuming the UserId should be a positive integer
	if roleId <= 0 {
		return fmt.Errorf("'role_id' must be a positive integer")
	}
	return nil
}

func ValidateStatus(status string) error {
	// Should be either 'active' or 'inactive'
	if status != "active" && status != "inactive" {
		return fmt.Errorf("must be either 'active' or 'inactive'")
	}
	return nil
}

// Function to validate offset parameters
func ValidateLimit(limit int32) error {
	// Define a reasonable range for limit (e.g.,  1-100)
	if limit <  1 || limit >  MaxLimit {
		return fmt.Errorf("limit must be between  1 and  %d", MaxLimit)
	}

	return nil
}

// Function to validate limit parameters
func ValidateOffset(offset int32) error {
	// Offset should not be negative
	if offset <  0 || MaxOffset >  100{
		return fmt.Errorf("offset must be non-negative and less than %d", MaxOffset)
	}

	return nil
}