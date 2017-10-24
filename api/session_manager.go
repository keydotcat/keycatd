package api

import "time"

type Session struct {
	Id           string    `json:"id"`
	UserId       string    `json:"user_id"`
	Agent        string    `json:"agent"`
	RequiresCSRF bool      `json:"-"`
	LastAccess   time.Time `json:"last_access"`
}

type SessionManager interface {
	NewSession(userId string, agent string, csrf bool) (Session, error)
	UpdateSession(id, agent string) error
	DeleteSession(id string) error
	GetAllSessions(userId string) ([]Session, error)
	DeleteAllSessions(userId string) error
}