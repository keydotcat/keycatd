package db

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("postgres", "user=root dbname=test sslmode=disable port=26257")
	if err != nil {
		panic(err)
	}
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
	for _, tname := range tables {
		if _, err = db.Exec("DROP TABLE \"" + tname + "\" CASCADE"); err != nil {
			panic(err)
		}
	}
}

func addMigration(m *MigrateMgr) {
	i := len(m.migrations) + 1
	m.migrations[i] = fmt.Sprintf(`
	CREATE TABLE "a%d" ( "something" int );
	CREATE TABLE "b%d" ( "else" string );`, i, i)
}

func TestMigrations(t *testing.T) {
	m := NewMigrateMgr(db)
	m.migrations = map[int]string{}
	addMigration(m)
	exists, err := m.checkIfMigrationsTableExists()
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatalf("Table should not exist")
	}
	lid, err := m.GetLastMigrationInstalled()
	if err != nil {
		t.Fatal(err)
	}
	if lid != 0 {
		t.Fatalf("Expected 0 as last migration and got %d", lid)
	}
	exists, err = m.checkIfMigrationsTableExists()
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatalf("Table should exist")
	}
	lid, err = m.ApplyRequiredMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if lid != 1 {
		t.Fatalf("Expected migrations up to 1 and got %d", lid)
	}
	addMigration(m)
	addMigration(m)
	req, err := m.CheckIfMigrationIsRequired()
	if err != nil {
		t.Fatal(err)
	}
	if req != 2 {
		t.Fatalf("Expected to require 2 migrations and got %d", req)
	}
	lid, err = m.ApplyRequiredMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if lid != 3 {
		t.Fatalf("Expected migrations up to 3 and got %d", lid)
	}
}
