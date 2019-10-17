// +build all user postgres_repository

package postgres_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/sarulabs/di"
	"github.com/stretchr/testify/suite"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	log "github.com/sirupsen/logrus"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	DB             *sqlx.DB
	UserRepository userland.UserRepository
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	_, err := suite.DB.Query("DELETE FROM users")
	if err != nil {
		log.Fatal("Failed to setup database ", err)
	}
}

func (suite *UserRepositoryTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		postgres.ConnectionBuilder,
		postgres.UserRepositoryBuilder,
	)

	return builder.Build()
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")
	ctn := suite.BuildContainer()

	// Repositories
	suite.DB = ctn.Get("postgres-connection").(*sqlx.DB)
	suite.UserRepository = ctn.Get("user-repository").(userland.UserRepository)
}

func TestUserRepository(t *testing.T) {
	suiteTest := new(UserRepositoryTestSuite)
	suite.Run(t, suiteTest)
}

func (suite *UserRepositoryTestSuite) TestInsert() {
	testCases := []struct {
		User        userland.User
		ExpectError bool
	}{
		{
			User: userland.User{
				Email:    "adhitya.ramadhanus@icehousecorp.com",
				Fullname: "Adhitya Ramadhanus",
				Password: "test123",
			},
			ExpectError: false,
		},
		{
			User: userland.User{
				Email:    "adhitya.ramadhanus@icehousecorp.com",
				Fullname: "Adhitya Ramadhanus",
				Password: "test123",
			},
			ExpectError: true,
		},
	}
	for _, testCase := range testCases {
		err := suite.UserRepository.Insert(testCase.User)
		if testCase.ExpectError {
			suite.NotNil(err)
		} else {
			suite.Nil(err)
		}
	}
}

func (suite *UserRepositoryTestSuite) TestFindByEmail() {
	suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now())`,
		security.HashPassword("test123"),
	)

	email := "adhitya.ramadhanus@gmail.com"
	user, err := suite.UserRepository.FindByEmail(email)

	suite.Nil(err, "Failed to find user by email")
	suite.Equal(user.Email, email)
}

func (suite *UserRepositoryTestSuite) TestFindByID() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)
	_, err := suite.UserRepository.Find(lastUserID)
	suite.Nil(err, "Failed to find user by id")
}

func (suite *UserRepositoryTestSuite) TestUpdate() {
	email := "adhitya.ramadhanus@icehousecorp.com"
	suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', $1, $2, now(), now()) RETURNING id`,
		email,
		security.HashPassword("test123"),
	)
	user, _ := suite.UserRepository.FindByEmail(email)

	user.Phone = "0812567823823"
	user.Bio = "Test Aja"
	err := suite.UserRepository.Update(user)
	suite.Nil(err, "Failed to update user by id")
}

func (suite *UserRepositoryTestSuite) TestStoreBackupCodes() {
	email := "adhitya.ramadhanus@icehousecorp.com"
	suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', $1, $2, now(), now()) RETURNING id`,
		email,
		security.HashPassword("test123"),
	)
	user, _ := suite.UserRepository.FindByEmail(email)

	user.BackupCodes = []string{"xxx", "xxx"}
	err := suite.UserRepository.StoreBackupCodes(user)
	suite.Nil(err, "Failed to store backupd codes user")
}

func (suite *UserRepositoryTestSuite) TestDelete() {
	var lastUserID int
	row := suite.DB.QueryRow(
		`INSERT INTO users (fullname, email, password, created_at, updated_at)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', $1, now(), now()) RETURNING id`,
		security.HashPassword("test123"),
	)
	row.Scan(&lastUserID)

	err := suite.UserRepository.Delete(lastUserID)
	suite.Nil(err, "Failed to delete user by id")
}
