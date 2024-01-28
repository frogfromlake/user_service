package util

import (
	"bytes"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/argon2"
)

type Argon2idHash struct {
	// time represents the number of passed over the specified memory.
	time uint32
	// cpu memory to be used.
	memory uint32
	// threads for parallelism aspect of the algorithm.
	threads uint8
	// keyLen of the generate hash key.
	keyLen uint32
	// saltLen the length of the salt used.
	saltLen uint32
}

// NewArgon2idHash constructor function for Argon2idHash.
func newArgon2idHash(time, saltLen uint32, memory uint32, threads uint8, keyLen uint32) *Argon2idHash {
	return &Argon2idHash{
		time:    time,
		saltLen: saltLen,
		memory:  memory,
		threads: threads,
		keyLen:  keyLen,
	}
}

func randomSecret(length uint32) ([]byte, error) {
	secret := make([]byte, length)

	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// GenerateHash using the password and provided salt. If no salt value is provided, fallback
// to a random value generated of a given length.
type HashSalt struct {
	Hash []byte
	Salt []byte
}

func (a *Argon2idHash) generateHash(password, salt []byte) (*HashSalt, error) {
	var err error
	// If salt is not provided generate a salt of
	// the configured salt length.
	if len(salt) == 0 {
		salt, err = randomSecret(a.saltLen)
	}
	if err != nil {
		return nil, err
	}
	// Generate hash
	hash := argon2.IDKey(password, salt, a.time, a.memory, a.threads, a.keyLen)
	// Return the generated hash and salt used for storage.
	return &HashSalt{Hash: hash, Salt: salt}, nil
}

// Compare generated hash with store hash.
func (a *Argon2idHash) compare(hash, salt, password []byte) error {
	// Generate hash for comparison.
	hashSalt, err := a.generateHash(password, salt)
	if err != nil {
		return err
	}
	// Compare the generated hash with the stored hash.
	// If they don't match return error.
	if !bytes.Equal(hash, hashSalt.Hash) {
		return errors.New("invalid password")
	}
	return nil
}

type Argon2idParams struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

// HashPassword generates a hash of the given password using the Argon2id algorithm.
// Returns a HashSalt struct containing the generated hash and salt.
func HashPassword(password string) (*HashSalt, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	defaultParams := Argon2idParams{
		Time:    2,
		Memory:  19, // in MiB
		Threads: 1,
		KeyLen:  32, // in bytes
		SaltLen: 16, // in bytes
	}
	argon2idHash := newArgon2idHash(defaultParams.Time, defaultParams.SaltLen, defaultParams.Memory, defaultParams.Threads, defaultParams.KeyLen)
	hash, err := argon2idHash.generateHash([]byte(password), nil)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// ComparePassword compares a given hash, salt, and password.
// If the hash of the password matches the given hash, it returns nil.
// Otherwise, it returns an error indicating the password is invalid.
func ComparePassword(hash, salt []byte, password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	defaultParams := Argon2idParams{
		Time:    2,
		Memory:  19, // in MiB
		Threads: 1,
		KeyLen:  32, // in bytes
		SaltLen: 16, // in bytes
	}
	argon2idHash := newArgon2idHash(defaultParams.Time, defaultParams.SaltLen, defaultParams.Memory, defaultParams.Threads, defaultParams.KeyLen)

	hashBytes := []byte(hash)
	saltBytes := []byte(salt)
	passwordBytes := []byte(password)

	err := argon2idHash.compare(hashBytes, saltBytes, passwordBytes)
	if err != nil {
		return err
	}
	return nil
}
