package util

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is a struct that holds all configurations for the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	ServerName           string        `mapstructure:"SERVER_NAME"`
	DBSource             string        `mapstructure:"DB_SOURCE_USER_SERVICE"`
	DBSourceLocal        string        `mapstructure:"DB_SOURCE_USER_SERVICE_LOCAL"`
	MigrationURL         string        `mapstructure:"MIGRATION_URL"`
	HttpServerAddress    string        `mapstructure:"HTTP_SERVER_ADDRESS_USER_SERVICE"`
	GrpcServerAddress    string        `mapstructure:"GRPC_SERVER_ADDRESS_USER_SERVICE"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	CertPem              string        `mapstructure:"CERT_PEM"`
	KeyPem               string        `mapstructure:"KEY_PEM"`
	CaCertPem            string        `mapstructure:"CA_CERT_PEM"`
}

// LoadConfig loads the configuration from the environment variables using viper package.
// If the environment variable is not set, it reads the missing key from the configuration file but only if the application
// is not running in a CI environment. The CI environment is determined by the CI environment variable.
// The configuration file is located in the env directory (app.env).
func LoadConfig() (config Config, err error) {
	viper.AutomaticEnv()

	// Define a list of keys to check
	keys := []string{
		"SERVER_NAME",
		"DB_SOURCE_USER_SERVICE",
		"DB_SOURCE_USER_SERVICE_LOCAL",
		"MIGRATION_URL",
		"HTTP_SERVER_ADDRESS_USER_SERVICE",
		"GRPC_SERVER_ADDRESS_USER_SERVICE",
		"TOKEN_SYMMETRIC_KEY",
		"ACCESS_TOKEN_DURATION",
		"REFRESH_TOKEN_DURATION",
		"CERT_PEM",
		"KEY_PEM",
		"CA_CERT_PEM",
	}

	// Check each environment variable individually and read the missing ones from the env file
	// if the application is not running in a CI environment
	for _, key := range keys {
		if viper.GetString(key) == "" &&
			viper.GetString("CI") != "true" {
			value, err := readEnvFromFile(key)
			if err != nil {
				return config, err
			}
			// Set the value in the config struct
			err = setConfigValue(&config, key, value)
			if err != nil {
				return config, err
			}
		}
	}

	// Load configuration from Viper
	config.ServerName = viper.GetString("SERVER_NAME")
	config.DBSource = viper.GetString("DB_SOURCE_USER_SERVICE")
	config.DBSourceLocal = viper.GetString("DB_SOURCE_USER_SERVICE_LOCAL")
	config.MigrationURL = viper.GetString("MIGRATION_URL")
	config.HttpServerAddress = viper.GetString("HTTP_SERVER_ADDRESS_USER_SERVICE")
	config.GrpcServerAddress = viper.GetString("GRPC_SERVER_ADDRESS_USER_SERVICE")
	config.TokenSymmetricKey = viper.GetString("TOKEN_SYMMETRIC_KEY")
	config.AccessTokenDuration = viper.GetDuration("ACCESS_TOKEN_DURATION")
	config.RefreshTokenDuration = viper.GetDuration("REFRESH_TOKEN_DURATION")
	certPemPath := viper.GetString("CERT_PEM")
	keyPemPath := viper.GetString("KEY_PEM")
	caCertPemPath := viper.GetString("CA_CERT_PEM")

	if viper.GetString("CONTAINER_ENV") == "true" && viper.Get("CI") != "true" {
		if !strings.HasPrefix(certPemPath, "/") {
			certPemPath = "/" + certPemPath
		}
		if !strings.HasPrefix(keyPemPath, "/") {
			keyPemPath = "/" + keyPemPath
		}
		if !strings.HasPrefix(caCertPemPath, "/") {
			caCertPemPath = "/" + caCertPemPath
		}
	}

	config.CertPem = certPemPath
	config.KeyPem = keyPemPath
	config.CaCertPem = caCertPemPath

	if viper.GetString("CONTAINER_ENV") != "true" {
		config.DBSource = config.DBSourceLocal
	}
	
	return config, err
}

// readEnvFromFile reads the missing key from the configuration file.
func readEnvFromFile(key string) (string, error) {
	// Determine the project root dynamically.
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../.")

	// Construct the absolute path to the configuration file.
	configPath := filepath.Join(projectRoot, "env", "app.env")

	// Set the absolute path as the config file.
	viper.SetConfigFile(configPath)

	// viper.SetConfigFile("env/app.env")
	if err := viper.ReadInConfig(); err != nil {
		return "", err
	}

	// Directly get the value as a string since we're dealing with a single key
	value := viper.GetString(key)
	if value == "" {
		return "", fmt.Errorf("key %s not found in configuration file", key)
	}

	return value, nil
}

// setConfigValue sets the value in the config struct based on the key.
func setConfigValue(config *Config, key string, value string) error {
	var err error
	switch key {
	case "SERVER_NAME":
		config.ServerName = value
	case "DB_SOURCE_USER_SERVICE":
		config.DBSource = value
	case "DB_SOURCE_USER_SERVICE_LOCAL":
		config.DBSourceLocal = value
	case "MIGRATION_URL":
		config.MigrationURL = value
	case "HTTP_SERVER_ADDRESS_USER_SERVICE":
		config.HttpServerAddress = value
	case "GRPC_SERVER_ADDRESS_USER_SERVICE":
		config.GrpcServerAddress = value
	case "TOKEN_SYMMETRIC_KEY":
		config.TokenSymmetricKey = value
	case "ACCESS_TOKEN_DURATION":
		config.AccessTokenDuration, err = time.ParseDuration(value)
	case "REFRESH_TOKEN_DURATION":
		config.RefreshTokenDuration, err = time.ParseDuration(value)
	case "CERT_PEM":
		config.CertPem = value
	case "KEY_PEM":
		config.KeyPem = value
	case "CA_CERT_PEM":
		config.CaCertPem = value
	}
	return err
}
