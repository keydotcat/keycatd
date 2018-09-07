package util

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/keydotcat/server/static"
)

type VersionTime time.Time

const isotime = "2006-01-02 15:04:05 -0700"

func (vt *VersionTime) UnmarshalJSON(data []byte) error {
	t, err := time.Parse(isotime, strings.Trim(string(data), `"`))
	if err != nil {
		return err
	}
	*vt = (VersionTime)(t)
	return nil
}

type VersionEntry struct {
	Commit string       `json:"commit"`
	Date   *VersionTime `json:"date"`
	Behind int          `json:"-"`
}

var (
	currentVersion *VersionEntry
	versionHistory map[string]*VersionEntry
)

func init() {
	vs := []*VersionEntry{}
	data, err := static.Asset("version/history")
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

func GetServerVersion() string {
	data, err := static.Asset("version/current.server")
	if err != nil {
		return GetVersion()
	}
	return strings.TrimSpace(string(data))
}

func GetWebVersion() string {
	data, err := static.Asset("version/current.web")
	if err != nil {
		return "none"
	}
	return strings.TrimSpace(string(data))
}

func GetVersionInfo(cid string) (*VersionEntry, error) {
	v, ok := versionHistory[cid]
	if !ok {
		return nil, NewErrorf("Unknown version %s", cid)
	}
	return v, nil
}
