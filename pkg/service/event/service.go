package event

import (
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/pkg/errors"
)

var (
	ErrInvalidEvent = errors.New("Invalid Event")
)

//Service provide an interface to story domain service
type Service interface {
	Log(eventName string, userID int, clientInfo map[string]interface{}) error
	ListEvents(filter userland.EventFilterOptions, paging userland.EventPagingOptions) (events userland.Events, count int, err error)
}

func WithEventRepository(eventRepository userland.EventRepository) func(service *service) {
	return func(service *service) {
		service.eventRepository = eventRepository
	}
}

func NewService(options ...func(*service)) Service {
	service := &service{}
	for _, option := range options {
		option(service)
	}

	return service
}

type service struct {
	eventRepository userland.EventRepository
}

func (s service) Log(eventName string, userID int, clientInfo map[string]interface{}) (err error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = errors.Wrapf(ErrInvalidEvent, "Error in inserting event %s", panicErr)
		}
	}()

	// may panic on assertion
	event := userland.Event{
		UserAgent:  clientInfo["user_agent"].(string),
		UserID:     userID,
		Event:      eventName,
		ClientID:   clientInfo["client_id"].(int),
		ClientName: clientInfo["client_name"].(string),
		IP:         clientInfo["ip"].(string),
		Timestamp:  time.Now(),
	}
	if err := s.eventRepository.Insert(event); err != nil {
		return err
	}

	return nil
}

func (s service) ListEvents(filter userland.EventFilterOptions, paging userland.EventPagingOptions) (events userland.Events, count int, err error) {
	return s.eventRepository.FindAll(filter, paging)
}
