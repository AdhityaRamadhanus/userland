package serializers

import "github.com/AdhityaRamadhanus/userland"

func SerializeEventToJSON(event userland.Event) map[string]interface{} {
	return map[string]interface{}{
		"event": event.Event,
		"ua":    event.UserAgent,
		"ip":    event.IP,
		"client": map[string]interface{}{
			"id":   event.ClientID,
			"name": event.ClientName,
		},
		"created_at": event.Timestamp,
	}
}
