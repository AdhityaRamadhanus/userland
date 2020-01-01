package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Configuration provide package level configuration for vendtron by reading from config.yaml first then overwrite if any with env (Default read from .env)
type Configuration struct {
	Env       string         `yaml:"env"`
	API       ApiConfig      `yaml:"api"`
	Mail      MailConfig     `yaml:"mail"`
	Redis     RedisConfig    `yaml:"redis"`
	Postgres  PostgresConfig `yaml:"postgres"`
	Mailjet   MailjetConfig  `yaml:"mailjet"`
	GCP       GCPConfig      `yaml:"gcp"`
	Log       LogConfig      `yaml:"log"`
	JWTSecret string         `yaml:"jwt_secret" envconfig:"JWT_SECRET"`
}

type ApiConfig struct {
	Port int    `yaml:"port" envconfig:"API_PORT"`
	Host string `yaml:"host" envconfig:"API_HOST"`
}

type MailConfig struct {
	Port         int    `yaml:"port" envconfig:"MAIL_PORT"`
	Host         string `yaml:"host" envconfig:"MAIL_HOST"`
	Sender       string `yaml:"sender" envconfig:"MAIL_SENDER"`
	Queue        string `yaml:"queue" envconfig:"MAIL_QUEUE"`
	AuthUser     string `yaml:"auth_user" envconfig:"MAIL_AUTH_USER"`
	AuthPassword string `yaml:"auth_password" envconfig:"MAIL_AUTH_PASSWORD"`
	WorkerSpace  string `yaml:"worker_space" envconfig:"MAIL_WORKER_SPACE"`
}

type RedisConfig struct {
	Port     int    `yaml:"port" envconfig:"REDIS_PORT"`
	Host     string `yaml:"host" envconfig:"REDIS_HOST"`
	Password string `yaml:"password" envconfig:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" envconfig:"REDIS_DB"`
}

type PostgresConfig struct {
	Port     int    `yaml:"port" envconfig:"POSTGRES_PORT"`
	Host     string `yaml:"host" envconfig:"POSTGRES_HOST"`
	User     string `yaml:"user" envconfig:"POSTGRES_USER"`
	Password string `yaml:"password" envconfig:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"dbname" envconfig:"POSTGRES_DBNAME"`
	SSLMode  string `yaml:"sslmode" envconfig:"POSTGRES_SSLMODE"`
}

type GCPConfig struct {
	ApplicationCredentials string `yaml:"application_credentials" envconfig:"GOOGLE_APPLICATION_CRENDETIALS"`
	BucketName             string `yaml:"bucket" envconfig:"GOOGLE_GCS_BUCKET"`
}

type MailjetConfig struct {
	PublicKey  string `yaml:"public_key" envconfig:"MAILJET_PUBLIC_KEY"`
	PrivateKey string `yaml:"private_key" envconfig:"MAILJET_PRIVATE_KEY"`
}

type LogConfig struct {
	Level  string `yaml:"level" envconfig:"LOG_LEVEL"`
	Format string `yaml:"format" envconfig:"LOG_FORMAT"`
}

func Build(yamlPath, envPrefix string) (*Configuration, error) {
	var cfg Configuration
	f, err := os.Open(yamlPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if err := envconfig.Process(envPrefix, &cfg.API); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.API) err")
	}

	if err := envconfig.Process(envPrefix, &cfg.Mail); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.API) err")
	}

	if err := envconfig.Process(envPrefix, &cfg.Redis); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.Redis) err")
	}

	if err := envconfig.Process(envPrefix, &cfg.Postgres); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.Postgres) err")
	}

	if err := envconfig.Process(envPrefix, &cfg.Mailjet); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.Mailjet) err")
	}

	if err := envconfig.Process(envPrefix, &cfg.GCP); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.GCP) err")
	}

	if err := envconfig.Process(envPrefix, &cfg.Log); err != nil {
		return nil, errors.Wrap(err, "envconfig.Process(envPrefix, &cfg.Log) err")
	}

	return &cfg, nil
}
