// +build integration

package authentication_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"

	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/service/authentication"
	"github.com/stretchr/testify/suite"
)

type AuthenticationServiceTestSuite struct {
	suite.Suite
	Config                *config.Configuration
	DB                    *sqlx.DB
	RedisClient           *_redis.Client
	UserRepository        userland.UserRepository
	KeyValueService       userland.KeyValueService
	AuthenticationService authentication.Service
}

func NewAuthenticationServiceTestSuite(cfg *config.Configuration) *AuthenticationServiceTestSuite {
	return &AuthenticationServiceTestSuite{
		Config: cfg,
	}
}

// before each test
func (suite *AuthenticationServiceTestSuite) SetupSuite() {
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
	suite.AuthenticationService = authentication.NewService(
		authentication.WithKeyValueService(suite.KeyValueService),
		authentication.WithMailingClient(mailing.NewMailingClient("")),
		authentication.WithUserRepository(suite.UserRepository),
	)
}

func (suite AuthenticationServiceTestSuite) SetupTest() {
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

func (suite AuthenticationServiceTestSuite) TestRegister() {
	type args struct {
		user userland.User
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
					Email:    "adhitya.ramadhanus@gmail.com",
					Fullname: "Adhitya Ramadhanus",
					Password: "test1234",
				},
			},
			wantErr: nil,
		},
		{
			name: "error registered",
			args: args{
				user: userland.User{
					Email:    "adhitya.ramadhanus@gmail.com",
					Fullname: "Adhitya Ramadhanus",
					Password: "test1234",
				},
			},
			wantErr: authentication.ErrUserRegistered,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.AuthenticationService.Register(tc.args.user)
			if err != tc.wantErr {
				t.Fatalf("AuthenticationService.Register(user) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestRequestVerification_email() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	type args struct {
		email string
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				email: "adhitya.ramadhanus@gmail.com",
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			args: args{
				email: "adhitya@gmail.com",
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			_, err := suite.AuthenticationService.RequestVerification("email.verify", tc.args.email)
			if err != tc.wantErr {
				t.Fatalf("AuthenticationService.RequestVerification(%q, %q) err = %v; want %v", "email.verify", tc.args.email, err, tc.wantErr)
			}
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestVerifyAccount_email() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	type args struct {
		email string
	}
	testCases := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				email: "adhitya.ramadhanus@gmail.com",
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			verificationID, err := suite.AuthenticationService.RequestVerification("email.verify", tc.args.email)
			if err != nil {
				t.Fatalf("AuthenticationService.RequestVerification(%q, %q) err = %v; want nil", "email.verify", tc.args.email, err)
			}
			user, err := suite.UserRepository.FindByEmail(tc.args.email)
			if err != nil {
				t.Fatalf("UserRepository.FindByEmail(%q) err = %v; want nil", tc.args.email, err)
			}
			// get code, must be within security.EmailVerificationExpiration
			key := keygenerator.EmailVerificationKey(user.ID, verificationID)
			val, err := suite.KeyValueService.Get(key)
			if err != nil {
				t.Fatalf("KeyValueService.Get(%q) err = %v; want nil", key, err)
			}

			if err := suite.AuthenticationService.VerifyAccount(verificationID, tc.args.email, string(val)); err != nil {
				t.Fatalf("AuthenticationService.VerifyAccount(%q, %q, <val>) err = %v; want nil", verificationID, tc.args.email, err)
			}

			user, err = suite.UserRepository.FindByEmail(tc.args.email)
			if err != nil {
				t.Fatalf("UserRepository.FindByEmail(%q) err = %v; want nil", tc.args.email, err)
			}

			if !user.Verified {
				t.Errorf("User.Verified = false; want true")
			}
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestLogin_withoutTFA() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	type args struct {
		email    string
		password string
		verified bool
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "not_verified_user",
			args: args{
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test123",
				verified: false,
			},
			wantErr: authentication.ErrUserNotVerified,
		},
		{
			name: "verified_user",
			args: args{
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test123",
				verified: true,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// setup
			user, err := suite.UserRepository.FindByEmail(tc.args.email)
			if err != nil {
				t.Fatalf("UserRepository.FindByEmail(%q) err = %v; want nil", tc.args.email, err)
			}

			user.Verified = tc.args.verified
			if err := suite.UserRepository.Update(user); err != nil {
				t.Fatalf("UserRepository.Update(user) err = %v; want nil", err)
			}

			requireTFA, _, err := suite.AuthenticationService.Login(tc.args.email, tc.args.password)
			if err != tc.wantErr {
				t.Fatalf("AuthenticationService.Login(%q, %q) err = %v; want %v", tc.args.email, tc.args.password, err, tc.wantErr)
			}

			// no need to check
			if tc.wantErr != nil {
				return
			}

			if requireTFA {
				t.Errorf("AuthenticationService.Login(%q, %q) requireTFA = true; want false", tc.args.email, tc.args.password)
			}

			// TODO ceck accessToken scope
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestLogin_withTFA() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	type args struct {
		email    string
		password string
		verified bool
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "not_verified_user",
			args: args{
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test123",
				verified: false,
			},
			wantErr: authentication.ErrUserNotVerified,
		},
		{
			name: "verified_user",
			args: args{
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test123",
				verified: true,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// setup
			user, err := suite.UserRepository.FindByEmail(tc.args.email)
			if err != nil {
				t.Fatalf("UserRepository.FindByEmail(%q) err = %v; want nil", tc.args.email, err)
			}

			user.Verified = tc.args.verified
			user.TFAEnabled = true
			user.TFAEnabledAt = time.Now()
			if err := suite.UserRepository.Update(user); err != nil {
				t.Fatalf("UserRepository.Update(user) err = %v; want nil", err)
			}

			requireTFA, _, err := suite.AuthenticationService.Login(tc.args.email, tc.args.password)
			if err != tc.wantErr {
				t.Fatalf("AuthenticationService.Login(%q, %q) err = %v; want %v", tc.args.email, tc.args.password, err, tc.wantErr)
			}

			// no need to check
			if tc.wantErr != nil {
				return
			}

			if !requireTFA {
				t.Errorf("AuthenticationService.Login(%q, %q) requireTFA = false; want true", tc.args.email, tc.args.password)
			}

			// TODO ceck accessToken scope
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestVerifyTFA() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	type args struct {
		email    string
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
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test123",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// setup
			user, err := suite.UserRepository.FindByEmail(tc.args.email)
			if err != nil {
				t.Fatalf("UserRepository.FindByEmail(%q) err = %v; want nil", tc.args.email, err)
			}
			user.TFAEnabled = true
			user.Verified = true
			if err := suite.UserRepository.Update(user); err != nil {
				t.Fatalf("UserRepository.Update(user) err = %v; want nil", err)
			}

			requireTFA, tfaToken, err := suite.AuthenticationService.Login(tc.args.email, tc.args.password)
			if err != nil {
				t.Fatalf("AuthenticationService.Login(%q, %q) err = %v; want %v", tc.args.email, tc.args.password, err, tc.wantErr)
			}

			if !requireTFA {
				t.Errorf("AuthenticationService.Login(%q, %q) requireTFA = false; want true see TestLogin_WithTFA", tc.args.email, tc.args.password)
			}

			tfaKey := keygenerator.TFAVerificationKey(user.ID, tfaToken.Key)
			expectedCode, err := suite.KeyValueService.Get(tfaKey)
			if err != nil {
				t.Fatalf("KeyValueService.Get(%q) err = %v; want nil", tfaKey, err)
			}

			// TODO check access token
			if _, err := suite.AuthenticationService.VerifyTFA(tfaToken.Key, user.ID, string(expectedCode)); err != tc.wantErr {
				t.Fatalf("AuthenticationService.VerifyTFA(%q, %d, <code>) err = %v; want %v", tfaToken.Key, user.ID, err, tc.wantErr)
			}
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestVerifyTFABypass() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test1234"),
	})

	type args struct {
		email       string
		password    string
		backupCodes []string
		usedCode    string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test1234",
				backupCodes: []string{
					"backupCode1",
					"backupCode3",
					"backupCode2",
				},
				usedCode: "backupCode1",
			},
			wantErr: nil,
		},
		{
			name: "wrong backup code",
			args: args{
				email:    "adhitya.ramadhanus@gmail.com",
				password: "test1234",
				backupCodes: []string{
					"backupCode1",
				},
				usedCode: "backupCode2",
			},
			wantErr: authentication.ErrWrongBackupCode,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// setup
			user, err := suite.UserRepository.FindByEmail(tc.args.email)
			if err != nil {
				t.Fatalf("UserRepository.FindByEmail(%q) err = %v; want nil", tc.args.email, err)
			}
			user.TFAEnabled = true
			user.Verified = true
			if err := suite.UserRepository.Update(user); err != nil {
				t.Fatalf("UserRepository.Update(user) err = %v; want nil", err)
			}

			hashedBackupCodes := []string{}
			for _, backupCode := range tc.args.backupCodes {
				hashedBackupCodes = append(hashedBackupCodes, security.HashPassword(backupCode))
			}
			user.BackupCodes = hashedBackupCodes
			if err := suite.UserRepository.StoreBackupCodes(user); err != nil {
				t.Fatalf("UserRepository.StoreBackupCodes(user) err = %v; want nil", err)
			}

			requireTFA, tfaToken, err := suite.AuthenticationService.Login(tc.args.email, tc.args.password)
			if err != nil {
				t.Fatalf("AuthenticationService.Login(%q, %q) err = %v; want %v", tc.args.email, tc.args.password, err, tc.wantErr)
			}

			if !requireTFA {
				t.Errorf("AuthenticationService.Login(%q, %q) requireTFA = false; want true see TestLogin_WithTFA", tc.args.email, tc.args.password)
			}

			// TODO check access token
			if _, err := suite.AuthenticationService.VerifyTFABypass(tfaToken.Key, user.ID, tc.args.usedCode); err != tc.wantErr {
				t.Fatalf("AuthenticationService.VerifyTFABypass(%q, %d, %q) err = %v; want %v", tfaToken.Key, user.ID, tc.args.usedCode, err, tc.wantErr)
			}
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestForgotPassword() {
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test1234"),
	})

	type args struct {
		email string
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				email: "adhitya.ramadhanus@gmail.com",
			},
			wantErr: nil,
		},
		{
			name: "user not found",
			args: args{
				email: "adhitya@gmail.com",
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if _, err := suite.AuthenticationService.ForgotPassword(tc.args.email); err != tc.wantErr {
				t.Fatalf("AuthenticationService.ForgotPassword(%q) err = %v; want %v", tc.args.email, err, tc.wantErr)
			}
		})
	}
}

func (suite AuthenticationServiceTestSuite) TestResetPasswordIntegration() {
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@gmail.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test1234"),
		Verified: true,
	})

	type args struct {
		email       string
		newPassword string
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				email:       "adhitya.ramadhanus@gmail.com",
				newPassword: "test123",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			verificationID, err := suite.AuthenticationService.ForgotPassword(tc.args.email)
			if err != nil {
				t.Fatalf("AuthenticationService.ForgotPassword(%q) err = %v; want %v", tc.args.email, err, tc.wantErr)
			}

			if err := suite.AuthenticationService.ResetPassword(verificationID, tc.args.newPassword); err != tc.wantErr {
				t.Fatalf("AuthenticationService.ResetPassword(%q, %q) err = %v; want %v", verificationID, tc.args.newPassword, err, tc.wantErr)
			}

			if _, _, err := suite.AuthenticationService.Login(tc.args.email, tc.args.newPassword); err != nil {
				t.Fatalf("AuthenticationService.Login(%q, %q) err = %v; want nil", tc.args.email, tc.args.newPassword, err)
			}
		})
	}
}
