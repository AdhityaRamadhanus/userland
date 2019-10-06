package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/jmoiron/sqlx"
)

type EventScanStruct struct {
	ID         int
	UserID     int `db:"user_id"`
	Event      string
	UserAgent  sql.NullString `db:"user_agent"`
	IP         sql.NullString
	ClientID   int    `db:"client_id"`
	ClientName string `db:"client_name"`
	Timestamp  time.Time
	CreatedAt  time.Time `db:"created_at"`
}

/*
EventRepository is implementation of EventRepository interface
of userland domain using postgre
*/
type EventRepository struct {
	db *sqlx.DB
}

//NewEventRepository is constructor to create story repository
func NewEventRepository(conn *sqlx.DB) *EventRepository {
	return &EventRepository{
		db: conn,
	}
}

//Find User by id
func (e EventRepository) FindAllByUserID(userID int, options userland.EventPagingOptions) (events userland.Events, eventsCount int, err error) {
	scanStructEvents := []EventScanStruct{}
	selectQuery := fmt.Sprintf(
		`SELECT
			id,
			user_id, 
			event, 
			user_agent,
			ip,
			client_id,
			client_name,
			timestamp,
			created_at
		FROM events 
		WHERE user_id=$1
		ORDER BY %s %s 
		LIMIT %d 
		OFFSET %d`,
		options.SortBy,
		options.Order,
		options.Limit,
		options.Offset,
	)
	err = e.db.Select(&scanStructEvents, selectQuery, userID)
	if err != nil {
		return userland.Events{}, 0, err
	}

	countQuery := `SELECT count(*) FROM events WHERE user_id=$1`
	row := e.db.QueryRow(countQuery, userID)
	row.Scan(&eventsCount)

	events = userland.Events{}
	for _, scanStructEvent := range scanStructEvents {
		events = append(events, e.convertStructScanToEntity(scanStructEvent))
	}
	return events, eventsCount, nil
}

func (s EventRepository) DeleteAllByUserID(userID int) error {
	query := `DELETE FROM events where user_id=$1`

	deleteStatement, err := s.db.Prepare(query)
	if err != nil {
		return err
	}

	defer deleteStatement.Close()
	_, err = deleteStatement.Exec(userID)
	return err
}

func (e EventRepository) Insert(event userland.Event) error {
	query := `INSERT INTO events (
				user_id,
				event, 
				user_agent,
				ip,
				client_id,
				client_name,
				timestamp,
				created_at
			) VALUES (
				:userid, 
				:event, 
				:useragent, 
				:ip,
				:clientid,
				:clientname,
				:timestamp,
				now()
			)`

	_, err := e.db.NamedQuery(query, event)
	return err
}

func (e EventRepository) convertStructScanToEntity(eventScanStruct EventScanStruct) userland.Event {
	event := userland.Event{
		ID:         eventScanStruct.ID,
		UserID:     eventScanStruct.UserID,
		Event:      eventScanStruct.Event,
		ClientID:   eventScanStruct.ClientID,
		ClientName: eventScanStruct.ClientName,
		Timestamp:  eventScanStruct.Timestamp,
		CreatedAt:  eventScanStruct.CreatedAt,
	}

	if eventScanStruct.IP.Valid {
		event.IP = eventScanStruct.IP.String
	}
	if eventScanStruct.UserAgent.Valid {
		event.UserAgent = eventScanStruct.UserAgent.String
	}

	return event
}
