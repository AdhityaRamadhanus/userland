package session

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/stretchr/testify/mock"
)

type SessionService struct {
	mock.Mock
}

func (m SessionService) CreateSession(userID int, session userland.Session) error {
	args := m.Called(userID, session)

	return args.Get(0).(error)
}

func (m SessionService) ListSession(userID int) (userland.Sessions, error) {
	args := m.Called(userID)

	if args.Get(1) == nil {
		return args.Get(0).(userland.Sessions), nil
	}

	return nil, args.Get(1).(error)
}

func (m SessionService) EndSession(userID int, currentSessionID string) error {
	args := m.Called(userID, currentSessionID)

	return args.Get(0).(error)
}

func (m SessionService) EndOtherSessions(userID int, currentSessionID string) error {
	args := m.Called(userID, currentSessionID)

	return args.Get(0).(error)
}

func (m SessionService) CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error) {
	args := m.Called(user, currentSessionID)

	if args.Get(1) == nil {
		return args.Get(0).(security.AccessToken), nil
	}

	return security.AccessToken{}, args.Get(1).(error)
}

func (m SessionService) CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error) {
	args := m.Called(user, refreshTokenID)

	if args.Get(1) == nil {
		return args.Get(0).(security.AccessToken), nil
	}

	return security.AccessToken{}, args.Get(1).(error)
}

//Service provide an interface to story domain service
// type Service interface {
// 	CreateSession(userID int, session userland.Session) error
// 	ListSession(userID int) (userland.Sessions, error)
// 	EndSession(userID int, currentSessionID string) error
// 	EndOtherSessions(userID int, currentSessionID string) error
// 	CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error)
// 	CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error)
// }
