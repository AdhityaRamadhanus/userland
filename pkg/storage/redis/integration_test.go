// +build integration

package redis_test

import (
	"flag"
	"log"
	"os"
	"testing"

	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

var cfg *config.Configuration

func TestMain(m *testing.M) {
	var envPath string
	var envPrefix string
	var yamlPath string
	flag.StringVar(&envPath, "env-path", ".env", "set env path for test")
	flag.StringVar(&envPrefix, "env-prefix", "TEST", "set env prefix for test")
	flag.StringVar(&yamlPath, "config-yaml", ".config.yaml", "set config.yaml for test")

	flag.Parse()

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("godotenv.Load(%q) err = %v; want nil", envPath, err)
	}
	c, err := config.Build(yamlPath, envPrefix)
	if err != nil {
		log.Fatalf("config.Build(%q, %q) err = %v; want nil", yamlPath, envPrefix, err)
	}

	cfg = c
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestKeyValueService(t *testing.T) {
	suiteTest := NewKeyValueServiceTestSuite(cfg)
	suite.Run(t, suiteTest)
	suiteTest.Teardown()
}

func TestSessionRepository(t *testing.T) {
	suiteTest := NewSessionRepositoryTestSuite(cfg)
	suite.Run(t, suiteTest)
	suiteTest.Teardown()
}
