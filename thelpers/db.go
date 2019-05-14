package thelpers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lib/pq"
	"github.com/luna-duclos/instrumentedsql"
)

func GetTestDBType() string {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if pair[0] == "KEYCAT_TEST_DB" {
			return strings.ToLower(pair[1])
		}
	}
	return "postgresql"
}

func GetDBConnString() string {
	switch GetTestDBType() {
	case "postgresql":
		return "sslmode=disable dbname=test"
	case "cockroachdb":
		return "user=root dbname=test sslmode=disable port=26257"
	}
	panic("Unkown db type from env: " + GetTestDBType())
}

func GetDBConn() *sql.DB {
	dbType := "postgres"
	if len(os.Getenv("dblog")) > 0 {
		logger := instrumentedsql.LoggerFunc(func(ctx context.Context, msg string, keyvals ...interface{}) {
			switch msg {
			case "sql-conn-exec", "sql-conn-query":
				log.Printf("SQL: %v", keyvals)
			}
		})
		dbType = "instrumented-postgres"
		sql.Register(dbType, instrumentedsql.WrapDriver(
			&pq.Driver{},
			instrumentedsql.WithOpsExcluded(instrumentedsql.OpSQLRowsNext, instrumentedsql.OpSQLTxBegin, instrumentedsql.OpSQLTxCommit),
			instrumentedsql.WithLogger(logger),
		))
	}
	mdb, err := sql.Open(dbType, GetDBConnString())
	if err != nil {
		panic(err)
	}
	return mdb
}

func getCockroachDBTables(db *sql.DB) []string {
	tables := []string{}
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		tables = append(tables, name)
	}
	return tables
}

func getPsqlTables(db *sql.DB) []string {
	tables := []string{}
	rows, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		tables = append(tables, name)
	}
	return tables
}

func DropAllTables(db *sql.DB) {
	fmt.Println("Dropping all tables")
	var tables []string
	switch GetTestDBType() {
	case "postgresql":
		tables = getPsqlTables(db)
	case "cockroachdb":
		tables = getCockroachDBTables(db)
	}
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	for _, tname := range tables {
		if _, err = tx.Exec("DROP TABLE \"" + tname + "\" CASCADE"); err != nil {
			panic(err)
		}
	}
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}
