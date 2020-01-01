package postgres

import (
	"fmt"

	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/jmoiron/sqlx"
)

func CreateConnection(cfg config.PostgresConfig) (*sqlx.DB, error) {
	connString := fmt.Sprintf(
		`host=%s port=%d user=%s dbname=%s sslmode=%s password=%s`,
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.DBName,
		cfg.SSLMode,
		cfg.Password,
	)
	return sqlx.Open("postgres", connString)
}
