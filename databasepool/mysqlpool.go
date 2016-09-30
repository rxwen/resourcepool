// Package databasepool provides ...
package databasepool

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	gocommon "github.com/rxwen/go-common"
	"github.com/rxwen/resourcepool"
) // blank import mysql driver for sql.

// CreateDBConnectionPool function creates a connection for specified mysql database.
func CreateDBConnectionPool(host, port, username, password, database string, poolSize int, timeoutSecond int) (*resourcepool.ResourcePool, error) {
	connectionString := gocommon.GetMySQLConnectionString(
		host,
		port,
		database,
		username,
		password,
	)
	pool, err := resourcepool.NewResourcePool(host, port, func(host, port string) (interface{}, error) {
		db, err := sql.Open("mysql", connectionString)
		return db, err
	}, func(c interface{}) error {
		err := c.(*sql.DB).Close()
		return err
	}, poolSize, timeoutSecond)
	return pool, err
}
