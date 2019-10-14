// +build all session_service

package session_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/metrics"

	_redis "github.com/go-redis/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/service/session"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

// keyValueService   userland.KeyValueService
// sessionRepository userland.SessionRepository
// userRepository    userland.UserRepository

type SessionServiceTestSuite struct {
	suite.Suite
	RedisClient       *_redis.Client
	KeyValueService   *redis.KeyValueService
	SessionRepository *redis.SessionRepository
	SessionService    session.Service
}

func (suite *SessionServiceTestSuite) SetupTest() {
	if err := suite.RedisClient.FlushAll().Err(); err != nil {
		log.Fatal("Cannot setup redis")
	}
}

// before each test
func (suite *SessionServiceTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")

	redisClient := _redis.NewClient(&_redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("TEST_REDIS_HOST"), os.Getenv("TEST_REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		log.WithError(err).Error("Failed to connect to redis")
	}

	keyValueService := redis.NewKeyValueService(redisClient)
	sessionRepository := redis.NewSessionRepository(redisClient)

	sessionService := session.NewService(
		session.WithSessionRepository(sessionRepository),
		session.WithKeyValueService(keyValueService),
	)
	sessionService = session.NewInstrumentorService(
		metrics.PrometheusRequestCounter("service", "session", session.MetricKeys),
		metrics.PrometheusRequestLatency("service", "session", session.MetricKeys),
		sessionService,
	)

	suite.RedisClient = redisClient
	suite.KeyValueService = keyValueService
	suite.SessionService = sessionService
}

func TestSessionService(t *testing.T) {
	suiteTest := new(SessionServiceTestSuite)
	suite.Run(t, suiteTest)
}

// CreateSession(userID int, session userland.Session) error
// ListSession(userID int) (userland.Sessions, error)
// EndSession(userID int, currentSessionID string) error
// EndOtherSessions(userID int, currentSessionID string) error
// CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error)
// CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error)

func (suite *SessionServiceTestSuite) TestCreateSession() {
	testCases := []struct {
		UserID  int
		Session userland.Session
	}{
		{
			UserID: 1,
			Session: userland.Session{
				ID:         security.GenerateUUID(),
				Token:      "test",
				IP:         "123.123.13.123",
				ClientID:   1,
				ClientName: "test",
				Expiration: security.UserAccessTokenExpiration,
			},
		},
	}

	for _, testCase := range testCases {
		err := suite.SessionService.CreateSession(testCase.UserID, testCase.Session)
		suite.Nil(err, "should create session")
	}
}

func (suite *SessionServiceTestSuite) TestListSession() {
	testCases := []struct {
		UserID int
	}{
		{
			UserID: 1,
		},
	}

	for _, testCase := range testCases {
		err := suite.SessionService.CreateSession(testCase.UserID, userland.Session{
			ID:         security.GenerateUUID(),
			Token:      "test",
			IP:         "123.123.13.123",
			ClientID:   1,
			ClientName: "test",
			Expiration: security.UserAccessTokenExpiration,
		})
		suite.Nil(err, "should create session")

		sessions, err := suite.SessionService.ListSession(testCase.UserID)
		suite.Nil(err, "should list session")
		suite.Equal(len(sessions), 1, "should return 1 sessions")
	}
}

func (suite *SessionServiceTestSuite) TestEndSession() {
	testCases := []struct {
		UserID int
	}{
		{
			UserID: 1,
		},
	}

	for _, testCase := range testCases {
		sessionID := security.GenerateUUID()
		err := suite.SessionService.CreateSession(testCase.UserID, userland.Session{
			ID:         sessionID,
			Token:      "test",
			IP:         "123.123.13.123",
			ClientID:   1,
			ClientName: "test",
			Expiration: security.UserAccessTokenExpiration,
		})
		suite.Nil(err, "should create session")

		err = suite.SessionService.EndSession(testCase.UserID, sessionID)
		suite.Nil(err, "should delete session")
	}
}

func (suite *SessionServiceTestSuite) TestOtherSessions() {
	testCases := []struct {
		UserID     int
		SessionIDs []string
	}{
		{
			UserID: 1,
			SessionIDs: []string{
				security.GenerateUUID(),
				security.GenerateUUID(),
				security.GenerateUUID(),
			},
		},
	}

	for _, testCase := range testCases {
		for _, sessionID := range testCase.SessionIDs {
			suite.SessionService.CreateSession(testCase.UserID, userland.Session{
				ID:         sessionID,
				Token:      "test",
				IP:         "123.123.13.123",
				ClientID:   1,
				ClientName: "test",
				Expiration: security.UserAccessTokenExpiration,
			})
		}

		err := suite.SessionService.EndOtherSessions(testCase.UserID, testCase.SessionIDs[0])
		suite.Nil(err, "should delete sessions")
	}
}

func (suite *SessionServiceTestSuite) TestCreateRefreshToken() {
	testCases := []struct {
		User      userland.User
		SessionID string
	}{
		{
			User: userland.User{
				ID:       1,
				Fullname: "adhitya",
				Email:    "test@coba.com",
			},
			SessionID: security.GenerateUUID(),
		},
	}

	for _, testCase := range testCases {
		_, err := suite.SessionService.CreateRefreshToken(testCase.User, testCase.SessionID)
		suite.Nil(err, "should create refresh token")
	}
}

func (suite *SessionServiceTestSuite) TestCreateNewAccessToken() {
	testCases := []struct {
		User userland.User
	}{
		{
			User: userland.User{
				ID:       1,
				Fullname: "adhitya",
				Email:    "test@coba.com",
			},
		},
	}

	for _, testCase := range testCases {
		sessionID := security.GenerateUUID()
		refreshToken, err := suite.SessionService.CreateRefreshToken(testCase.User, sessionID)
		suite.Nil(err, "should create refresh token")

		_, err = suite.SessionService.CreateNewAccessToken(testCase.User, refreshToken.Key)
		suite.Nil(err, "should create new access token")
	}
}
