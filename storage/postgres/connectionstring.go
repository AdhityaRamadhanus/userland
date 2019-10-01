package postgres

import (
	"fmt"
	"os"
)

func CreateConnectionString() string {
	return fmt.Sprintf(`
		host=%s 
		port=%s 
		user=%s 
		password=%s 
		dbname=%s 
		sslmode=%s`,
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DBNAME"),
		os.Getenv("POSTGRES_SSLMODE"),
	)
}
