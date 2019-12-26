package session

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
)

type SimpleSessionService struct {
	CalledMethods map[string]bool
}

func (m SimpleSessionService) CreateSession(userID int, session userland.Session) error {
	m.CalledMethods["CreateSession"] = true

	return nil
}

func (m SimpleSessionService) ListSession(userID int) (userland.Sessions, error) {
	m.CalledMethods["ListSession"] = true

	return userland.Sessions{}, nil
}

func (m SimpleSessionService) EndSession(userID int, currentSessionID string) error {
	m.CalledMethods["EndSession"] = true

	return nil
}

func (m SimpleSessionService) EndOtherSessions(userID int, currentSessionID string) error {
	m.CalledMethods["EndOtherSessions"] = true

	return nil
}

func (m SimpleSessionService) CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error) {
	m.CalledMethods["CreateRefreshToken"] = true

	return security.AccessToken{}, nil
}

func (m SimpleSessionService) CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error) {
	m.CalledMethods["CreateNewAccessToken"] = true

	return security.AccessToken{}, nil
}
