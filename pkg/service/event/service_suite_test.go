// +build integration

package event_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/metrics"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/service/event"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type EventServiceTestSuite struct {
	suite.Suite
	Config          *config.Configuration
	DB              *sqlx.DB
	EventRepository userland.EventRepository
	EventService    event.Service
}

func NewEventServiceTestSuite(cfg *config.Configuration) *EventServiceTestSuite {
	return &EventServiceTestSuite{
		Config: cfg,
	}
}

// before each test
func (suite *EventServiceTestSuite) SetupSuite() {
	suite.T().Log("Connecting to postgres at", suite.Config.Postgres)
	pgConn, err := postgres.CreateConnection(suite.Config.Postgres)
	if err != nil {
		suite.T().Fatalf("postgres.CreateConnection() err = %v", err)
	}

	suite.DB = pgConn
	suite.EventRepository = postgres.NewEventRepository(pgConn)
	suite.EventService = event.NewService(event.WithEventRepository(suite.EventRepository))
	suite.EventService = event.NewInstrumentorService(metrics.PrometheusRequestLatency("service", "event", event.MetricKeys), suite.EventService)
}

func (suite EventServiceTestSuite) SetupTest() {
	queries := []string{
		"DELETE FROM users",
		"DELETE FROM events",
	}

	for _, query := range queries {
		if _, err := suite.DB.Query(query); err != nil {
			suite.T().Fatalf("DB.Query(%q) err = %v; want nil", query, err)
		}
	}
}

func (suite EventServiceTestSuite) TestLog() {
	type args struct {
		name   string
		info   map[string]interface{}
		userID int
	}
	testCases := []struct {
		name         string
		args         args
		wantErrCause error
	}{
		{
			name: "success",
			args: args{
				userID: 1,
				name:   "something.something",
				info: map[string]interface{}{
					"client_id":   1,
					"user_agent":  "test",
					"client_name": "test",
					"ip":          "test",
				},
			},
			wantErrCause: nil,
		},
		{
			name: "error in client_id",
			args: args{
				userID: 1,
				name:   "something.something",
				info: map[string]interface{}{
					"client_id":   "this should be int",
					"user_agent":  "test",
					"client_name": "test",
					"ip":          "test",
				},
			},
			wantErrCause: event.ErrInvalidEvent,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if err := suite.EventService.Log(tc.args.name, tc.args.userID, tc.args.info); errors.Cause(err) != tc.wantErrCause {
				t.Fatalf("EventService.Log(%q, %d, <info>) err = %v; want nil", tc.args.name, tc.args.userID, err)
			}
		})
	}
}
