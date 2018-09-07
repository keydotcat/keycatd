package managers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/keydotcat/server/models"
	"github.com/keydotcat/server/util"
	radix "github.com/mediocregopher/radix.v3"
)

type sessionMgrRedis struct {
	prefix string
	dbId   string
	pool   *radix.Pool
}

func NewSessionMgrRedis(connUrl string, dbId int) (SessionMgr, error) {
	pool, err := radix.NewPool("tcp", connUrl, 10, nil)
	if err != nil {
		return nil, err
	}
	return sessionMgrRedis{"kc-", strconv.Itoa(dbId), pool}, nil
}

func (r sessionMgrRedis) skey(i string) string {
	return fmt.Sprintf("%ss:%s", r.prefix, i)
}

func (r sessionMgrRedis) ukey(i string) string {
	return fmt.Sprintf("%su:%s", r.prefix, i)
}

func (r sessionMgrRedis) NewSession(userId, agent string, csrf bool) (*Session, error) {
	s := &Session{util.GenerateRandomToken(15), userId, agent, csrf, time.Now().UTC(), util.GenerateRandomToken(15)}
	b := util.BufPool.Get()
	defer util.BufPool.Put(b)
	if err := encodeSession(b, s); err != nil {
		return nil, err
	}
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SELECT", r.dbId),
		radix.FlatCmd(nil, "SET", r.skey(s.Id), b.String()),
		radix.FlatCmd(nil, "SADD", r.ukey(s.User), s.Id),
	)
	if err := r.pool.Do(p); err != nil {
		return nil, err
	}
	return s, nil
}

func (r sessionMgrRedis) purgeAllData() error {
	s := radix.NewScanner(r.pool, radix.ScanOpts{Command: "SCAN", Pattern: r.prefix + "*"})
	var key string
	for s.Next(&key) {
		p := radix.Pipeline(
			radix.FlatCmd(nil, "SELECT", r.dbId),
			radix.FlatCmd(nil, "DEL", key),
		)
		if err := r.pool.Do(p); err != nil {
			return err
		}
	}
	return s.Close()
}

func (r sessionMgrRedis) getSession(id string) (*Session, error) {
	b := util.BufPool.Get()
	defer util.BufPool.Put(b)
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SELECT", r.dbId),
		radix.Cmd(b, "GET", r.skey(id)),
	)
	if err := r.pool.Do(p); err != nil {
		return nil, err
	}
	if b.Len() == 0 {
		return nil, util.NewErrorFrom(models.ErrDoesntExist)
	}
	s := &Session{}
	if err := decodeSession(b, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r sessionMgrRedis) storeSession(s *Session) error {
	b := util.BufPool.Get()
	defer util.BufPool.Put(b)
	s.LastAccess = time.Now()
	if err := encodeSession(b, s); err != nil {
		return err
	}
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SELECT", r.dbId),
		radix.FlatCmd(nil, "SET", r.skey(s.Id), b.String()),
		radix.FlatCmd(nil, "SADD", r.ukey(s.User), s.Id),
	)
	if err := r.pool.Do(p); err != nil {
		return err
	}
	return nil
}

func (r sessionMgrRedis) GetSession(id string) (*Session, error) {
	s, err := r.getSession(id)
	if err != nil {
		return nil, err
	}
	return s, err
}

func (r sessionMgrRedis) UpdateSession(id, agent string) (*Session, error) {
	s, err := r.getSession(id)
	if err != nil {
		return nil, err
	}
	s.Agent = agent
	s.LastAccess = time.Now().UTC()
	return s, r.storeSession(s)
}

func (r sessionMgrRedis) DeleteSession(id string) error {
	s, err := r.getSession(id)
	if err != nil {
		return err
	}
	return r.delete(s)
}

func (r sessionMgrRedis) delete(s *Session) error {
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SELECT", r.dbId),
		radix.FlatCmd(nil, "DEL", r.skey(s.Id), nil),
		radix.FlatCmd(nil, "SREM", r.ukey(s.User), s.Id),
	)
	if err := r.pool.Do(p); err != nil {
		return err
	}
	return nil
}

func (r sessionMgrRedis) DeleteAllSessions(userId string) error {
	sids := []string{}
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SELECT", r.dbId),
		radix.Cmd(&sids, "SMEMBERS", r.ukey(userId)),
	)
	if err := r.pool.Do(p); err != nil {
		return err
	}
	as := make([]radix.CmdAction, len(sids)+2)
	as[0] = radix.FlatCmd(nil, "SELECT", r.dbId)
	for i, sid := range sids {
		as[i+1] = radix.FlatCmd(nil, "DEL", r.skey(sid))
	}
	as[len(as)-1] = radix.FlatCmd(nil, "DEL", r.ukey(userId))
	return r.pool.Do(radix.Pipeline(as...))
}

func (r sessionMgrRedis) GetAllSessions(userId string) ([]*Session, error) {
	sids := []string{}
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SELECT", r.dbId),
		radix.Cmd(&sids, "SMEMBERS", r.ukey(userId)),
	)
	if err := r.pool.Do(p); err != nil {
		return nil, err
	}
	ses := make([]*Session, len(sids))
	b := util.BufPool.Get()
	defer util.BufPool.Put(b)
	for i, sid := range sids {
		b.Reset()
		p = radix.Pipeline(
			radix.FlatCmd(nil, "SELECT", r.dbId),
			radix.Cmd(b, "GET", r.skey(sid)),
		)
		if err := r.pool.Do(p); err != nil {
			return nil, err
		}
		ses[i] = &Session{}
		if err := decodeSession(b, ses[i]); err != nil {
			return nil, err
		}
	}
	return ses, nil
}
