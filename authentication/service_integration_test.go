package authentication_test

import (
	"fmt"
	"os"
	"testing"

	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/authentication"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	log "github.com/sirupsen/logrus"
)

var (
	userRepository        *postgres.UserRepository
	keyValueService       *redis.KeyValueService
	authenticationService authentication.Service
)

func Setup(db *sqlx.DB, redisClient *_redis.Client) {
	_, err := db.Query("DELETE FROM users")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from users"))
	}

	err = redisClient.FlushAll().Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}
}

func TestMain(m *testing.M) {
	godotenv.Load("../.env")
	pgConnString := postgres.CreateConnectionString()
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	redisClient := _redis.NewClient(&_redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
	})

	_, err = redisClient.Ping().Result()
	if err != nil {
		log.WithError(err).Error("Failed to connect to redis")
	}

	Setup(db, redisClient)

	// Repositories
	keyValueService = redis.NewKeyValueService(redisClient)
	userRepository = postgres.NewUserRepository(db)
	authenticationService = authentication.NewService(userRepository, keyValueService)

	code := m.Run()
	os.Exit(code)
}

func TestRegisterIntegration(t *testing.T) {
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
		err := authenticationService.Register(testCase.User)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestRequestVerificationIntegration(t *testing.T) {
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
		_, err := authenticationService.RequestVerification("email.verify", testCase.Email)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestVerifyAccountIntegration(t *testing.T) {
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
		verificationID, err := authenticationService.RequestVerification("email.verify", testCase.Email)
		assert.Nil(t, err)

		user, err := userRepository.FindByEmail(testCase.Email)
		assert.Nil(t, err)
		// get code
		key := fmt.Sprintf("%s:%d:%s", "email-verify", user.ID, verificationID)
		val, err := keyValueService.Get(key)
		assert.Nil(t, err)

		err = authenticationService.VerifyAccount(verificationID, testCase.Email, string(val))
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestLoginIntegration(t *testing.T) {
	testCases := []struct {
		Email       string
		Password    string
		RequireTFA  bool
		Verified    bool
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test1234",
			RequireTFA:  false,
			Verified:    true,
			ExpectError: false,
		},
		{
			Email:       "adhitya.ramadhanus@gmail.com",
			Password:    "test123",
			RequireTFA:  true,
			Verified:    true,
			ExpectError: false,
		},
		{
			Email:       "adhitya.ramadhanus@gmail.com",
			Password:    "test123",
			RequireTFA:  false,
			Verified:    false,
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		// setup
		user, err := userRepository.FindByEmail(testCase.Email)
		assert.Nil(t, err)
		user.TFAEnabled = testCase.RequireTFA
		user.Verified = testCase.Verified
		userRepository.Update(user)

		requireTFA, _, err := authenticationService.Login(testCase.Email, testCase.Password)
		if testCase.ExpectError {
			assert.NotNil(t, err, "should return error")
		} else {
			assert.Nil(t, err)
			assert.Equal(t, requireTFA, testCase.RequireTFA)
		}
	}
}

func TestVerifyTFAIntegration(t *testing.T) {
	testCases := []struct {
		Email       string
		Password    string
		ExpectError bool
	}{
		{
			Email:       "adhitya.ramadhanus@icehousecorp.com",
			Password:    "test1234",
			ExpectError: false,
		},
		{
			Email:       "adhitya.ramadhanus@gmail.com",
			Password:    "test123",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		// setup
		user, err := userRepository.FindByEmail(testCase.Email)
		assert.Nil(t, err)
		user.TFAEnabled = true
		user.Verified = true
		userRepository.Update(user)

		_, tfaToken, err := authenticationService.Login(testCase.Email, testCase.Password)
		assert.Nil(t, err)

		tfaKey := fmt.Sprintf("%s:%d:%s", "tfa-verify", user.ID, tfaToken.Key)
		expectedCode, err := keyValueService.Get(tfaKey)
		assert.Nil(t, err)

		_, err = authenticationService.VerifyTFA(tfaToken.Key, user.ID, string(expectedCode))
		if testCase.ExpectError {
			assert.NotNil(t, err, "should return error")
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestVerifyTFABypassIntegration(t *testing.T) {
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
		{
			Email:       "adhitya.ramadhanus@gmail.com",
			Password:    "test123",
			BackupCode:  "backupaja1232",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		// setup
		user, err := userRepository.FindByEmail(testCase.Email)
		assert.Nil(t, err)

		user.TFAEnabled = true
		user.Verified = true
		userRepository.Update(user)
		hash, err := bcrypt.GenerateFromPassword([]byte(testCase.BackupCode), bcrypt.MinCost)

		assert.Nil(t, err)
		user.BackupCodes = []string{string(hash)}
		userRepository.StoreBackupCodes(user)

		_, tfaToken, err := authenticationService.Login(testCase.Email, testCase.Password)
		assert.Nil(t, err)

		_, err = authenticationService.VerifyTFABypass(tfaToken.Key, user.ID, string(testCase.BackupCode))
		if testCase.ExpectError {
			assert.NotNil(t, err, "should return error")
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestForgotPasswordIntegration(t *testing.T) {
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
		_, err := authenticationService.ForgotPassword(testCase.Email)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestResetPasswordIntegration(t *testing.T) {
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
		verificationID, err := authenticationService.ForgotPassword(testCase.Email)
		assert.Nil(t, err)

		err = authenticationService.ResetPassword(verificationID, testCase.NewPassword)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}

		_, _, err = authenticationService.Login(testCase.Email, testCase.NewPassword)
		assert.Nil(t, err)
	}
}
