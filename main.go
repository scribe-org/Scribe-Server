// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"fmt"

	"github.com/scribe-org/scribe-server/api"
	"github.com/spf13/viper"

	_ "github.com/scribe-org/scribe-server/docs"
)

// @title           Scribe Server API
// @version         1.0
// @description     Scribe-Server is a backend service that provides the API by which data is available for download within Scribe apps.
// @termsOfService  https://github.com/scribe-org/Scribe-Server/blob/main/.github/CODE_OF_CONDUCT.md

// @contact.name   Scribe Team
// @contact.url    https://scri.be/
// @contact.email  team@scri.be

// @license.name  GPL-3.0 license
// @license.url   https://github.com/scribe-org/Scribe-Server/blob/main/LICENSE.txt

// @host      localhost:8080
// @BasePath  /

// @externalDocs.description  GitHub Repository
// @externalDocs.url          https://github.com/scribe-org/Scribe-Server
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
