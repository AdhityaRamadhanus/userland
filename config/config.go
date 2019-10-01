package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Init() error {
	godotenv.Load()
	viper.AutomaticEnv()
	return viper.ReadInConfig()
}
