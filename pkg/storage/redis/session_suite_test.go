// +build integration

package redis_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	_redis "github.com/go-redis/redis"
	"github.com/stretchr/testify/suite"
)

type SessionRepositoryTestSuite struct {
	suite.Suite
	Config            *config.Configuration
	RedisClient       *_redis.Client
	SessionRepository userland.SessionRepository
}

func NewSessionRepositoryTestSuite(cfg *config.Configuration) *SessionRepositoryTestSuite {
	return &SessionRepositoryTestSuite{
		Config: cfg,
	}
}

func (suite *SessionRepositoryTestSuite) Teardown() {
	suite.T().Log("Teardown SessionRepositoryTestSuite")
	suite.RedisClient.Close()
}

func (suite *SessionRepositoryTestSuite) SetupTest() {
	if err := suite.RedisClient.FlushAll().Err(); err != nil {
		suite.T().Fatalf("RedisClient.FlushAll() err = %v; want nil", err)
	}
}

func (suite *SessionRepositoryTestSuite) SetupSuite() {
	suite.T().Logf("Connecting to redis at %v", suite.Config.Redis)
	redisClient, err := redis.CreateClient(suite.Config.Redis, 0)
	if err != nil {
		suite.T().Fatalf("redis.CreateClient() err = %v; want nil", err)
	}
	suite.RedisClient = redisClient
	suite.SessionRepository = redis.NewSessionRepository(redisClient)
}

func (suite *SessionRepositoryTestSuite) TestCreate() {
	type args struct {
		userID  int
		session userland.Session
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				userID: 1,
				session: userland.Session{
					ID:         security.GenerateUUID(),
					Token:      "test",
					IP:         "123.123.13.123",
					ClientID:   1,
					ClientName: "test",
					Expiration: security.UserAccessTokenExpiration,
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.SessionRepository.Create(tc.args.userID, tc.args.session)
			if err != nil && !tc.wantErr {
				t.Fatalf("SessionRepository.Create() err = %v; want nil", err)
			}

			if err == nil && tc.wantErr {
				t.Fatal("SessionRepository.Create() err = nil; want not nil")
			}
		})
	}
}

func (suite *SessionRepositoryTestSuite) TestFindAllByUserID() {
	type args struct {
		userID             int
		createSessionCount int
	}
	testCases := []struct {
		name      string
		args      args
		wantCount int
	}{
		{
			name: "return 1",
			args: args{
				userID:             1,
				createSessionCount: 2,
			},
			wantCount: 2,
		},
		{
			name: "return 0",
			args: args{
				userID:             2,
				createSessionCount: 0,
			},
			wantCount: 0,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// setup
			for i := 0; i < tc.args.createSessionCount; i++ {
				suite.SessionRepository.Create(tc.args.userID, userland.Session{
					ID:         security.GenerateUUID(),
					Token:      "test",
					IP:         "123.123.13.123",
					ClientID:   1,
					ClientName: "test",
					Expiration: security.UserAccessTokenExpiration,
				})
			}
			sessions, err := suite.SessionRepository.FindAllByUserID(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionRepository.FindAllByUserID(%d) err = %v; want nil", tc.args.userID, err)
			}

			gotCount := len(sessions)
			if gotCount != tc.wantCount {
				t.Errorf("SessionRepository.FindAllByUserID(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, tc.wantCount)
			}
		})
	}
}

func (suite *SessionRepositoryTestSuite) TestDeleteExpiredSessions() {
	type args struct {
		userID                    int
		createSessionCount        int
		createExpiredSessionCount int
	}
	testCases := []struct {
		name      string
		args      args
		wantCount int
	}{
		{
			name: "create 2, expired 1, should return 1",
			args: args{
				userID:                    1,
				createSessionCount:        2,
				createExpiredSessionCount: 1,
			},
			wantCount: 1,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// setup
			sessionExpiration := 100 * time.Millisecond
			expiredSessionCount := tc.args.createExpiredSessionCount
			for i := 0; i < tc.args.createSessionCount; i++ {
				exp := sessionExpiration
				if expiredSessionCount <= 0 {
					exp += 100 * time.Second
				}
				suite.SessionRepository.Create(tc.args.userID, userland.Session{
					ID:         security.GenerateUUID(),
					Token:      "test",
					IP:         "123.123.13.123",
					ClientID:   1,
					ClientName: "test",
					Expiration: exp,
				})
				expiredSessionCount--
			}

			time.Sleep(2 * sessionExpiration)
			if err := suite.SessionRepository.DeleteExpiredSessions(tc.args.userID); err != nil {
				t.Fatalf("SessionRepository.DeleteExpiredSessions(%d) err = %v; want nil", tc.args.userID, err)
			}
			sessions, err := suite.SessionRepository.FindAllByUserID(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionRepository.FindAllByUserID(%d) err = %v; want nil", tc.args.userID, err)
			}

			gotCount := len(sessions)
			if gotCount != tc.wantCount {
				t.Errorf("SessionRepository.FindAllByUserID(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, tc.wantCount)
			}
		})
	}
}

func (suite *SessionRepositoryTestSuite) TestDeleteBySessionID() {
	type args struct {
		userID    int
		sessionID string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				userID:    1,
				sessionID: security.GenerateUUID(),
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			suite.SessionRepository.Create(tc.args.userID, userland.Session{
				ID:         tc.args.sessionID,
				Token:      "test",
				IP:         "123.123.13.123",
				ClientID:   1,
				ClientName: "test",
				Expiration: security.UserAccessTokenExpiration,
			})

			err := suite.SessionRepository.DeleteBySessionID(tc.args.userID, tc.args.sessionID)
			if err != nil {
				t.Fatalf("SessionRepository.DeleteBySessionID(%d, %s) err = %v; want nil", tc.args.userID, tc.args.sessionID, err)
			}
			sessions, err := suite.SessionRepository.FindAllByUserID(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionRepository.FindAllByUserID(%d) err = %v; want nil", tc.args.userID, err)
			}

			gotCount := len(sessions)
			wantCount := 0
			if gotCount != wantCount {
				t.Errorf("SessionRepository.FindAllByUserID(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, wantCount)
			}
		})
	}
}

func (suite *SessionRepositoryTestSuite) TestDeleteOtherSessions() {
	type args struct {
		userID     int
		sessionIDs []string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				userID: 1,
				sessionIDs: []string{
					security.GenerateUUID(),
					security.GenerateUUID(),
					security.GenerateUUID(),
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			for _, sessionID := range tc.args.sessionIDs {
				suite.SessionRepository.Create(tc.args.userID, userland.Session{
					ID:         sessionID,
					Token:      "test",
					IP:         "123.123.13.123",
					ClientID:   1,
					ClientName: "test",
					Expiration: security.UserAccessTokenExpiration,
				})
			}

			// just pick the first session ID
			keepSessionID := tc.args.sessionIDs[0]
			_, err := suite.SessionRepository.DeleteOtherSessions(tc.args.userID, keepSessionID)
			if err != nil {
				t.Fatalf("SessionRepository.DeleteBySessionID(%d, %s) err = %v; want nil", tc.args.userID, keepSessionID, err)
			}
			sessions, err := suite.SessionRepository.FindAllByUserID(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionRepository.FindAllByUserID(%d) err = %v; want nil", tc.args.userID, err)
			}

			gotCount := len(sessions)
			wantCount := 1
			if gotCount != wantCount {
				t.Errorf("SessionRepository.FindAllByUserID(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, wantCount)
			}
		})
	}
}
