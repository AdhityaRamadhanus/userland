// +build all session redis_repository

package redis_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	"github.com/sarulabs/di"
	"github.com/stretchr/testify/suite"

	_redis "github.com/go-redis/redis"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

type SessionRepositoryTestSuite struct {
	suite.Suite
	RedisClient       *_redis.Client
	SessionRepository userland.SessionRepository
}

func (suite *SessionRepositoryTestSuite) SetupTest() {
	err := suite.RedisClient.FlushAll().Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}
}

func (suite *SessionRepositoryTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		redis.ConnectionBuilder("redis-connection", 0),
		redis.SessionRepositoryBuilder,
	)

	return builder.Build()
}

func (suite *SessionRepositoryTestSuite) SetupSuite() {
	godotenv.Load("../../../.env")
	ctn := suite.BuildContainer()
	suite.RedisClient = ctn.Get("redis-connection").(*_redis.Client)
	suite.SessionRepository = ctn.Get("session-repository").(userland.SessionRepository)
}

func TestSessionRepository(t *testing.T) {
	suiteTest := new(SessionRepositoryTestSuite)
	suite.Run(t, suiteTest)
}

// Create(userID int, session Session) error
// 	FindAllByUserID(userID int) (Sessions, error)
// 	DeleteExpiredSessions(userID int) error
// 	DeleteBySessionID(userID int, sessionID string) error
// 	DeleteOtherSessions(userID int, sessionID string) (deletedSessionIDs []string, err error)

func (suite *SessionRepositoryTestSuite) TestCreate() {
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
		err := suite.SessionRepository.Create(testCase.UserID, testCase.Session)
		suite.Nil(err, "should create session")
	}
}

func (suite *SessionRepositoryTestSuite) TestFindAllByUserID() {
	testCases := []struct {
		UserID        int
		SessionsCount int
	}{
		{
			UserID:        1,
			SessionsCount: 1,
		},
		{
			UserID:        2,
			SessionsCount: 0,
		},
	}

	for _, testCase := range testCases {
		// setup
		for i := 0; i < testCase.SessionsCount; i++ {
			suite.SessionRepository.Create(testCase.UserID, userland.Session{
				ID:         security.GenerateUUID(),
				Token:      "test",
				IP:         "123.123.13.123",
				ClientID:   1,
				ClientName: "test",
				Expiration: security.UserAccessTokenExpiration,
			})
		}
		sessions, err := suite.SessionRepository.FindAllByUserID(testCase.UserID)
		suite.Nil(err, "should create session")
		suite.Equal(len(sessions), testCase.SessionsCount, "should return session")
	}
}

func (suite *SessionRepositoryTestSuite) TestDeleteExpiredSessions() {
	testCases := []struct {
		UserID        int
		SessionsCount int
	}{
		{
			UserID:        1,
			SessionsCount: 1,
		},
	}

	for _, testCase := range testCases {
		// setup
		sessionExpiration := time.Second * 1
		for i := 0; i < testCase.SessionsCount; i++ {
			suite.SessionRepository.Create(testCase.UserID, userland.Session{
				ID:         security.GenerateUUID(),
				Token:      "test",
				IP:         "123.123.13.123",
				ClientID:   1,
				ClientName: "test",
				Expiration: sessionExpiration,
			})
		}

		time.Sleep(sessionExpiration + time.Second)
		err := suite.SessionRepository.DeleteExpiredSessions(testCase.UserID)
		suite.Nil(err, "should delete expired session")
		sessions, err := suite.SessionRepository.FindAllByUserID(testCase.UserID)
		suite.Nil(err, "should find no session")
		suite.Equal(len(sessions), 0, "should return no session")
	}
}

func (suite *SessionRepositoryTestSuite) TestDeleteBySessionID() {
	testCases := []struct {
		UserID    int
		SessionID string
	}{
		{
			UserID:    1,
			SessionID: security.GenerateUUID(),
		},
	}

	for _, testCase := range testCases {
		// setup
		suite.SessionRepository.Create(testCase.UserID, userland.Session{
			ID:         testCase.SessionID,
			Token:      "test",
			IP:         "123.123.13.123",
			ClientID:   1,
			ClientName: "test",
			Expiration: security.UserAccessTokenExpiration,
		})

		err := suite.SessionRepository.DeleteBySessionID(testCase.UserID, testCase.SessionID)
		suite.Nil(err, "should delete session")
		sessions, err := suite.SessionRepository.FindAllByUserID(testCase.UserID)
		suite.Nil(err, "should find no session")
		suite.Equal(len(sessions), 0, "should return no session")
	}
}

func (suite *SessionRepositoryTestSuite) TestDeleteOtherSessions() {
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
		// setup
		for _, sessionID := range testCase.SessionIDs {
			suite.SessionRepository.Create(testCase.UserID, userland.Session{
				ID:         sessionID,
				Token:      "test",
				IP:         "123.123.13.123",
				ClientID:   1,
				ClientName: "test",
				Expiration: security.UserAccessTokenExpiration,
			})
		}

		_, err := suite.SessionRepository.DeleteOtherSessions(testCase.UserID, testCase.SessionIDs[0])
		suite.Nil(err, "should delete session")
		sessions, err := suite.SessionRepository.FindAllByUserID(testCase.UserID)
		suite.Nil(err, "should find no session")
		suite.Equal(len(sessions), 1, "should return no session")
	}
}
