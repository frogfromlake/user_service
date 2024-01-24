package util

import "github.com/spf13/viper"

// Config is a struct that holds all configurations for the application.
// The values are read by viper from a config file or environment variables.
type Config struct {
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig loads the configuration from the given path.
func LoadConfig(path string) (config Config, err error) {
	viper.SetEnvPrefix("APP")
	// AutomaticEnv makes Viper check if environment variables match any of the existing keys.
	// If matching env vars are found, they are loaded into Viper.
	viper.AutomaticEnv()

	// Check if environment variables are set
	if dbSource := viper.GetString("DB_SOURCE"); dbSource != "" {
		config.DBSource = dbSource
	}
	if serverAddress := viper.GetString("SERVER_ADDRESS"); serverAddress != "" {
		config.ServerAddress = serverAddress
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
