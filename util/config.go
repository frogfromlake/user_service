package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config is a struct that holds all configurations for the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBSource             string        `mapstructure:"DB_SOURCE_USER_SERVICE"`
	ServerAddress        string        `mapstructure:"SERVER_ADDRESS_USER_SERVICE"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (config Config, err error) {
	viper.SetEnvPrefix("APP")
	// AutomaticEnv makes Viper check if environment variables match any of the existing keys.
	// If matching env vars are found, they are loaded into Viper.
	viper.AutomaticEnv()

	// Check if environment variables are set
	if dbSource := viper.GetString("DB_SOURCE_USER_SERVICE"); dbSource != "" {
		config.DBSource = dbSource
	}
	if serverAddress := viper.GetString("SERVER_ADDRESS_USER_SERVICE"); serverAddress != "" {
		config.ServerAddress = serverAddress
	}
	if tokenSymmetricKey := viper.GetString("TOKEN_SYMMETRIC_KEY"); tokenSymmetricKey != "" {
		config.TokenSymmetricKey = tokenSymmetricKey
	}
	if accessTokenDuration := viper.GetDuration("ACCESS_TOKEN_DURATION"); accessTokenDuration != 0 {
		config.AccessTokenDuration = accessTokenDuration
	}
	if refreshTokenDuration := viper.GetDuration("REFRESH_TOKEN_DURATION"); refreshTokenDuration != 0 {
		config.RefreshTokenDuration = refreshTokenDuration
	}

	// If environment variables are not set, attempt to load from the config file
	if config.DBSource == "" || config.ServerAddress == "" {
		viper.AddConfigPath(path)
		viper.SetConfigName("app")
		viper.SetConfigType("env")

		err = viper.MergeInConfig()
		if err != nil {
			return
		}

		err = viper.Unmarshal(&config)
		if err != nil {
			return
		}
	}

	return
}
