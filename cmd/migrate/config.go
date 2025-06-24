// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"fmt"

	"github.com/scribe-org/scribe-server/cmd/migrate/types"
	"github.com/spf13/viper"
)

// LoadConfig reads and unmarshals the YAML config file using Viper.
func LoadConfig(path string) (*types.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config types.Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return &config, nil
}
