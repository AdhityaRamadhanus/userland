package serializers

import "github.com/AdhityaRamadhanus/userland"

func SerializeSessionToJSON(session userland.Session) map[string]interface{} {
	return map[string]interface{}{
		"session_id": session.ID,
		"ip":         session.IP,
		"client": map[string]interface{}{
			"id":   session.ClientID,
			"name": session.ClientName,
		},
		"created_at": session.CreatedAt,
		"updated_at": session.UpdatedAt,
	}
}
