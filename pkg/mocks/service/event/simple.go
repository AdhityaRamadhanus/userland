package event

import "github.com/AdhityaRamadhanus/userland"

type SimpleEventService struct {
	CalledMethods map[string]bool
}

func (m SimpleEventService) Log(eventName string, userID int, clientInfo map[string]interface{}) error {
	m.CalledMethods["Log"] = true

	return nil
}

func (m SimpleEventService) ListEvents(filter userland.EventFilterOptions, paging userland.EventPagingOptions) (events userland.Events, count int, err error) {
	m.CalledMethods["ListEvents"] = true

	return userland.Events{}, 0, nil
}

func (m SimpleEventService) DeleteEventsByUserID(userID int) error {
	m.CalledMethods["DeleteEventsByUserID"] = true

	return nil
}
