package postgres

import (
	"fmt"
	"os"
)

func CreateConnectionString() string {
	if os.Getenv("ENV") == "testing" {
		return fmt.Sprintf(
			`host=%s port=%s user=%s dbname=%s sslmode=%s password=%s`,
			os.Getenv("TEST_POSTGRES_HOST"),
			os.Getenv("TEST_POSTGRES_PORT"),
			os.Getenv("TEST_POSTGRES_USER"),
			os.Getenv("TEST_POSTGRES_DBNAME"),
			os.Getenv("TEST_POSTGRES_SSLMODE"),
			os.Getenv("TEST_POSTGRES_PASSWORD"),
		)
	}

	return fmt.Sprintf(
		`host=%s port=%s user=%s dbname=%s sslmode=%s password=%s`,
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_DBNAME"),
		os.Getenv("POSTGRES_SSLMODE"),
		os.Getenv("POSTGRES_PASSWORD"),
	)
}
