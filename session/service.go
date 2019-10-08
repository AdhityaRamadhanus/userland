package session

import (
	"encoding/json"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
)

type SessionDetail struct {
	IP         string
	ClientID   int
	ClientName string
}

//Service provide an interface to story domain service
type Service interface {
	CreateSession(userID int, currentSessionID string, sessionDetail SessionDetail) error
	ListSession(userID int, currentSessionID string) (userland.Sessions, error)
	// EndSession(user userland.User, currentSessionID string) error
	// EndOtherSessions(user userland.User, currentSessionID string) error
}

func NewService(keyValueService userland.KeyValueService) Service {
	return &service{
		keyValueService: keyValueService,
	}
}

type service struct {
	keyValueService userland.KeyValueService
}

func (s *service) CreateSession(userID int, currentSessionID string, sessionDetail SessionDetail) error {
	sessionTimestamp := time.Now().UnixNano()
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionMap := map[string]interface{}{
		"session_id":  currentSessionID,
		"ip":          sessionDetail.IP,
		"client_id":   sessionDetail.ClientID,
		"client_name": sessionDetail.ClientName,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
	sessionMapBytes, err := json.Marshal(sessionMap)
	if err != nil {
		return err
	}
	return s.keyValueService.AddToSortedSet(sessionListKey, string(sessionMapBytes), float64(sessionTimestamp))
}

func (s *service) ListSession(userID int, currentSessionID string) (userland.Sessions, error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.keyValueService.GetSortedSet(sessionListKey)
	if err != nil {
		return nil, err
	}

	sessions := userland.Sessions{}
	for _, sessionStr := range sessionsStr {
		sessionMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(sessionStr), &sessionMap); err != nil {
			continue
		}
		session := userland.Session{
			IsCurrent:  sessionMap["session_id"].(string) == currentSessionID,
			ClientID:   sessionMap["client_id"].(int),
			ClientName: sessionMap["client_name"].(string),
			CreatedAt:  sessionMap["created_at"].(time.Time),
			UpdatedAt:  sessionMap["updated_at"].(time.Time),
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}
