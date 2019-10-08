package userland

import (
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("Session not found")
)

type Session struct {
	ID         string
	Token      string
	IP         string
	ClientID   int
	ClientName string
	Expiration time.Duration
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Sessions []Session

type SessionRepository interface {
	Create(userID int, session Session) error
	FindAllByUserID(userID int) (Sessions, error)
	DeleteBySessionID(userID int, sessionID string) error
	DeleteOtherSessions(userID int, sessionID string) (deletedSessionIDs []string, err error)
}
