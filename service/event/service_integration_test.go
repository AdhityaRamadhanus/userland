// +build all event_service

package event_test

import (
	"os"
	"testing"

	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/jmoiron/sqlx"

	"github.com/sarulabs/di"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/service/event"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

type EventServiceTestSuite struct {
	suite.Suite
	DB              *sqlx.DB
	EventRepository userland.EventRepository
	EventService    event.Service
}

func (suite EventServiceTestSuite) SetupTest() {
	if _, err := suite.DB.Exec("DELETE FROM events"); err != nil {
		log.Fatal("Failed to setup database ", err)
	}

}

func (suite EventServiceTestSuite) BuildContainer() di.Container {
	builder, _ := di.NewBuilder()
	builder.Add(
		postgres.ConnectionBuilder,
		postgres.EventRepositoryBuilder,
		event.ServiceBuilder,
		event.ServiceInstrumentorBuilder,
	)

	return builder.Build()
}

// before each test
func (suite *EventServiceTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")

	ctn := suite.BuildContainer()

	suite.DB = ctn.Get("postgres-connection").(*sqlx.DB)
	suite.EventRepository = ctn.Get("event-repository").(userland.EventRepository)
	suite.EventService = ctn.Get("event-instrumentor-service").(event.Service)
}

func TestEventService(t *testing.T) {
	suiteTest := new(EventServiceTestSuite)
	suite.Run(t, suiteTest)
}

func (suite EventServiceTestSuite) TestLog() {
	testCases := []struct {
		EventName string
		UserID    int
		LogInfo   map[string]interface{}
	}{
		{
			UserID:    1,
			EventName: "authentication.login",
			LogInfo: map[string]interface{}{
				"client_id":   1,
				"user_agent":  "test",
				"client_name": "test",
				"ip":          "test",
			},
		},
	}

	for _, testCase := range testCases {
		err := suite.EventService.Log(testCase.EventName, testCase.UserID, testCase.LogInfo)
		suite.Nil(err, "should create log")
	}
}
