package factory

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	sqliteDriverName   = "*sqlite3.SQLiteDriver"
	postgresDriverName = "*pq.Driver"
)

// data persistence utils

// insertRow inserts data into database with sql string and values.
// Parameter db represents the target database connection.
// Parameter sql and values will conbined to generate a SQL.
// It returns last insert ID. And it return error if failed to insert data into database.
func insertRow(db *sql.DB, sql string, values ...interface{}) (int64, error) {
	if DebugMode {
		info.Println("INSERT SQL string: ", sql)
		info.Println("INSERT SQL arguments: ", values)
	}

	stmt, err := db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(values...)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastID, nil
}

// selectRow queries data from database.
// db represents the database connection.
// sql and values are conbined to generate a SQL.
// selectFieldPointers will store the data scaned from the query result, the *sql.Row instance.
// It will return errors if failed to query data.
func selectRow(db *sql.DB, sql string, values []interface{}, selectFieldPointers []interface{}) error {
	if DebugMode {
		info.Println("SELECT SQL string: ", sql)
		info.Println("SELECT SQL arguments: ", values)
	}

	return db.QueryRow(sql, values...).Scan(selectFieldPointers...)
}

// insertSQL generates an insert SQL string, like `INSERT INTO table (field1, field2) VALUES (?, ?)`
func insertSQL(table string, fields []string) string {
	params := make([]string, len(fields))

	for i := range params {
		params[i] = param(i + 1)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(fields, ","), strings.Join(params, ","))
}

// selectSQL generates a query SQL string, like `SELECT selectField1, selectField2 FROM table WHERE primaryField=?`
// selectFields declares which fields will be returned in the query.
// primaryFields represents all primary keys of table. They will be use in WERE clause to identify data from table.
func selectSQL(table string, selectFields []string, primaryFields []string) string {
	return fmt.Sprintf("SELECT %s FROM %s %s", strings.Join(selectFields, ","), table, whereClause(primaryFields))
}

// deleteSQL generates a delete SQL.
func deleteSQL(table string, primaryFields []string) string {
	return fmt.Sprintf("DELETE FROM %s %s", table, whereClause(primaryFields))
}

// param returns the parameters symbol used in prepared
// sql statements.
// TODO: The parameter symbol may be different in mysql and postgres.
func param(position int) string {
	if getDBDriverName() == postgresDriverName {
		return fmt.Sprintf("$%d", position)
	}
	return "?"
}

// helper function to generate the whereClause, like "WHERE name=%s AND nick_name=%s"
// section of a SQL statement
func whereClause(fields []string) string {
	whereClause := ""

	for i, field := range fields {
		if i == 0 {
			whereClause = whereClause + "WHERE"
		} else {
			whereClause = whereClause + "AND"
		}

		whereClause = whereClause + fmt.Sprintf(" %s=%s ", field, param(1))
	}

	return strings.Trim(whereClause, " ")
}
