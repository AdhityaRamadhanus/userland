package postgres

import (
	"fmt"

	"github.com/spf13/viper"
)

func CreateConnectionString() string {
	return fmt.Sprintf(`
		host=%s 
		port=%d 
		user=%s 
		password=%s 
		dbname=%s 
		sslmode=%s`,
		viper.GetString("postgres_host"),
		viper.GetInt("postgres_port"),
		viper.GetString("postgres_user"),
		viper.GetString("postgres_password"),
		viper.GetString("postgres_dbname"),
		viper.GetString("postgres_sslmode"),
	)
}
