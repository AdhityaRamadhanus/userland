// +build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type EventRepositoryTestSuite struct {
	suite.Suite
	Config          *config.Configuration
	DB              *sqlx.DB
	EventRepository userland.EventRepository
}

func NewEventRepositoryTestSuite(cfg *config.Configuration) *EventRepositoryTestSuite {
	return &EventRepositoryTestSuite{
		Config: cfg,
	}
}

func (suite *EventRepositoryTestSuite) SetupSuite() {
	suite.T().Log("Connecting to postgres at", suite.Config.Postgres)
	pgConn, err := postgres.CreateConnection(suite.Config.Postgres)
	if err != nil {
		suite.T().Fatalf("postgres.CreateConnection() err = %v; want nil", err)
	}

	suite.DB = pgConn
	suite.EventRepository = postgres.NewEventRepository(pgConn)
}

func (suite *EventRepositoryTestSuite) SetupTest() {
	query := "DELETE FROM events"
	if _, err := suite.DB.Query(query); err != nil {
		suite.T().Fatalf("suite.DB.Query(%q) err = %v; want nil", query, err)
	}
}

func (suite *EventRepositoryTestSuite) TestFindAllByUserID() {
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES (1, 'authentication.login', 1, 'userland-app', now(), now())`)
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES (1, 'authentication.tfa', 1, 'userland-app', now(), now())`)

	type args struct {
		userID int
		paging userland.EventPagingOptions
	}
	testCases := []struct {
		name           string
		args           args
		wantTotalCount int
		wantCount      int
	}{
		{
			name: "limit=1,user_id=1",
			args: args{
				paging: userland.EventPagingOptions{
					SortBy: "timestamp",
					Offset: 0,
					Limit:  1,
					Order:  "desc",
				},
				userID: 1,
			},
			wantTotalCount: 2,
			wantCount:      1,
		},
		{
			name: "limit=2,user_id=1",
			args: args{
				paging: userland.EventPagingOptions{
					SortBy: "timestamp",
					Offset: 0,
					Limit:  2,
					Order:  "desc",
				},
				userID: 1,
			},
			wantTotalCount: 2,
			wantCount:      2,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			events, count, err := suite.EventRepository.FindAllByUserID(tc.args.userID, tc.args.paging)
			if err != nil {
				t.Fatalf("EventRepository.FindAllByUserID() err = %v; want nil", err)
			}

			gotEventsCount := len(events)
			if gotEventsCount != tc.wantCount {
				t.Errorf("EventRepository.FindAllByUserID() len(events) = %d; want %d", gotEventsCount, tc.wantCount)
			}

			if count != tc.wantTotalCount {
				t.Errorf("EventRepository.FindAllByUserID() totalCount = %d; want %d", count, tc.wantTotalCount)
			}
		})
	}
}

func (suite *EventRepositoryTestSuite) TestInsert() {
	type args struct {
		event userland.Event
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				event: userland.Event{
					UserID:     1,
					Event:      "authentication.login",
					IP:         "127.0.0.1",
					UserAgent:  "Postman",
					ClientName: "userland-app",
					ClientID:   1,
					Timestamp:  time.Now(),
				},
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if err := suite.EventRepository.Insert(tc.args.event); err != tc.wantErr {
				t.Errorf("EventRepository.Insert(event) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite *EventRepositoryTestSuite) TestDeleteAllByUserID() {
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES (1, 'authentication.login', 1, 'userland-app', now(), now())`)
	suite.DB.QueryRow(`INSERT INTO events (user_id, event, client_id, client_name, timestamp, created_at) VALUES (1, 'authentication.tfa', 1, 'userland-app', now(), now())`)

	type args struct {
		userID int
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "success",
			args: args{
				userID: 1,
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			if err := suite.EventRepository.DeleteAllByUserID(tc.args.userID); err != tc.wantErr {
				t.Errorf("EventRepository.DeleteAllByUserID(userID) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}
