// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/scribe-org/scribe-server/api"
	"github.com/spf13/viper"

	_ "github.com/scribe-org/scribe-server/docs"
)

// MARK: Environment Logging

// logEnvironment prints out key runtime environment info such as ENV and GIN_MODE.
func logEnvironment() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	log.Println("========================================")
	log.Printf("🌍 Environment: %s", env)
	log.Printf("🚀 GIN_MODE: %s", ginMode)
	log.Println("========================================")
	fmt.Println()
}

// MARK: Swagger Annotations
//
// These annotations define Scribe Server's API documentation,
// versioning, and contact/license information.

// @title           Scribe Server API
// @version         1.0
// @description     Scribe-Server is a backend service that provides the API by which data is available for download within Scribe apps.
// @termsOfService  https://github.com/scribe-org/Scribe-Server/blob/main/.github/CODE_OF_CONDUCT.md
// @contact.name    Scribe Team
// @contact.url     https://scri.be/
// @contact.email   team@scri.be
// @license.name    GPL-3.0 license
// @license.url     https://github.com/scribe-org/Scribe-Server/blob/main/LICENSE.txt
// @host            scribe-server.toolforge.org
// @BasePath        /
// @externalDocs.description  GitHub Repository
// @externalDocs.url          https://github.com/scribe-org/Scribe-Server

// MARK: Config Setup

// loadConfig reads the configuration from config.yaml, or from environment variables if there's no config file.
func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		_, configFileNotFound := err.(viper.ConfigFileNotFoundError)

		// Config file not found; fall back to environment variables.
		if configFileNotFound {
			viper.AutomaticEnv()

			// Environment variables also not set.
			if !viper.IsSet("hostPort") || !viper.IsSet("fileSystem") {
				return fmt.Errorf("fatal error config environment: %w", err)
			}
		} else {
			// Config file was found, but another error occurred.
			return fmt.Errorf("fatal error config file: %w", err)
		}
	}
	viper.SetDefault("contractsDir", "./contracts")

	return nil
}

// MARK: Main Entry Point

// handleRequests starts the server. It's a package-level variable to allow mocking in unit tests.
var handleRequests = api.HandleRequests

func main() {
	logEnvironment()

	// Read configuration from file or environment variables.
	if err := loadConfig(); err != nil {
		panic(err)
	}

	// MARK: Start Server

	handleRequests()
}
