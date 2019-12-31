// +build integration

package profile_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	"github.com/AdhityaRamadhanus/userland/pkg/userlandtest"
	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type ProfileServiceTestSuite struct {
	suite.Suite
	Config          *config.Configuration
	DB              *sqlx.DB
	RedisClient     *_redis.Client
	UserRepository  userland.UserRepository
	KeyValueService userland.KeyValueService
	ProfileService  profile.Service
}

func NewProfileServiceTestSuite(cfg *config.Configuration) *ProfileServiceTestSuite {
	return &ProfileServiceTestSuite{
		Config: cfg,
	}
}

func (suite *ProfileServiceTestSuite) Teardown() {
	suite.T().Log("Teardown ProfileServiceTestSuite")
	suite.RedisClient.Close()
	suite.DB.Close()
}

func (suite *ProfileServiceTestSuite) SetupSuite() {
	suite.T().Log("Connecting to postgres at", suite.Config.Postgres)
	pgConn, err := postgres.CreateConnection(suite.Config.Postgres)
	if err != nil {
		suite.T().Fatalf("postgres.CreateConnection() err = %v", err)
	}
	suite.T().Log("Connecting to redis at", suite.Config.Redis)
	redisClient, err := redis.CreateClient(suite.Config.Redis, 0)
	if err != nil {
		suite.T().Fatalf("redis.CreateClient() err = %v", err)
	}

	suite.DB = pgConn
	suite.RedisClient = redisClient
	suite.KeyValueService = redis.NewKeyValueService(redisClient)
	suite.UserRepository = postgres.NewUserRepository(pgConn)
	suite.ProfileService = profile.NewService(
		profile.WithKeyValueService(suite.KeyValueService),
		profile.WithMailingClient(mailing.NewMailingClient("")),
		profile.WithUserRepository(suite.UserRepository),
	)
	suite.ProfileService = profile.NewInstrumentorService(
		metrics.PrometheusRequestLatency("service", "authentication", profile.MetricKeys),
		suite.ProfileService,
	)
}

func (suite *ProfileServiceTestSuite) SetupTest() {
	queries := []string{
		"DELETE FROM users",
		"DELETE FROM events",
	}

	for _, query := range queries {
		if _, err := suite.DB.Query(query); err != nil {
			suite.T().Fatalf("DB.Query(%q) err = %v; want nil", query, err)
		}
	}

	if err := suite.RedisClient.FlushAll().Err(); err != nil {
		suite.T().Fatalf("RedisClient.FlushAll() err = %v; want nil", err)
	}
}

func (suite ProfileServiceTestSuite) TestProfile() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)
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
				userID: defaultUser.ID,
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			args: args{
				userID: defaultUser.ID + 1,
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != tc.wantErr {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want %v", tc.args.userID, err, tc.wantErr)
			}

			if tc.wantErr != nil {
				return
			}

			if user.ID != tc.args.userID {
				t.Errorf("ProfileService.Profile(%d) user.ID = %d; want %d", tc.args.userID, user.ID, tc.args.userID)
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestSetProfile() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		userID   int
		fullname string
		location string
		bio      string
		web      string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID:   defaultUser.ID,
				fullname: "Awesome User",
				location: "Jakarta, Indonesia",
				bio:      "My Short Bio",
				web:      "https://example.com",
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			args: args{
				userID:   defaultUser.ID + 1,
				fullname: "Awesome User",
				location: "Jakarta, Indonesia",
				bio:      "My Short Bio",
				web:      "https://example.com",
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user := userland.User{
				ID:       tc.args.userID,
				Fullname: tc.args.fullname,
				Location: tc.args.location,
				Bio:      tc.args.bio,
				WebURL:   tc.args.web,
			}

			if err := suite.ProfileService.SetProfile(user); err != tc.wantErr {
				t.Fatalf("ProfileService.SetProfile(user) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestRequestChangeEmail() {
	// email is adhitya.ramadhanus@gmail.com
	userlandtest.TestCreateUser(suite.T(), suite.UserRepository)
	anotherUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository, userlandtest.WithUserEmail("adhitya.ramadhanus_1993@gmail.com"))

	type args struct {
		userID   int
		newEmail string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID:   anotherUser.ID,
				newEmail: "adhitya.ramadhanus_1889@gmail.com",
			},
			wantErr: nil,
		},
		{
			name: "err already used",
			args: args{
				userID:   anotherUser.ID,
				newEmail: "adhitya.ramadhanus@gmail.com",
			},
			wantErr: profile.ErrEmailAlreadyUsed,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}

			if _, err := suite.ProfileService.RequestChangeEmail(user, tc.args.newEmail); err != tc.wantErr {
				t.Fatalf("ProfileService.RequestChangeEmail(<user>, %s) err = %v; want %v", tc.args.newEmail, err, tc.wantErr)
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestChangeEmail() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		newEmail string
		userID   int
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				newEmail: "adhitya.ramadhanus_1993@gmail.com",
				userID:   defaultUser.ID,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}
			verificationID, err := suite.ProfileService.RequestChangeEmail(user, tc.args.newEmail)
			if err != nil {
				t.Fatalf("ProfileService.RequestChangeEmail(<user>, %q) err = %v; want nil", tc.args.newEmail, err)
			}
			if err := suite.ProfileService.ChangeEmail(user, verificationID); err != tc.wantErr {
				t.Fatalf("ProfileService.ChangeEmail(<user>, verificationID) err = %v; want %v", err, tc.wantErr)
			}
			// check by finding user by email
			if _, err := suite.ProfileService.ProfileByEmail(tc.args.newEmail); err != nil {
				t.Errorf("ProfileService.ProfileByEmail(%q) err = %v; want nil", tc.args.newEmail, err)
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestChangePassword() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository, userlandtest.Verified(true))

	type args struct {
		newPassword string
		oldPassword string
		userID      int
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID:      defaultUser.ID,
				oldPassword: "test123",
				newPassword: "test12345",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}

			if err := suite.ProfileService.ChangePassword(user, tc.args.oldPassword, tc.args.newPassword); err != tc.wantErr {
				t.Fatalf("ProfileService.ChangePassword(user, oldpass, newpass) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestEnrollTFA() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)
	tfaEnabledUser := userlandtest.TestCreateTFAEnabledUser(suite.T(), suite.UserRepository, userlandtest.WithUserEmail("tfa@gmail.com"))

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
				userID: defaultUser.ID,
			},
			wantErr: nil,
		},
		{
			name: "tfa already enabled",
			args: args{
				userID: tfaEnabledUser.ID,
			},
			wantErr: profile.ErrTFAAlreadyEnabled,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}
			if _, _, err := suite.ProfileService.EnrollTFA(user); err != tc.wantErr {
				t.Fatalf("ProfileService.EnrollTFA(user) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestActivateTFA() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

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
				userID: defaultUser.ID,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}
			secret, _, err := suite.ProfileService.EnrollTFA(user)
			if err != nil {
				t.Fatalf("ProfileService.EnrollTFA(user) err = %v; want %v", err, tc.wantErr)
			}

			tfaActivationKey := keygenerator.TFAActivationKey(user.ID, secret)
			code, err := suite.KeyValueService.Get(tfaActivationKey)
			if err != nil {
				t.Fatalf("KeyValueService.Get(%q) err = %v; want %v", tfaActivationKey, err, tc.wantErr)
			}

			if _, err := suite.ProfileService.ActivateTFA(user, secret, string(code)); err != tc.wantErr {
				t.Fatalf("ProfileService.ActivateTFA(<user>, secret, code) err = %v; want %v", err, tc.wantErr)
			}

			if tc.wantErr != nil {
				return
			}

			user, err = suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}

			if !user.TFAEnabled {
				t.Error("user.TFAEnabled = false; want true")
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestRemoveTFA() {
	defaultUser := userlandtest.TestCreateTFAEnabledUser(suite.T(), suite.UserRepository)
	type args struct {
		userID   int
		password string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID:   defaultUser.ID,
				password: userlandtest.DefaultUserPassword,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}
			if err := suite.ProfileService.RemoveTFA(user, tc.args.password); err != tc.wantErr {
				t.Fatalf("ProfileService.RemoveTFA(user, password) err = %v; want %v", err, tc.wantErr)
			}

			user, err = suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}

			if user.TFAEnabled {
				t.Error("user.TFAEnabled = true; want false")
			}
		})
	}
}

func (suite ProfileServiceTestSuite) TestDeleteAccount() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)
	anotherUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository, userlandtest.WithUserEmail("another@gmail.com"))
	type args struct {
		userID   int
		password string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID:   defaultUser.ID,
				password: userlandtest.DefaultUserPassword,
			},
			wantErr: nil,
		},
		{
			name: "wrong password",
			args: args{
				userID:   anotherUser.ID,
				password: "some-gibberish-password",
			},
			wantErr: profile.ErrWrongPassword,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.ProfileService.Profile(tc.args.userID)
			if err != nil {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}
			if err := suite.ProfileService.DeleteAccount(user, tc.args.password); err != tc.wantErr {
				t.Fatalf("ProfileService.Profile(%d) err = %v; want nil", tc.args.userID, err)
			}

			if tc.wantErr != nil {
				return
			}

			if _, err := suite.ProfileService.Profile(tc.args.userID); err != userland.ErrUserNotFound {
				t.Errorf("failed in deleting account, user %d still exist", tc.args.userID)
			}
		})
	}
}
