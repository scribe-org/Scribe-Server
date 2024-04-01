package main

import (
	"fmt"

	"github.com/scribe-org/scribe-server/api"

	"github.com/spf13/viper"
)

func main() {

	// Read in the config file.
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		_, configFileNotFound := err.(viper.ConfigFileNotFoundError)

		// Config file not found; try reading config from environment variables.
		if configFileNotFound {
			viper.AutomaticEnv()

			// Environment variables also not set.
			if !viper.IsSet("hostPort") || !viper.IsSet("fileSystem") {
				panic(fmt.Errorf("fatal error config environment: %w", err))
			}
		} else {

			// Config file was found, but another error was produced.
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	api.HandleRequests()
}
