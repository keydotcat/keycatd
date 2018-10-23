package db

import (
	"database/sql"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/keydotcat/keycatd/static"
	"github.com/keydotcat/keycatd/util"
)

type MigrateMgr struct {
	db         *sql.DB
	dbType     string
	migrations map[int]string
}

func NewMigrateMgr(db *sql.DB, dbType string) *MigrateMgr {
	return &MigrateMgr{db, dbType, make(map[int]string)}
}

func (m *MigrateMgr) LoadMigrations() error {
	return static.Walk("migrations/"+m.dbType, func(path string, fi os.FileInfo, err error) error {
		log.Println("Found migration", path)
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		data, err := static.Asset(path)
		if err != nil {
			return util.NewErrorFrom(err)
		}
		b := strings.LastIndex(path, "/")
		e := strings.Index(path[b:], "_")
		if b == -1 || e == -1 {
			return util.NewErrorf("Could not find migration id for %s. Seems to have an invalid format", path)
		}
		idx, err := strconv.Atoi(path[b+1 : b+e])
		if err != nil {
			return util.NewErrorf("Could not parse number for db migration %s: %s", path, err)
		}
		m.migrations[idx] = string(data)
		return nil
	})
}

func (m *MigrateMgr) GetLastMigrationInstalled() (int, error) {
	if exists, err := m.checkIfMigrationsTableExists(); err != nil {
		return 0, err
	} else if !exists {
		log.Println("Creating migrations table")
		if err = m.createMigrationsTable(); err != nil {
			return 0, err
		}
	}
	var mid int
	err := m.db.QueryRow("SELECT \"Id\" FROM \"db_migrations\" ORDER BY \"Id\" DESC LIMIT 1").Scan(&mid)
	if err == sql.ErrNoRows {
		return 0, nil
	} else if err != nil {
		return 0, util.NewErrorf("Could not retrieve last migration installed: %s", err)
	}
	return mid, nil
}

func (m *MigrateMgr) CheckIfMigrationIsRequired() (int, error) {
	mid, err := m.GetLastMigrationInstalled()
	if err != nil {
		return 0, err
	}
	required := 0
	for kid := range m.migrations {
		if kid > mid {
			required += 1
		}
	}
	return required, nil
}

func (m *MigrateMgr) ApplyRequiredMigrations() (int, int, error) {
	lid, err := m.GetLastMigrationInstalled()
	if err != nil {
		return 0, 0, err
	}
	ids := make([]int, 0, len(m.migrations))
	for kid := range m.migrations {
		ids = append(ids, kid)
	}
	sort.Ints(ids)
	applied := 0
	for _, mid := range ids {
		if mid <= lid {
			continue
		}
		if err = m.applyMigration(mid); err != nil {
			return 0, 0, err
		}
		lid = mid
		applied += 1
	}
	return lid, applied, nil
}

func (m *MigrateMgr) applyMigration(mid int) error {
	tx, err := m.db.Begin()
	if err != nil {
		return util.NewErrorFrom(err)
	}
	data, ok := m.migrations[mid]
	if !ok {
		panic("Requested migration of non existant id")
	}
	_, err = tx.Exec(data)
	if err != nil {
		return util.NewErrorf("Could not process migration %d: %s", mid, err)
	}
	_, err = tx.Exec(`INSERT INTO "db_migrations" ("Id","CreatedAt") VALUES ($1,$2)`, mid, time.Now().UTC())
	if err != nil {
		return util.NewErrorf("Could not write executed migration to table: %s", err)
	}
	return util.NewErrorFrom(tx.Commit())
}

func (m *MigrateMgr) checkIfMigrationsTableExists() (bool, error) {
	var query string
	switch m.dbType {
	case "cockroach":
		query = "SHOW TABLES"
	case "postgresql":
		query = `SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname != 'pg_catalog' AND schemaname != 'information_schema'`
	default:
		return false, util.NewErrorf("Unknown database type: %s", m.dbType)
	}
	rows, err := m.db.Query(query)
	if err != nil {
		return false, util.NewErrorf("Could not retrieve tables: %s", err)
	}
	defer rows.Close()
	var name string
	for rows.Next() {
		if err := rows.Scan(&name); err != nil {
			return false, util.NewErrorf("Could not retrieve table name: %s", err)
		}
		if name == "db_migrations" {
			return true, nil
		}
	}
	return false, nil
}

func (m *MigrateMgr) createMigrationsTable() error {
	var query string
	switch m.dbType {
	case "cockroach":
		query = `CREATE TABLE "db_migrations" ("Id" INT NOT NULL, "CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL, CONSTRAINT "primary" PRIMARY KEY ("Id" DESC), FAMILY "primary" ("Id", "CreatedAt") )`
	case "postgresql":
		query = `CREATE TABLE "db_migrations" ("Id" INT NOT NULL, "CreatedAt" TIMESTAMP WITH TIME ZONE NOT NULL, CONSTRAINT "primary" PRIMARY KEY ("Id") )`
	default:
		return util.NewErrorf("Unknown database type: %s", m.dbType)
	}
	_, err := m.db.Exec(query)
	if err != nil {
		return util.NewErrorf("Could not create migrations table: %s", err)
	}
	return nil
}
