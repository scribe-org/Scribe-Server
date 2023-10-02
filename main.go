package main

import (
	"fmt"

	"scribe-org/scribe-server/internal/handler"

	"github.com/spf13/viper"
)

func main() {

	// Read in the config file.
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	handler.HandleRequests()
}
