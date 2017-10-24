package api

import (
	"encoding/json"
	"time"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
	radix "github.com/mediocregopher/radix.v3"
)

type redisSessionManager struct {
	pool *radix.Pool
}

func NewRedisSessionManager(connUrl string) (SessionManager, error) {
	pool, err := radix.NewPool("tcp", connUrl, 10, nil)
	if err != nil {
		return nil, err
	}
	return redisSessionManager{pool}, nil
}

func (r redisSessionManager) NewSession(userId, agent string, csrf bool) (Session, error) {
	s := Session{util.GenerateRandomToken(15), userId, agent, csrf, time.Now().UTC()}
	b := bufPool.Get()
	defer bufPool.Put(b)
	if err := json.NewEncoder(b).Encode(s); err != nil {
		return s, err
	}
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SET", "s:"+s.Id, b.String()),
		radix.FlatCmd(nil, "SADD", "u:"+s.UserId, s.Id),
	)
	if err := r.pool.Do(p); err != nil {
		return s, err
	}
	return s, nil
}

func (r redisSessionManager) getSession(id string) (*Session, error) {
	b := bufPool.Get()
	defer bufPool.Put(b)
	if err := r.pool.Do(radix.Cmd(b, "GET", "s:"+id)); err != nil {
		return nil, err
	}
	if b.Len() == 0 {
		return nil, util.NewErrorFrom(models.ErrDoesntExist)
	}
	s := &Session{}
	if err := json.NewDecoder(b).Decode(s); err != nil {
		return nil, err
	}
	return s, nil
}

func (r redisSessionManager) storeSession(s *Session) error {
	b := bufPool.Get()
	defer bufPool.Put(b)
	s.LastAccess = time.Now()
	if err := json.NewEncoder(b).Encode(s); err != nil {
		return err
	}
	p := radix.Pipeline(
		radix.FlatCmd(nil, "SET", "s:"+s.Id, b.String()),
		radix.FlatCmd(nil, "SADD", "u:"+s.UserId, s.Id),
	)
	if err := r.pool.Do(p); err != nil {
		return err
	}
	return nil
}

func (r redisSessionManager) UpdateSession(id, agent string) error {
	s, err := r.getSession(id)
	if err != nil {
		return err
	}
	s.Agent = agent
	return r.storeSession(s)
}

func (r redisSessionManager) DeleteSession(id string) error {
	s, err := r.getSession(id)
	if err != nil {
		return err
	}
	return r.delete(s)
}

func (r redisSessionManager) delete(s *Session) error {
	p := radix.Pipeline(
		radix.FlatCmd(nil, "DEL", "s:"+s.Id, nil),
		radix.FlatCmd(nil, "SREM", "u:"+s.UserId, s.Id),
	)
	if err := r.pool.Do(p); err != nil {
		return err
	}
	return nil
}

func (r redisSessionManager) DeleteAllSessions(userId string) error {
	sids := []string{}
	if err := r.pool.Do(radix.Cmd(&sids, "GET", "u:"+userId)); err != nil {
		return err
	}
	as := make([]radix.CmdAction, len(sids)+1)
	for i, sid := range sids {
		as[i] = radix.FlatCmd(nil, "DEL", "s:"+sid)
	}
	as[len(as)-1] = radix.FlatCmd(nil, "DEL", "u:"+userId)
	return r.pool.Do(radix.Pipeline(as...))
}

func (r redisSessionManager) GetAllSessions(userId string) ([]Session, error) {
	sids := []string{}
	if err := r.pool.Do(radix.Cmd(&sids, "SMEMBERS", "u:"+userId)); err != nil {
		return nil, err
	}
	ses := make([]Session, len(sids))
	b := bufPool.Get()
	defer bufPool.Put(b)
	for i, sid := range sids {
		b.Reset()
		if err := r.pool.Do(radix.Cmd(b, "GET", "s:"+sid)); err != nil {
			return nil, err
		}
		if err := json.NewDecoder(b).Decode(&(ses[i])); err != nil {
			return nil, err
		}
	}
	return ses, nil
}
