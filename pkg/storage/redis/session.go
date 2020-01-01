package redis

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/pkg/errors"

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

func (s SessionRepository) Create(userID int, session userland.Session) (err error) {
	sessionTimestamp := time.Now().Unix() + int64(session.Expiration.Seconds())
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
		return errors.Wrap(err, "json.Marshal() err")
	}

	sessionListKey := keygenerator.SessionListKey(userID)
	if err := s.redisClient.ZAdd(sessionListKey, redis.Z{Score: float64(sessionTimestamp), Member: string(sessionMapBytes)}).Err(); err != nil {
		return errors.Wrapf(err, "redisClient.ZAdd(%q, redisZ) err", sessionListKey)
	}

	return nil
}

func (s SessionRepository) FindAllByUserID(userID int) (userland.Sessions, error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.redisClient.ZRange(sessionListKey, math.MinInt64, math.MaxInt64).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "redisClient.ZRange(%q) err", sessionListKey)
	}

	sessions := userland.Sessions{}
	for _, sessionStr := range sessionsStr {
		sessionMap := map[string]interface{}{}
		if err := json.Unmarshal([]byte(sessionStr), &sessionMap); err != nil {
			continue
		}
		createdAt, err := time.Parse(time.RFC3339, sessionMap["created_at"].(string))
		if err != nil {
			continue
		}
		updatedAt, err := time.Parse(time.RFC3339, sessionMap["updated_at"].(string))
		if err != nil {
			continue
		}
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

func (s SessionRepository) DeleteExpiredSessions(userID int) (err error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	nowEpochStr := strconv.FormatInt(time.Now().Unix(), 10)
	if err := s.redisClient.ZRemRangeByScore(sessionListKey, "-inf", nowEpochStr).Err(); err != nil {
		return errors.Wrapf(err, "redisClient.ZRemRangeByScore(%q, -inf, %s) err", sessionListKey, nowEpochStr)
	}

	return nil
}

func (s SessionRepository) DeleteBySessionID(userID int, sessionID string) (err error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.redisClient.ZRange(sessionListKey, math.MinInt64, math.MaxInt64).Result()
	if err != nil {
		return errors.Wrapf(err, "redisClient.ZRange(%q) err", sessionListKey)
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

	if err := s.redisClient.ZRem(sessionListKey, toBedeletedSession).Err(); err != nil {
		return errors.Wrapf(err, "redisClient.ZRem(%q, %q) err", sessionListKey, toBedeletedSession)
	}

	return nil
}

func (s SessionRepository) DeleteOtherSessions(userID int, currentSessionID string) (deletedSessionIDs []string, err error) {
	sessionListKey := keygenerator.SessionListKey(userID)
	sessionsStr, err := s.redisClient.ZRange(sessionListKey, math.MinInt64, math.MaxInt64).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "redisClient.ZRange(%q) err", sessionListKey)
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
