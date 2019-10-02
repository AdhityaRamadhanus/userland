package postgres_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	log "github.com/sirupsen/logrus"
)

var (
	userRepository *postgres.UserRepository
)

// make test kind of idempotent
func Setup(db *sqlx.DB) {
	_, err := db.Query("DELETE FROM users")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from users"))
	}

	_, err = db.Query(
		`INSERT INTO users (fullname, email, password, createdat, updatedat)
		VALUES ('Adhitya Ramadhanus', 'adhitya.ramadhanus@gmail.com', crypt('test123', gen_salt('bf')), now(), now())`)

	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in filling users table"))
	}
}

func TestMain(m *testing.M) {
	godotenv.Load("../../.env")
	pgConnString := postgres.CreateConnectionString()
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	Setup(db)
	// Repositories
	userRepository = postgres.NewUserRepository(db)

	code := m.Run()
	os.Exit(code)
}

func TestCreateUserIntegration(t *testing.T) {
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
				Email:    "adhitya.ramadhanus@gmail.com",
				Fullname: "Adhitya Ramadhanus",
				Password: "test123",
			},
			ExpectError: true,
		},
	}
	for _, testCase := range testCases {
		err := userRepository.Insert(testCase.User)
		if testCase.ExpectError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestFindUserByEmailIntegration(t *testing.T) {
	email := "adhitya.ramadhanus@gmail.com"
	user, err := userRepository.FindByEmail(email)

	assert.Nil(t, err, "Failed to find user by email")
	assert.Equal(t, user.Email, email)
}

func TestFindUserByIDIntegration(t *testing.T) {
	email := "adhitya.ramadhanus@gmail.com"
	user, err := userRepository.FindByEmail(email)

	userID := user.ID
	user, err = userRepository.Find(userID)

	assert.Nil(t, err, "Failed to find user by id")
	assert.Equal(t, user.ID, userID)
}

func TestUpdateUserByIDIntegration(t *testing.T) {
	email := "adhitya.ramadhanus@icehousecorp.com"
	user, err := userRepository.FindByEmail(email)

	user.Phone = "0812567823823"
	user.Bio = "Test Aja"
	err = userRepository.Update(user)
	assert.Nil(t, err, "Failed to update user by id")
}

func TestStoreBackupCodesByIDIntegration(t *testing.T) {
	email := "adhitya.ramadhanus@icehousecorp.com"
	user, err := userRepository.FindByEmail(email)

	user.BackupCodes = []string{"xxx", "xxx"}
	err = userRepository.StoreBackupCodes(user)
	assert.Nil(t, err, "Failed to store backupd codes user")
}

func TestDeleteUserByIDIntegration(t *testing.T) {
	email := "adhitya.ramadhanus@gmail.com"
	user, err := userRepository.FindByEmail(email)

	userID := user.ID
	err = userRepository.Delete(userID)
	assert.Nil(t, err, "Failed to delete user by id")
}
