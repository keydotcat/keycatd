package managers

import (
	"database/sql"
	"time"

	"github.com/keydotcat/backend/models"
	"github.com/keydotcat/backend/util"
)

type sessionMgrPSQL struct {
	dbp *sql.DB
}

func NewSessionMgrPSQL(dbp *sql.DB) SessionMgr {
	return sessionMgrPSQL{dbp}
}

func (r sessionMgrPSQL) doTx(ftor func(*sql.Tx) error) error {
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

func (r sessionMgrPSQL) NewSession(userId, agent string, csrf bool) (*Session, error) {
	o := Session{util.GenerateRandomToken(15), userId, agent, csrf, time.Now().UTC(), util.GenerateRandomToken(15)}
	_, err := r.dbp.Exec("INSERT INTO \"session\" "+insertSessionFields+" VALUES "+insertSessionBinds, o.Id, o.UserId, o.Agent, o.RequiresCSRF, o.LastAccess, o.StoreToken)
	if err == nil {
		return &o, nil
	}
	if models.IsDuplicateErr(err) {
		return r.NewSession(userId, agent, csrf)
	}
	panic(err)
}

func (r sessionMgrPSQL) GetSession(id string) (*Session, error) {
	o := &Session{}
	row := r.dbp.QueryRow("SELECT "+selectSessionFields+" FROM \"session\" WHERE "+findSessionCondition, id)
	return o, o.dbScanRow(row)
}

func (r sessionMgrPSQL) UpdateSession(id, agent string) (*Session, error) {
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
		_, err := o.dbUpdate(tx)
		return err
	})
}

func (r sessionMgrPSQL) DeleteSession(id string) error {
	_, err := r.dbp.Exec("DELETE FROM \"session\" WHERE "+findSessionCondition, id)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r sessionMgrPSQL) DeleteAllSessions(userId string) error {
	_, err := r.dbp.Exec("DELETE FROM \"session\" WHERE \"user_id\"=$1", userId)
	if err != nil {
		panic(err)
	}
	return nil
}

func (r sessionMgrPSQL) GetAllSessions(userId string) ([]*Session, error) {
	rows, err := r.dbp.Query("SELECT "+selectSessionFields+" FROM \"session\" WHERE \"user_id\"=$1", userId)
	if err != nil {
		panic(err)
	}
	return scanSessions(rows)
}

func (r sessionMgrPSQL) purgeAllData() {
	_, err := r.dbp.Exec("DELETE FROM \"session\"")
	if err != nil {
		panic(err)
	}
}
