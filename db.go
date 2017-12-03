package factory

import "database/sql"

var dbConnection *sql.DB

// SetDB sets database connection for factory
func SetDB(db *sql.DB) {
	dbConnection = db
}

func getDB() *sql.DB {
	return dbConnection
}
