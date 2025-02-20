package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/scribe-org/scribe-server/cmd/migrate/queries"
	"github.com/scribe-org/scribe-server/cmd/migrate/types"
)

// ExecuteBatch executes a batch of SQL statements
func ExecuteBatch(stmt *sql.Stmt, batch [][]interface{}) error {
	for _, values := range batch {
		if _, err := stmt.Exec(values...); err != nil {
			return err
		}
	}
	return nil
}

// MigrateTable migrates a single table from SQLite to MariaDB
func MigrateTable(sqlite *sql.DB, mariaDB *sql.DB, langCode, tableName string) error {
	log.Printf("Migrating table %s for language %s", tableName, langCode)

	// Get table schema
	schema, err := queries.GetTableSchema(sqlite, tableName)
	if err != nil {
		return fmt.Errorf("failed to get schema: %v", err)
	}

	// Create table in MariaDB
	mariaTableName := fmt.Sprintf("%s_%s", langCode, strings.TrimPrefix(tableName, "sqlite_"))
	createSQL := queries.GenerateCreateTableSQL(mariaTableName, schema)
	if _, err := mariaDB.Exec(createSQL); err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return performDataMigration(sqlite, mariaDB, schema, tableName, mariaTableName)
}

// performDataMigration handles the actual data transfer between databases
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
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(schema.ColumnNames))
	placeholders = placeholders[:len(placeholders)-1]
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT IGNORE INTO `%s` (%s) VALUES (%s)", 
		destTable, columns, placeholders))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	return processBatches(rows, stmt, schema.ColumnNames, destTable)
}

// processBatches handles processing rows in batches
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
		
		// Execute batch insert when batch is full
		if len(batch) >= batchSize {
			if err := ExecuteBatch(stmt, batch); err != nil {
				return fmt.Errorf("failed to execute batch: %v", err)
			}
			count += len(batch)
			batch = batch[:0] // Clear batch
			log.Printf("Migrated %d rows for table %s", count, tableName)
		}
	}

	// Insert remaining rows
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