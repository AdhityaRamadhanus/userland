// +build all event_service

package event_test

import (
	"os"
	"testing"

	"github.com/AdhityaRamadhanus/userland/metrics"
	"github.com/AdhityaRamadhanus/userland/storage/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/AdhityaRamadhanus/userland/service/event"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

// keyValueService   userland.KeyValueService
// sessionRepository userland.SessionRepository
// userRepository    userland.UserRepository

type EventServiceTestSuite struct {
	suite.Suite
	DB              *sqlx.DB
	EventRepository *postgres.EventRepository
	EventService    event.Service
}

func (suite *EventServiceTestSuite) SetupTest() {
	_, err := suite.DB.Exec("DELETE FROM events")
	if err != nil {
		log.Fatal("Failed to setup database ", errors.Wrap(err, "Failed in delete from events"))
	}

}

// before each test
func (suite *EventServiceTestSuite) SetupSuite() {
	godotenv.Load("../../.env")
	os.Setenv("ENV", "testing")
	pgConnString := postgres.CreateConnectionString()
	db, err := sqlx.Open("postgres", pgConnString)
	if err != nil {
		log.Fatal(err)
	}

	eventRepository := postgres.NewEventRepository(db)

	eventService := event.NewService(
		event.WithEventRepository(eventRepository),
	)
	eventService = event.NewInstrumentorService(
		metrics.PrometheusRequestCounter("service", "event", event.MetricKeys),
		metrics.PrometheusRequestLatency("service", "event", event.MetricKeys),
		eventService,
	)

	suite.DB = db
	suite.EventRepository = eventRepository
	suite.EventService = eventService
}

func TestEventService(t *testing.T) {
	suiteTest := new(EventServiceTestSuite)
	suite.Run(t, suiteTest)
}

// CreateSession(userID int, session userland.Session) error
// ListSession(userID int) (userland.Sessions, error)
// EndSession(userID int, currentSessionID string) error
// EndOtherSessions(userID int, currentSessionID string) error
// CreateRefreshToken(user userland.User, currentSessionID string) (security.AccessToken, error)
// CreateNewAccessToken(user userland.User, refreshTokenID string) (security.AccessToken, error)

func (suite *EventServiceTestSuite) TestLog() {
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
