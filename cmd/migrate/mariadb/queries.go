// SPDX-License-Identifier: GPL-3.0-or-later
package mariadb

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/scribe-org/scribe-server/cmd/migrate/schema"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
)

// ExecuteBatch executes a batch of SQL statements.
func ExecuteBatch(stmt *sql.Stmt, batch [][]interface{}) error {
	for _, values := range batch {
		if _, err := stmt.Exec(values...); err != nil {
			return err
		}
	}
	return nil
}

// MigrateTable migrates a single table from SQLite to MariaDB.
func MigrateTable(sqlite *sql.DB, mariaDB *sql.DB, langCode, tableName string) error {
	log.Printf("Migrating table %s for language %s", tableName, langCode)

	// Get table schema.
	tableSchema, err := schema.GetTableSchema(sqlite, tableName)
	if err != nil {
		return fmt.Errorf("failed to get schema: %v", err)
	}

	// Generate table names.
	mariaTableName := fmt.Sprintf("%s_%s", langCode, strings.TrimPrefix(tableName, "sqlite_"))
	backupTableName := mariaTableName + "_old"

	// Check if table exists and rename it to backup.
	exists, err := tableExists(mariaDB, mariaTableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %v", err)
	}

	if exists {
		// Drop old backup table if it exists.
		_, _ = mariaDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", backupTableName))

		// Rename existing table to backup.
		if _, err := mariaDB.Exec(fmt.Sprintf("RENAME TABLE `%s` TO `%s`", mariaTableName, backupTableName)); err != nil {
			return fmt.Errorf("failed to rename existing table: %v", err)
		}
		log.Printf("Existing table renamed to %s", backupTableName)
	}

	// Create new table.
	createSQL := schema.GenerateCreateTableSQL(mariaTableName, tableSchema)
	if _, err := mariaDB.Exec(createSQL); err != nil {
		// If creation fails and we had a backup, restore it.
		if exists {
			if _, restoreErr := mariaDB.Exec(fmt.Sprintf("RENAME TABLE `%s` TO `%s`", backupTableName, mariaTableName)); restoreErr != nil {
				return fmt.Errorf("failed to create table and restore backup: original error: %v, restore error: %v", err, restoreErr)
			}
		}
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Perform the data migration.
	if err := performDataMigration(sqlite, mariaDB, tableSchema, tableName, mariaTableName); err != nil {
		// If migration fails and we had a backup, restore it.
		if exists {
			// Clean up the failed new table.
			_, _ = mariaDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", mariaTableName))

			if _, restoreErr := mariaDB.Exec(fmt.Sprintf("RENAME TABLE `%s` TO `%s`", backupTableName, mariaTableName)); restoreErr != nil {
				return fmt.Errorf("failed to migrate data and restore backup: original error: %v, restore error: %v", err, restoreErr)
			}
		}
		return fmt.Errorf("failed to migrate data: %v", err)
	}

	// If everything succeeded, drop the backup table.
	if exists {
		if _, err := mariaDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", backupTableName)); err != nil {
			log.Printf("Warning: Failed to drop backup table %s: %v", backupTableName, err)
		} else {
			log.Printf("Backup table %s dropped successfully", backupTableName)
		}
	}

	return nil
}

// tableExists checks if a table exists in the MariaDB database.
func tableExists(db *sql.DB, tableName string) (bool, error) {
	var exists int
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// performDataMigration handles the actual data transfer between databases.
func performDataMigration(sqlite *sql.DB, mariaDB *sql.DB, schema *types.TableSchema, srcTable, destTable string) error {
	columns := "`" + strings.Join(schema.ColumnNames, "`, `") + "`"
	rows, err := sqlite.Query(fmt.Sprintf("SELECT %s FROM `%s`", columns, srcTable))
	if err != nil {
		return fmt.Errorf("failed to select data: %v", err)
	}
	defer rows.Close()

	tx, err := mariaDB.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Set up defer to rollback on error, but we'll commit explicitly on success.
	var committed bool
	defer func() {
		if !committed {
			if err := tx.Rollback(); err != nil {
				log.Printf("Error rolling back transaction: %v", err)
			}
		}
	}()

	placeholders := strings.Repeat("?,", len(schema.ColumnNames))
	placeholders = placeholders[:len(placeholders)-1]
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT IGNORE INTO `%s` (%s) VALUES (%s)",
		destTable, columns, placeholders))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	if err := processBatches(rows, stmt, schema.ColumnNames, destTable); err != nil {
		return err
	}

	// Commit the transaction.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	return nil
}

// processBatches handles processing rows in batches.
func processBatches(rows *sql.Rows, stmt *sql.Stmt, columnNames []string, tableName string) error {
	batchSize := 5000
	batch := make([][]interface{}, 0, batchSize)
	count := 0

	for rows.Next() {
		scanArgs := make([]interface{}, len(columnNames))
		for i := range scanArgs {
			var value interface{}
			scanArgs[i] = &value
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		values := make([]interface{}, len(scanArgs))
		for i, arg := range scanArgs {
			values[i] = *arg.(*interface{})
		}

		batch = append(batch, values)

		// Execute batch insert when batch is full.
		if len(batch) >= batchSize {
			if err := ExecuteBatch(stmt, batch); err != nil {
				return fmt.Errorf("failed to execute batch: %v", err)
			}
			count += len(batch)
			batch = batch[:0] // clear batch
			log.Printf("Migrated %d rows for table %s", count, tableName)
		}
	}

	// Insert remaining rows.
	if len(batch) > 0 {
		if err := ExecuteBatch(stmt, batch); err != nil {
			return fmt.Errorf("failed to execute final batch: %v", err)
		}
		count += len(batch)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	log.Printf("Completed migration of %d rows for table %s", count, tableName)
	return nil
}
