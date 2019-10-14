// +build all authentication_service

package authentication_test

import (
	"os"
	"testing"

	"github.com/AdhityaRamadhanus/userland/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/pkg/errors"

	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sarulabs/di"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/service/authentication"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type AuthenticationServiceTestSuite struct {
	suite.Suite
	DB                    *sqlx.DB
	RedisClient           *_redis.Client
	UserRepository        userland.UserRepository
	KeyValueService       userland.KeyValueService
	AuthenticationService authentication.Service
}

func (suite *AuthenticationServiceTestSuite) SetupTest() {
	_, err := suite.DB.Exec("DELETE FROM users")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from users"))
	}

	err = suite.RedisClient.FlushAll().Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}
}

func (suite *AuthenticationServiceTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		postgres.ConnectionBuilder,
		redis.ConnectionBuilder,
		mailing.ClientBuilder,
		redis.KeyValueServiceBuilder,
		postgres.UserRepositoryBuilder,
		authentication.ServiceBuilder,
		authentication.ServiceInstrumentorBuilder,
	)

	return builder.Build()
}

// before each test
func (suite *AuthenticationServiceTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")

	ctn := suite.BuildContainer()

	suite.DB = ctn.Get("postgres-connection").(*sqlx.DB)
	suite.RedisClient = ctn.Get("redis-connection").(*_redis.Client)
	suite.KeyValueService = ctn.Get("keyvalue-service").(userland.KeyValueService)
	suite.UserRepository = ctn.Get("user-repository").(userland.UserRepository)
	suite.AuthenticationService = ctn.Get("authentication-instrumentor-service").(authentication.Service)
}

func TestAuthenticationService(t *testing.T) {
	suiteTest := new(AuthenticationServiceTestSuite)
	suite.Run(t, suiteTest)
}

func (suite *AuthenticationServiceTestSuite) TestRegisterIntegration() {
	testCases := []struct {
		User        userland.User
		ExpectError bool
	}{
		{
			User: userland.User{
				Email:    "adhitya.ramadhanus@icehousecorp.com",
				Fullname: "Adhitya Ramadhanus",
				Password: "test1234",
			},
			ExpectError: false,
		},
		{
			User: userland.User{
				Email:    "adhitya.ramadhanus@gmail.com",
				Fullname: "Adhitya Ramadhanus",
				Password: "test123",
			},
			ExpectError: false,
		},
		{
			User: userland.User{
				Email:    "adhitya.ramadhanus@gmail.com",
				Fullname: "Adhitya Ramadhanus",
				Password: "test123",
			},
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		err := suite.AuthenticationService.Register(testCase.User)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestRequestVerificationIntegration() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	testCases := []struct {
		Email       string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		_, err := suite.AuthenticationService.RequestVerification("email.verify", testCase.Email)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestVerifyAccountIntegration() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	testCases := []struct {
		Email       string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		verificationID, err := suite.AuthenticationService.RequestVerification("email.verify", testCase.Email)
		suite.Nil(err)

		user, err := suite.UserRepository.FindByEmail(testCase.Email)
		suite.Nil(err)
		// get code
		key := keygenerator.EmailVerificationKey(user.ID, verificationID)
		val, err := suite.KeyValueService.Get(key)
		suite.Nil(err)

		err = suite.AuthenticationService.VerifyAccount(verificationID, testCase.Email, string(val))
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestLoginIntegration() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	testCases := []struct {
		Email       string
		Password    string
		RequireTFA  bool
		Verified    bool
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test123",
			RequireTFA:  false,
			Verified:    true,
			ExpectError: false,
		},
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test123",
			RequireTFA:  true,
			Verified:    true,
			ExpectError: false,
		},
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test123",
			RequireTFA:  false,
			Verified:    false,
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		// setup
		user, err := suite.UserRepository.FindByEmail(testCase.Email)
		suite.Nil(err)
		user.TFAEnabled = testCase.RequireTFA
		user.Verified = testCase.Verified
		suite.UserRepository.Update(user)

		requireTFA, _, err := suite.AuthenticationService.Login(testCase.Email, testCase.Password)
		if testCase.ExpectError {
			suite.NotNil(err, "should return error")
		} else {
			suite.Nil(err)
			suite.Equal(requireTFA, testCase.RequireTFA)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestVerifyTFAIntegration() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test123"),
	})

	testCases := []struct {
		Email       string
		Password    string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test123",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		// setup
		user, err := suite.UserRepository.FindByEmail(testCase.Email)
		suite.Nil(err)
		user.TFAEnabled = true
		user.Verified = true
		suite.UserRepository.Update(user)

		_, tfaToken, err := suite.AuthenticationService.Login(testCase.Email, testCase.Password)
		suite.Nil(err)

		tfaKey := keygenerator.TFAVerificationKey(user.ID, tfaToken.Key)
		expectedCode, err := suite.KeyValueService.Get(tfaKey)
		suite.Nil(err)

		_, err = suite.AuthenticationService.VerifyTFA(tfaToken.Key, user.ID, string(expectedCode))
		if testCase.ExpectError {
			suite.NotNil(err, "should return error")
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestVerifyTFABypassIntegration() {
	// setup
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test1234"),
	})

	testCases := []struct {
		Email       string
		Password    string
		BackupCode  string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test1234",
			BackupCode:  "backupaja1232",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		// setup
		user, err := suite.UserRepository.FindByEmail(testCase.Email)
		suite.Nil(err)

		user.TFAEnabled = true
		user.Verified = true
		err = suite.UserRepository.Update(user)
		suite.Nil(err)

		user.BackupCodes = []string{security.HashPassword(testCase.BackupCode)}
		suite.UserRepository.StoreBackupCodes(user)

		_, tfaToken, err := suite.AuthenticationService.Login(testCase.Email, testCase.Password)
		suite.Nil(err)

		_, err = suite.AuthenticationService.VerifyTFABypass(tfaToken.Key, user.ID, string(testCase.BackupCode))
		if testCase.ExpectError {
			suite.NotNil(err, "should return error")
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestForgotPasswordIntegration() {
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test1234"),
	})

	testCases := []struct {
		Email       string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			ExpectError: false,
		},
		{
			Email:       "adhitya.ramadhanus@secorp.com",
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		_, err := suite.AuthenticationService.ForgotPassword(testCase.Email)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *AuthenticationServiceTestSuite) TestResetPasswordIntegration() {
	suite.UserRepository.Insert(userland.User{
		Email:    "adhitya.ramadhanus@icehousecorp.com",
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword("test1234"),
		Verified: true,
	})

	testCases := []struct {
		Email       string
		NewPassword string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			NewPassword: "test12345",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		verificationID, err := suite.AuthenticationService.ForgotPassword(testCase.Email)
		suite.Nil(err)

		err = suite.AuthenticationService.ResetPassword(verificationID, testCase.NewPassword)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}

		_, _, err = suite.AuthenticationService.Login(testCase.Email, testCase.NewPassword)
		suite.Nil(err)
	}
}
