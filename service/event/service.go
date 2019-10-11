package event

import (
	"time"

	"github.com/AdhityaRamadhanus/userland"
)

//Service provide an interface to story domain service
type Service interface {
	Log(eventName string, userID int, clientInfo map[string]interface{}) error
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

func (s *service) Log(eventName string, userID int, clientInfo map[string]interface{}) error {
	event := userland.Event{
		UserAgent:  clientInfo["user_agent"].(string),
		UserID:     userID,
		Event:      eventName,
		ClientID:   clientInfo["client_id"].(int),
		ClientName: clientInfo["client_name"].(string),
		IP:         clientInfo["ip"].(string),
		Timestamp:  time.Now(),
	}
	return s.eventRepository.Insert(event)
}
