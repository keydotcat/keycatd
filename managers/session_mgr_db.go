package managers

import (
	"database/sql"
	"time"

	"github.com/keydotcat/keycatd/models"
	"github.com/keydotcat/keycatd/util"
)

type sessionMgrDB struct {
	dbp *sql.DB
}

func NewSessionMgrDB(dbp *sql.DB) SessionMgr {
	return sessionMgrDB{dbp}
}

func (r sessionMgrDB) doTx(ftor func(*sql.Tx) error) error {
	tx, err := r.dbp.Begin()
	if err != nil {
		panic(err)
	}
	if err = ftor(tx); err != nil {
		if util.CheckErr(err, sql.ErrTxDone) || util.CheckErr(err, sql.ErrConnDone) {
			return err
		}
		if rerr := tx.Rollback(); rerr != nil {
			return util.NewErrorf("Could not rollback transaction: %s (prev error was %s)", rerr, err)
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return util.NewErrorf("Could not commit transaction: %s", err)
	}
	return nil
}

func (r sessionMgrDB) NewSession(userId, ip, agent string, csrf bool) (*Session, error) {
	o := Session{util.GenerateRandomToken(15), userId, agent, csrf, time.Now().UTC(), util.GenerateRandomToken(15), ip}
	err := r.doTx(func(tx *sql.Tx) error {
		_, err := r.dbp.Exec("INSERT INTO \"session\" "+insertSessionFields+" VALUES "+insertSessionBinds, o.Id, o.User, o.Agent, o.RequiresCSRF, o.LastAccess, o.StoreToken, o.LastIp)
		return err
	})
	if err == nil {
		return &o, nil
	}
	if models.IsDuplicateErr(err) {
		return r.NewSession(userId, ip, agent, csrf)
	}
	panic(err)
}

func (r sessionMgrDB) GetSession(id string) (*Session, error) {
	o := &Session{}
	row := r.dbp.QueryRow("SELECT "+selectSessionFields+" FROM \"session\" WHERE "+findSessionCondition, id)
	return o, o.dbScanRow(row)
}

func (r sessionMgrDB) UpdateSession(id, ip, agent string) (*Session, error) {
	o := &Session{Id: id}
	return o, r.doTx(func(tx *sql.Tx) error {
		if err := o.dbFind(tx); err != nil {
			if util.CheckErr(err, sql.ErrNoRows) {
				return util.NewErrorFrom(models.ErrDoesntExist)
			}
			return err
		}
		o.Agent = agent
		o.LastAccess = time.Now().UTC()
		o.LastIp = ip
		_, err := o.dbUpdate(tx)
		return err
	})
}

func (r sessionMgrDB) DeleteSession(id string) error {
	_, err := r.dbp.Exec("DELETE FROM \"session\" WHERE "+findSessionCondition, id)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r sessionMgrDB) DeleteAllSessions(userId string) error {
	_, err := r.dbp.Exec("DELETE FROM \"session\" WHERE \"user\"=$1", userId)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r sessionMgrDB) GetAllSessions(userId string) ([]*Session, error) {
	rows, err := r.dbp.Query("SELECT "+selectSessionFields+" FROM \"session\" WHERE \"user\"=$1", userId)
	if err != nil {
		panic(err)
	}
	return scanSessions(rows)
}

func (r sessionMgrDB) purgeAllData() {
	_, err := r.dbp.Exec("DELETE FROM \"session\"")
	if err != nil {
		panic(err)
	}
}
