// +build all event postgres_repository

package postgres_test

import (
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sarulabs/di"
	"github.com/stretchr/testify/suite"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	log "github.com/sirupsen/logrus"
)

type EventRepositoryTestSuite struct {
	suite.Suite
	DB              *sqlx.DB
	EventRepository userland.EventRepository
}

func (suite *EventRepositoryTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		postgres.ConnectionBuilder,
		postgres.EventRepositoryBuilder,
	)

	return builder.Build()
}

func (suite *EventRepositoryTestSuite) SetupTest() {
	_, err := suite.DB.Query("DELETE FROM events")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from events"))
	}
}

func (suite *EventRepositoryTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")

	// Repositories
	ctn := suite.BuildContainer()
	suite.DB = ctn.Get("postgres-connection").(*sqlx.DB)
	suite.EventRepository = ctn.Get("event-repository").(userland.EventRepository)
}

func TestEventRepository(t *testing.T) {
	suiteTest := new(EventRepositoryTestSuite)
	suite.Run(t, suiteTest)
}

func (suite *EventRepositoryTestSuite) TestFindAllByUserID() {
	userID := 1
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES ($1, 'authentication.login', 1, 'userland-app', now(), now())`, userID)
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES ($1, 'authentication.tfa', 1, 'userland-app', now(), now())`, userID)

	testCases := []struct {
		PagingOptions userland.EventPagingOptions
		UserID        int
	}{
		{
			PagingOptions: userland.EventPagingOptions{
				SortBy: "timestamp",
				Offset: 0,
				Limit:  1,
				Order:  "desc",
			},
			UserID: 1,
		},
		{
			PagingOptions: userland.EventPagingOptions{
				SortBy: "timestamp",
				Offset: 0,
				Limit:  2,
				Order:  "desc",
			},
			UserID: 1,
		},
	}

	for _, testCase := range testCases {
		events, count, err := suite.EventRepository.FindAllByUserID(userID, testCase.PagingOptions)
		suite.Nil(err)
		suite.Equal(count, 2)
		suite.Equal(len(events), testCase.PagingOptions.Limit)
	}
}

func (suite *EventRepositoryTestSuite) TestInsert() {
	event := userland.Event{
		UserID:     1,
		Event:      "authentication.login",
		IP:         "127.0.0.1",
		UserAgent:  "Postman",
		ClientName: "userland-app",
		ClientID:   1,
		Timestamp:  time.Now(),
	}
	err := suite.EventRepository.Insert(event)
	suite.Nil(err)
}

func (suite *EventRepositoryTestSuite) TestDeletAllByUserID() {
	userID := 1
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES ($1, 'authentication.login', 1, 'userland-app', now(), now())`, userID)
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES ($1, 'authentication.tfa', 1, 'userland-app', now(), now())`, userID)

	err := suite.EventRepository.DeleteAllByUserID(userID)
	suite.Nil(err)
}
