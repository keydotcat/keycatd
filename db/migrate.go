package db

import (
	"database/sql"
	"os"
	"strconv"
	"strings"

	"github.com/keydotcat/backend/static"
	"github.com/keydotcat/backend/util"
)

type MigrateMgr struct {
	assets map[int][]byte
}

func NewMigrateMgr(db *sql.DB) error {
	return &MigrateMgr{db, make(map[int][]byte)}
}

func (m *MigrateMgr) loadAssets() error {
	return static.Walk("data/migrations", func(path string, fi os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		data, err := Asset(path)
		if err != nil {
			return util.NewErrorFrom(err)
		}
		b := string.LastIndex(path, "/")
		e := string.FirstIndex(path[b:], "_")
		idx, err := strconv.Atoi(path[b:e])
		if err != nil {
			return util.NewErrorf("Could not parse number for db migration %s: %s", path, err)
		}
		assets[idx] = data
		return nil
	})
}

func (m *MigrateMgr) GetLastMigrationInstalled() (int, error) {

}
