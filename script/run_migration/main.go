package main

import (
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/AdhityaRamadhanus/userland/config"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	config.Init()
	pgConnString := postgres.CreateConnectionString()
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"database": viper.GetString("postgres_dbname"),
		"host":     viper.GetString("postgres_host"),
		"port":     viper.GetString("postgres_port"),
	}).Info("Connected to postgres")

	migrationDir := "storage/postgres/migration"
	migrationFiles := []string{}
	files, _ := ioutil.ReadDir(migrationDir)
	for _, file := range files {
		if !file.IsDir() {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Sort(sort.StringSlice(migrationFiles))
	for _, migrationFile := range migrationFiles {
		log.Info("Running ", migrationFile)
		filePath := path.Join(migrationDir, migrationFile)
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		fileBytes, _ := ioutil.ReadAll(file)
		file.Close()

		sqlQuery := string(fileBytes)
		_, err = db.Queryx(sqlQuery)
		if err != nil {
			log.Fatal("Error running ", migrationFile, err)
		}
	}
}
