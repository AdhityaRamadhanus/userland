// +build all profile_service

package profile_test

import (
	"os"
	"testing"

	_redis "github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/sarulabs/di"
	"github.com/stretchr/testify/suite"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/service/profile"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/gcs"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	log "github.com/sirupsen/logrus"
)

type ProfileServiceTestSuite struct {
	suite.Suite
	DB              *sqlx.DB
	RedisClient     *_redis.Client
	EventRepository userland.EventRepository
	UserRepository  userland.UserRepository
	KeyValueService userland.KeyValueService
	ProfileService  profile.Service
}

func (suite ProfileServiceTestSuite) SetupTest() {
	if _, err := suite.DB.Query("DELETE FROM users"); err != nil {
		log.Fatal("Failed to setup database ", err)
	}

	if err := suite.RedisClient.FlushAll().Err(); err != nil {
		log.Fatal("Cannot setup redis")
	}
}

func (suite ProfileServiceTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		postgres.ConnectionBuilder,
		redis.ConnectionBuilder("redis-connection", 0),
		mailing.ClientBuilder,
		redis.KeyValueServiceBuilder,
		postgres.UserRepositoryBuilder,
		postgres.EventRepositoryBuilder,
		gcs.ServiceBuilder,
		profile.ServiceBuilder,
		profile.ServiceInstrumentorBuilder,
	)

	return builder.Build()
}

func (suite *ProfileServiceTestSuite) SetupSuite() {
	godotenv.Load("../../../.env")
	os.Setenv("ENV", "testing")

	ctn := suite.BuildContainer()
	suite.DB = ctn.Get("postgres-connection").(*sqlx.DB)
	suite.RedisClient = ctn.Get("redis-connection").(*_redis.Client)
	suite.KeyValueService = ctn.Get("keyvalue-service").(userland.KeyValueService)
	suite.UserRepository = ctn.Get("user-repository").(userland.UserRepository)
	suite.EventRepository = ctn.Get("event-repository").(userland.EventRepository)
	suite.ProfileService = ctn.Get("profile-instrumentor-service").(profile.Service)
}

func TestProfileService(t *testing.T) {
	suiteTest := new(ProfileServiceTestSuite)
	suite.Run(t, suiteTest)
}

func (suite ProfileServiceTestSuite) TestProfile() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			ExpectError: false,
		},
		{
			UserID:      lastUserID + 1,
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		_, err := suite.ProfileService.Profile(testCase.UserID)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite ProfileServiceTestSuite) TestSetProfile() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		Fullname string
		Location string
		Bio      string
		Web      string
		UserID   int
	}{
		{
			UserID:   lastUserID,
			Fullname: "Awesome User",
			Location: "Jakarta, Indonesia",
			Bio:      "My Short Bio",
			Web:      "https://example.com",
		},
		{
			UserID: lastUserID + 1,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)

		user.Fullname = testCase.Fullname
		user.Location = testCase.Location
		user.Bio = testCase.Bio
		user.WebURL = testCase.Web

		err := suite.ProfileService.SetProfile(user)
		suite.Nil(err)
	}
}

func (suite ProfileServiceTestSuite) TestRequestChangeEmail() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@icehousecorp.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)

	row = suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		NewEmail    string
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			NewEmail:    "adhitya.ice@housecorp.com",
			ExpectError: false,
		},
		{
			UserID:      lastUserID,
			NewEmail:    "adhitya.ramadhanus@icehousecorp.com",
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		_, err := suite.ProfileService.RequestChangeEmail(user, testCase.NewEmail)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite ProfileServiceTestSuite) TestChangeEmail() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		NewEmail string
		UserID   int
	}{
		{
			UserID:   lastUserID,
			NewEmail: "adhitya.ice@housecorp.com",
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		verificationID, _ := suite.ProfileService.RequestChangeEmail(user, testCase.NewEmail)
		err := suite.ProfileService.ChangeEmail(user, verificationID)
		suite.Nil(err)
	}
}

func (suite ProfileServiceTestSuite) TestChangePassword() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		NewPassword string
		OldPassword string
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			OldPassword: "test123",
			NewPassword: "test12345",
			ExpectError: false,
		},
		{
			UserID:      lastUserID,
			OldPassword: "test1234",
			NewPassword: "test12345",
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		err := suite.ProfileService.ChangePassword(user, testCase.OldPassword, testCase.NewPassword)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite ProfileServiceTestSuite) TestEnrollTFA() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		_, _, err := suite.ProfileService.EnrollTFA(user)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite ProfileServiceTestSuite) TestActivateTFA() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		UserID      int
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		secret, _, _ := suite.ProfileService.EnrollTFA(user)

		tfaActivationKey := keygenerator.TFAActivationKey(user.ID, secret)
		code, _ := suite.KeyValueService.Get(tfaActivationKey)

		_, err := suite.ProfileService.ActivateTFA(user, secret, string(code))
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite ProfileServiceTestSuite) TestRemoveTFA() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		UserID      int
		Password    string
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			Password:    "test123",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		err := suite.ProfileService.RemoveTFA(user, testCase.Password)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite ProfileServiceTestSuite) TestDeleteAccount() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	testCases := []struct {
		UserID      int
		Password    string
		ExpectError bool
	}{
		{
			UserID:      lastUserID,
			Password:    "test1234",
			ExpectError: true,
		},
		{
			UserID:      lastUserID,
			Password:    "test123",
			ExpectError: false,
		},
	}

	for _, testCase := range testCases {
		user, _ := suite.ProfileService.Profile(testCase.UserID)
		err := suite.ProfileService.DeleteAccount(user, testCase.Password)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}
