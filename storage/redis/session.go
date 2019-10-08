package redis

import (
	"encoding/json"
	"math"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"

	"github.com/go-redis/redis"
)

//SessionRepository implements userland.SessionRepository interface using redis
type SessionRepository struct {
	redisClient *redis.Client
}

//NewSessionRepository construct a new SessionRepository from redis client
func NewSessionRepository(redisClient *redis.Client) *SessionRepository {
	return &SessionRepository{
		redisClient: redisClient,
	}
}

//Get a cache in bytes from a key
func (s SessionRepository) Create(userID int, session userland.Session) (err error) {
	sessionTimestamp := time.Now().Unix() + int64(security.UserAccessTokenExpiration.Seconds())
	sessionMap := map[string]interface{}{
		"session_id":  session.ID,
		"ip":          session.IP,
		"client_id":   session.ClientID,
		"client_name": session.ClientName,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	sessionMapBytes, err := json.Marshal(sessionMap)
	if err != nil {
		return err
	}

	sessionListKey := keygenerator.SessionListKey(userID)
	return s.redisClient.ZAdd(sessionListKey, redis.Z{Score: float64(sessionTimestamp), Member: string(sessionMapBytes)}).Err()
}

func (s SessionRepository) FindAllByUserID(userID int) (userland.Sessions, error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.redisClient.ZRange(sessionListKey, math.MinInt64, math.MaxInt64).Result()
	if err != nil {
		return nil, err
	}

	sessions := userland.Sessions{}
	for _, sessionStr := range sessionsStr {
		sessionMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(sessionStr), &sessionMap); err != nil {
			continue
		}
		createdAt, _ := time.Parse(time.RFC3339, sessionMap["created_at"].(string))
		updatedAt, _ := time.Parse(time.RFC3339, sessionMap["updated_at"].(string))
		session := userland.Session{
			ID:         sessionMap["session_id"].(string),
			IP:         sessionMap["ip"].(string),
			ClientID:   int(sessionMap["client_id"].(float64)),
			ClientName: sessionMap["client_name"].(string),
			CreatedAt:  createdAt,
			UpdatedAt:  updatedAt,
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (s SessionRepository) DeleteBySessionID(userID int, sessionID string) (err error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.redisClient.ZRange(sessionListKey, math.MinInt64, math.MaxInt64).Result()
	if err != nil {
		return err
	}

	var toBedeletedSession string
	for _, sessionStr := range sessionsStr {
		sessionMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(sessionStr), &sessionMap); err != nil {
			return err
		}
		sessionID := sessionMap["session_id"].(string)
		if sessionID == sessionID {
			toBedeletedSession = sessionStr
			break
		}
	}
	if toBedeletedSession == "" {
		return userland.ErrSessionNotFound
	}

	return s.redisClient.ZRem(sessionListKey, toBedeletedSession).Err()
}

func (s SessionRepository) DeleteOtherSessions(userID int, currentSessionID string) (deletedSessionIDs []string, err error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.redisClient.ZRange(sessionListKey, math.MinInt64, math.MaxInt64).Result()
	if err != nil {
		return nil, err
	}

	deletedSessionIDs = []string{}
	for _, sessionStr := range sessionsStr {
		sessionMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(sessionStr), &sessionMap); err != nil {
			continue
		}
		sessionID := sessionMap["session_id"].(string)
		if sessionID == currentSessionID {
			continue
		}

		if err = s.redisClient.ZRem(sessionListKey, sessionStr).Err(); err != nil {
			continue
		}
		deletedSessionIDs = append(deletedSessionIDs, sessionID)
	}

	return deletedSessionIDs, nil
}
