package thelpers

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
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
	mdb, err := sql.Open("postgres", GetDBConnString())
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
