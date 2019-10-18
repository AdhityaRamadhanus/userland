package userland

import (
	"github.com/go-errors/errors"

	"time"
)

var (
	//ErrSessionNotFound represent session not found
	ErrSessionNotFound = errors.New("Session not found")
)

//Session is domain entity
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

//Sessions a collection of Session
type Sessions []Session

//SessionRepository provide an interface to get user sessions
type SessionRepository interface {
	Create(userID int, session Session) error
	FindAllByUserID(userID int) (Sessions, error)
	DeleteExpiredSessions(userID int) error
	DeleteBySessionID(userID int, sessionID string) error
	DeleteOtherSessions(userID int, sessionID string) (deletedSessionIDs []string, err error)
}
