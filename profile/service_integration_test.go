// +build all service

package profile_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/profile"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	log "github.com/sirupsen/logrus"
)

var (
	userRepository  *postgres.UserRepository
	keyValueService *redis.KeyValueService
	profileService  profile.Service
	currentUserID   int
)

func Setup(db *sqlx.DB, redisClient *_redis.Client) {
	_, err := db.Query("DELETE FROM users")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from users"))
	}

	row := db.QueryRow(
		`INSERT INTO users (fullname, email, password, createdat, updatedat)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@icehousecorp.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)

	row = db.QueryRow(
		`INSERT INTO users (fullname, email, password, createdat, updatedat)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&currentUserID)

	err = redisClient.FlushAll().Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}

	time.Sleep(time.Second * 2)
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
	profileService = profile.NewService(userRepository, keyValueService)

	code := m.Run()
	os.Exit(code)
}

func TestProfileIntegration(t *testing.T) {
	testCases := []struct {
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      currentUserID,
			ExpectError: false,
		},
		{
			UserID:      currentUserID + 1,
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		_, err := profileService.Profile(testCase.UserID)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestSetProfileIntegration(t *testing.T) {
	testCases := []struct {
		Fullname string
		Location string
		Bio      string
		Web      string
		UserID   int
	}{
		{
			UserID:   currentUserID,
			Fullname: "Awesome User",
			Location: "Jakarta, Indonesia",
			Bio:      "My Short Bio",
			Web:      "https://example.com",
		},
		{
			UserID: currentUserID + 1,
		},
	}

	for _, testCase := range testCases {
		user, _ := profileService.Profile(testCase.UserID)

		user.Fullname = testCase.Fullname
		user.Location = testCase.Location
		user.Bio = testCase.Bio
		user.WebURL = testCase.Web

		err := profileService.SetProfile(user)
		assert.Nil(t, err)
	}
}

func TestRequestChangeEmailIntegration(t *testing.T) {
	testCases := []struct {
		NewEmail    string
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      currentUserID,
			NewEmail:    "adhitya.ice@housecorp.com",
			ExpectError: false,
		},
		{
			UserID:      currentUserID,
			NewEmail:    "adhitya.ramadhanus@icehousecorp.com",
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		user, _ := profileService.Profile(testCase.UserID)
		_, err := profileService.RequestChangeEmail(user, testCase.NewEmail)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestChangeEmailIntegration(t *testing.T) {
	testCases := []struct {
		NewEmail string
		UserID   int
	}{
		{
			UserID:   currentUserID,
			NewEmail: "adhitya.ice@housecorp.com",
		},
	}

	for _, testCase := range testCases {
		user, _ := profileService.Profile(testCase.UserID)
		verificationID, _ := profileService.RequestChangeEmail(user, testCase.NewEmail)
		err := profileService.ChangeEmail(user, verificationID)
		assert.Nil(t, err)
	}
}
