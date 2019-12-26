package event

type SimpleEventService struct {
	CalledMethods map[string]bool
}

func (m SimpleEventService) Log(eventName string, userID int, clientInfo map[string]interface{}) error {
	m.CalledMethods["Log"] = true

	return nil
}
