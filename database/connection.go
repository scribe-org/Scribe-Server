// SPDX-License-Identifier: GPL-3.0-or-later

// Package database manages MySQL database connection and access functions.
package database

import (
	"database/sql"
	"fmt"
	"log"

	// Import MySQL driver for side effects (driver registration).
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// DB is the global database connection used across the application.
// It should be initialized via InitDatabase before use.
var DB *sql.DB

// InitDatabase initializes the database connection.
func InitDatabase() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.host"),
		viper.GetString("database.port"),
		viper.GetString("database.name"),
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	log.Println("✅ Database connection established")
	return nil
}
