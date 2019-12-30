package userland

import (
	"time"
)

//Event is domain entity
type Event struct {
	ID         int
	UserID     int
	Event      string
	UserAgent  string
	IP         string
	ClientID   int
	ClientName string
	Timestamp  time.Time
	CreatedAt  time.Time
}

//Events is collection of Event
type Events []Event

type EventFilterOptions struct {
	UserID int
	Event  string
	IP     string
}

//EventPagingOptions is a struct used as pagination option to get events
type EventPagingOptions struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

//EventRepository provide an interface to get user events
type EventRepository interface {
	FindAll(filter EventFilterOptions, paging EventPagingOptions) (Events, int, error)
	Insert(event Event) error
	DeleteAllByUserID(userID int) error
}
