package db

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/keydotcat/keycatd/thelpers"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	db = thelpers.GetDBConn()
	thelpers.DropAllTables(db)
}

func addMigration(m *MigrateMgr) {
	i := len(m.migrations) + 1
	m.migrations[i] = fmt.Sprintf(`
	CREATE TABLE "a%d" ( "something" INT );
	CREATE TABLE "b%d" ( "else" TEXT );`, i, i)
}

func TestMigrations(t *testing.T) {
	defer thelpers.DropAllTables(db)
	m := NewMigrateMgr(db, "postgresql")
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
	lid, ap, err := m.ApplyRequiredMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if lid != 1 {
		t.Fatalf("Expected migrations up to 1 and got %d", lid)
	}
	if ap != 1 {
		t.Fatalf("Expected migrations up to 1 and got %d", ap)
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
	lid, ap, err = m.ApplyRequiredMigrations()
	if err != nil {
		t.Fatal(err)
	}
	if lid != 3 {
		t.Fatalf("Expected migrations up to 3 and got %d", lid)
	}
	if ap != 2 {
		t.Fatalf("Expected to run 2 migrations and got %d", ap)
	}
}
