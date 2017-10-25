package managers

import "time"

type Session struct {
	Id           string    `json:"id"`
	UserId       string    `json:"user_id"`
	Agent        string    `json:"agent"`
	RequiresCSRF bool      `json:"csrf"`
	LastAccess   time.Time `json:"last_access"`
}

type SessionMgr interface {
	NewSession(userId string, agent string, csrf bool) (Session, error)
	UpdateSession(id, agent string) (Session, error)
	DeleteSession(id string) error
	GetAllSessions(userId string) ([]Session, error)
	DeleteAllSessions(userId string) error
}
