package userland

import (
	"time"
)

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

type Events []Event

//PagingOptions is a struct used as pagination option to get entities
type EventPagingOptions struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

//UserRepository provide an interface to get user entities
type EventRepository interface {
	FindAllByUserID(userID int, pagingOptions EventPagingOptions) (Events, error)
	Insert(event Event) error
}
