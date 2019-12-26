package serializers

import "github.com/AdhityaRamadhanus/userland"

func SerializeUserToJSON(user userland.User) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.ID,
		"fullname":   user.Fullname,
		"location":   user.Location,
		"bio":        user.Bio,
		"web":        user.WebURL,
		"picture":    user.PictureURL,
		"created_at": user.CreatedAt,
	}
}
