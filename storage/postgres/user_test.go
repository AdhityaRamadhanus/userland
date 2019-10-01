package postgres_test

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/config"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	log "github.com/sirupsen/logrus"
)

var (
	userRepository *postgres.UserRepository
)

// make test kind of idempotent
func setupDatabase(db *sqlx.DB) {
	_, err := db.Query("DELETE FROM users")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from users"))
	}
}

func TestMain(m *testing.M) {
	if err := config.Init(); err != nil {
		log.Fatal(err)
	}

	pgConnString := postgres.CreateConnectionString()
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	setupDatabase(db)

	// Repositories
	userRepository = postgres.NewUserRepository(db)

	code := m.Run()
	os.Exit(code)
}

func TestCreateUserIntegration(t *testing.T) {
	user := userland.User{
		Fullname: "Adhitya Ramadhanus",
		Email:    "adhitya.ramadhanus@gmail.com",
		Password: "test123",
	}

	err := userRepository.Insert(user)
	if err != nil {
		t.Error("Failed to create user", err)
	}
}
