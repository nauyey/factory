package factory

import (
	"database/sql"
	"reflect"
)

var dbConnection *sql.DB
var dbDriverName string

// SetDB sets database connection and a driver name for factory
func SetDB(db *sql.DB) {
	dbConnection = db
	// setting a driver name
	dr := dbConnection.Driver()
	dv := reflect.ValueOf(dr)
	dbDriverName = dv.Type().String()
}

func getDB() *sql.DB {
	return dbConnection
}

func GetDBDriverName() string {
	return dbDriverName
}
