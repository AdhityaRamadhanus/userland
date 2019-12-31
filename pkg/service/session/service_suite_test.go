// +build integration

package session_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/service/authentication"
	"github.com/AdhityaRamadhanus/userland/pkg/service/session"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	"github.com/AdhityaRamadhanus/userland/pkg/userlandtest"
	_redis "github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type SessionServiceTestSuite struct {
	suite.Suite
	Config            *config.Configuration
	RedisClient       *_redis.Client
	KeyValueService   userland.KeyValueService
	SessionRepository userland.SessionRepository
	SessionService    session.Service
}

func NewSessionServiceTestSuite(cfg *config.Configuration) *SessionServiceTestSuite {
	return &SessionServiceTestSuite{
		Config: cfg,
	}
}

// before each test
func (suite SessionServiceTestSuite) SetupTest() {
	if err := suite.RedisClient.FlushAll().Err(); err != nil {
		suite.T().Fatalf("RedisClient.FlushAll() err = %v; want nil", err)
	}
}

func (suite *SessionServiceTestSuite) SetupSuite() {
	suite.T().Log("Connecting to redis at", suite.Config.Redis)
	redisClient, err := redis.CreateClient(suite.Config.Redis, 0)
	if err != nil {
		suite.T().Fatalf("redis.CreateClient() err = %v", err)
	}

	suite.RedisClient = redisClient
	suite.KeyValueService = redis.NewKeyValueService(redisClient)
	suite.SessionRepository = redis.NewSessionRepository(redisClient)
	suite.SessionService = session.NewService(
		session.WithKeyValueService(suite.KeyValueService),
		session.WithSessionRepository(suite.SessionRepository),
	)
	suite.SessionService = session.NewInstrumentorService(
		metrics.PrometheusRequestLatency("service", "authentication", authentication.MetricKeys),
		suite.SessionService,
	)
}

func (suite SessionServiceTestSuite) TestCreateSession() {
	type args struct {
		userID  int
		session userland.Session
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
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
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if err := suite.SessionService.CreateSession(tc.args.userID, tc.args.session); err != tc.wantErr {
				t.Fatalf("SessionService.CreateSession(%d, <session>) err = %v; want %v", tc.args.userID, err, tc.wantErr)
			}
		})
	}
}

func (suite SessionServiceTestSuite) TestListSession() {
	type args struct {
		userID           int
		numberOfSessions int
	}
	testCases := []struct {
		name             string
		args             args
		wantSessionCount int
	}{
		{
			name: "list 3 sessions",
			args: args{
				userID:           1,
				numberOfSessions: 3,
			},
			wantSessionCount: 3,
		},
		{
			name: "list 0 sessions",
			args: args{
				userID:           2,
				numberOfSessions: 0,
			},
			wantSessionCount: 0,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			userlandtest.TestCreateSessions(t, suite.SessionRepository,
				userlandtest.WithNumberOfSessions(tc.args.numberOfSessions),
				userlandtest.WithUserID(tc.args.userID),
			)
			sessions, err := suite.SessionService.ListSession(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionService.ListSession(%d) err = %v; want nil", tc.args.userID, err)
			}
			gotCount := len(sessions)
			if gotCount != tc.wantSessionCount {
				t.Errorf("SessionService.ListSession(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, tc.wantSessionCount)
			}
		})
	}
}

func (suite SessionServiceTestSuite) TestEndSession() {
	type args struct {
		userID int
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID: 1,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			session := userlandtest.TestCreateSession(t, suite.SessionRepository, userlandtest.WithUserID(tc.args.userID))
			if err := suite.SessionService.EndSession(tc.args.userID, session.ID); err != tc.wantErr {
				t.Fatalf("SessionService.EndSession(%d, <sessionID>) err = %v; want %v", tc.args.userID, err, tc.wantErr)
			}
			sessions, err := suite.SessionService.ListSession(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionService.ListSession(%d) err = %v; want nil", tc.args.userID, err)
			}
			gotCount := len(sessions)
			if gotCount != 0 {
				t.Errorf("SessionService.ListSession(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, 0)
			}
		})
	}
}

func (suite SessionServiceTestSuite) TestEndOtherSessions() {
	type args struct {
		userID           int
		numberOfSessions int
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID:           1,
				numberOfSessions: 3,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			sessions := userlandtest.TestCreateSessions(t, suite.SessionRepository, userlandtest.WithNumberOfSessions(tc.args.numberOfSessions))
			if err := suite.SessionService.EndOtherSessions(tc.args.userID, sessions[0].ID); err != tc.wantErr {
				t.Fatalf("SessionService.EndOtherSessions(%d, <sessions[0].ID>) err = %v; want %v", tc.args.userID, err, tc.wantErr)
			}
			sessions, err := suite.SessionService.ListSession(tc.args.userID)
			if err != nil {
				t.Fatalf("SessionService.ListSession(%d) err = %v; want nil", tc.args.userID, err)
			}
			gotCount := len(sessions)
			if gotCount != 1 {
				t.Errorf("SessionService.ListSession(%d) len(sessions) = %d; want %d", tc.args.userID, gotCount, 1)
			}
		})
	}
}

func (suite SessionServiceTestSuite) TestCreateRefreshToken() {
	type args struct {
		user      userland.User
		sessionID string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				user: userland.User{
					ID:       1,
					Fullname: "adhitya",
					Email:    "test@coba.com",
				},
				sessionID: security.GenerateUUID(),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if _, err := suite.SessionService.CreateRefreshToken(tc.args.user, tc.args.sessionID); err != nil {
				t.Fatalf("SessionService.CreateRefreshToken() err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite SessionServiceTestSuite) TestCreateNewAccessToken() {
	type args struct {
		user      userland.User
		sessionID string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				user: userland.User{
					ID:       1,
					Fullname: "adhitya",
					Email:    "test@coba.com",
				},
				sessionID: security.GenerateUUID(),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			refreshToken, err := suite.SessionService.CreateRefreshToken(tc.args.user, tc.args.sessionID)
			if err != nil {
				t.Fatalf("SessionService.CreateRefreshToken() err = %v; want nil", err)
			}

			if _, err = suite.SessionService.CreateNewAccessToken(tc.args.user, refreshToken.Key); err != tc.wantErr {
				t.Fatalf("SessionService.CreateNewAccessToken() err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}
