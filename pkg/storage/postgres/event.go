package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
func (e EventRepository) FindAll(filter userland.EventFilterOptions, paging userland.EventPagingOptions) (events userland.Events, eventsCount int, err error) {
	scanStructEvents := []EventScanStruct{}

	whereArgs := []interface{}{}
	whereSubStatements := []string{}
	whereStatement := "WHERE %s"
	// build where sub queries, append to whereSubStatements eg: ["user_id=$1", "ip=$2", "event=$3"]
	if filter.UserID > 0 {
		whereArgs = append(whereArgs, filter.UserID)
		whereSubStatements = append(whereSubStatements, fmt.Sprintf("user_id=$%d", len(whereArgs)))
	}

	if len(filter.IP) > 0 {
		whereArgs = append(whereArgs, filter.IP)
		whereSubStatements = append(whereSubStatements, fmt.Sprintf("ip=$%d", len(whereArgs)))
	}

	if len(filter.Event) > 0 {
		whereArgs = append(whereArgs, filter.Event)
		whereSubStatements = append(whereSubStatements, fmt.Sprintf("event=$%d", len(whereArgs)))
	}

	if len(whereArgs) == 0 {
		whereStatement = ""
	} else {
		// build where queries from sub queries,
		// eg: whereSubStatements = ["order_id=$1", "product_id=$2", "name=$3"]
		// whereStatement = "WHERE order_id=$1 AND product_id=$2 AND name=$3"
		whereStatement = fmt.Sprintf(whereStatement, strings.Join(whereSubStatements, "AND "))
	}

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
		%s
		ORDER BY %s %s 
		LIMIT %d 
		OFFSET %d`,
		whereStatement,
		paging.SortBy,
		paging.Order,
		paging.Limit,
		paging.Offset,
	)

	stmt, err := e.db.Preparex(selectQuery)
	if err != nil {
		return userland.Events{}, 0, errors.Wrap(err, "db.Preparex() err")
	}

	if err := stmt.Select(&scanStructEvents, whereArgs...); err != nil {
		return userland.Events{}, 0, errors.Wrap(err, "stmt.Select() err")
	}

	countQuery := fmt.Sprintf(`SELECT count(*) FROM events %s`, whereStatement)
	row := e.db.QueryRow(countQuery, whereArgs...)
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
