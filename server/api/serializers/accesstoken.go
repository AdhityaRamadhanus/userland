package serializers

import "github.com/AdhityaRamadhanus/userland/common/security"

func SerializeAccessTokenToJSON(accessToken security.AccessToken) map[string]interface{} {
	return map[string]interface{}{
		"value":      accessToken.Key,
		"type":       accessToken.Type,
		"expired_at": accessToken.ExpiredAt,
	}
}
