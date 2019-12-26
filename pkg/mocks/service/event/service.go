package event

import (
	"github.com/stretchr/testify/mock"
)

type EventService struct {
	mock.Mock
}

func (m EventService) Log(eventName string, userID int, clientInfo map[string]interface{}) error {
	args := m.Called(eventName, userID, clientInfo)

	return args.Get(0).(error)
}
