package session

import (
	"github.com/go-errors/errors"


	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
)

var (
	ErrSessionNotFound = errors.New("Session Not Found")
)

//Service provide an interface to story domain service
type Service interface {
	CreateSession(userID int, session userland.Session) error
	ListSession(userID int) (userland.Sessions, error)
	EndSession(userID int, currentSessionID string) error
	EndOtherSessions(userID int, currentSessionID string) error
	CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error)
	CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error)
}

func NewService(keyValueService userland.KeyValueService, sessionRepository userland.SessionRepository) Service {
	return &service{
		keyValueService:   keyValueService,
		sessionRepository: sessionRepository,
	}
}

type service struct {
	keyValueService   userland.KeyValueService
	sessionRepository userland.SessionRepository
	userRepository    userland.UserRepository
}

func (s *service) CreateSession(userID int, session userland.Session) error {
	tokenKey := keygenerator.TokenKey(session.ID)
	if err := s.keyValueService.SetEx(tokenKey, []byte(session.Token), security.UserAccessTokenExpiration); err != nil {
		return err
	}
	return s.sessionRepository.Create(userID, session)
}

func (s *service) ListSession(userID int) (userland.Sessions, error) {
	// remove expired sessions
	return s.sessionRepository.FindAllByUserID(userID)
}

func (s *service) EndSession(userID int, currentSessionID string) error {
	if err := s.sessionRepository.DeleteBySessionID(userID, currentSessionID); err != nil {
		return err
	}

	tokenKey := keygenerator.TokenKey(currentSessionID)
	return s.keyValueService.Delete(tokenKey)
}

func (s *service) EndOtherSessions(userID int, currentSessionID string) error {
	deletedSessionIDs, err := s.sessionRepository.DeleteOtherSessions(userID, currentSessionID)
	if err != nil {
		return err
	}

	for _, deletedSessionID := range deletedSessionIDs {
		tokenKey := keygenerator.TokenKey(deletedSessionID)
		s.keyValueService.Delete(tokenKey)
	}
	return nil
}

func (s *service) CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error) {
	refreshToken, err := security.CreateAccessToken(user, security.AccessTokenOptions{
		Scope:      security.RefreshTokenScope,
		Expiration: security.RefreshAccessTokenExpiration,
		CustomClaim: map[string]interface{}{
			"previous_session_id": currentSessionID,
		},
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	tokenKey := keygenerator.TokenKey(refreshToken.Key)
	s.keyValueService.SetEx(tokenKey, []byte(refreshToken.Value), security.RefreshAccessTokenExpiration)

	return refreshToken, nil
}

func (s *service) CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error) {
	newAccessToken, err := security.CreateAccessToken(user, security.AccessTokenOptions{
		Scope:      security.UserTokenScope,
		Expiration: security.UserAccessTokenExpiration,
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	tokenKey := keygenerator.TokenKey(refreshTokenID)
	s.keyValueService.Delete(tokenKey)
	return newAccessToken, nil
}
