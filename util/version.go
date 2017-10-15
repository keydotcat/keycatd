package util

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/keydotcat/backend/static"
)

type VersionEntry struct {
	Commit string    `json:"commit"`
	Date   time.Time `json:"date"`
	Behind int       `json:"-"`
}

var (
	currentVersion *VersionEntry
	versionHistory map[string]*VersionEntry
)

func init() {
	vs := []*VersionEntry{}
	data, err := static.Asset("data/version/history")
	if err != nil {
		panic(fmt.Sprintf("Cannot retrieve version_history: %s", err))
	}
	if err = json.Unmarshal(data, &vs); err != nil {
		panic(err)
	}
	currentVersion = vs[0]
	versionHistory = make(map[string]*VersionEntry)
	for i, v := range vs {
		v.Behind = i
		versionHistory[v.Commit] = v
	}
}

func GetVersion() string {
	return currentVersion.Commit
}

func GetVersionInfo(cid string) (*VersionEntry, error) {
	v, ok := versionHistory[cid]
	if !ok {
		return nil, NewErrorf("Unknown version %s", cid)
	}
	return v, nil
}
