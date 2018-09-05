package managers

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"time"

	"github.com/golang/snappy"
	"github.com/keydotcat/backend/util"
)

type Session struct {
	Id           string    `json:"id" scaneo:"pk"`
	User         string    `json:"user"`
	Agent        string    `json:"agent"`
	RequiresCSRF bool      `json:"csrf_required"`
	LastAccess   time.Time `json:"last_access"`
	StoreToken   string    `json:"-"`
}

func encodeSession(buf *bytes.Buffer, s *Session) error {
	b64Sink := base64.NewEncoder(base64.RawStdEncoding, buf)
	snappySink := snappy.NewBufferedWriter(b64Sink)
	if err := gob.NewEncoder(snappySink).Encode(s); err != nil {
		return util.NewErrorFrom(err)
	}
	if err := snappySink.Close(); err != nil {
		return util.NewErrorFrom(err)
	}
	return util.NewErrorFrom(b64Sink.Close())
}

func decodeSession(buf *bytes.Buffer, s *Session) error {
	b64Source := base64.NewDecoder(base64.RawStdEncoding, buf)
	snappySource := snappy.NewReader(b64Source)
	return util.NewErrorFrom(gob.NewDecoder(snappySource).Decode(s))
}

type SessionMgr interface {
	NewSession(userId string, agent string, csrf bool) (*Session, error)
	UpdateSession(id, agent string) (*Session, error)
	GetSession(id string) (*Session, error)
	DeleteSession(id string) error
	GetAllSessions(userId string) ([]*Session, error)
	DeleteAllSessions(userId string) error
}
