package security

import (
	"fmt"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/dgrijalva/jwt-go"
)

var (
	clientName = "userland-app"

	TFATokenScope     = "tfa"
	UserTokenScope    = "user"
	RefreshTokenScope = "refresh"
)

type AccessToken struct {
	Key       string
	Value     string
	Type      string
	ExpiredAt time.Time
}

type AccessTokenOptions struct {
	Expiration  time.Duration
	Scope       string
	CustomClaim map[string]interface{}
}

func CreateAccessToken(user userland.User, jwtSecret string, options AccessTokenOptions) (AccessToken, error) {
	nowInSeconds := time.Now().Unix()

	// generate value token
	claims := jwt.MapClaims{
		"iss":      "userland-api",
		"aud":      clientName,
		"sub":      fmt.Sprintf("userland-access-token|%s|%d", clientName, nowInSeconds),
		"iat":      nowInSeconds,
		"client":   clientName,
		"scope":    options.Scope,
		"fullname": user.Fullname,
		"email":    user.Email,
		"userid":   user.ID,
	}

	if len(options.CustomClaim) > 0 {
		for key, value := range options.CustomClaim {
			claims[key] = value
		}
	}

	expirationEpoch := nowInSeconds + int64(options.Expiration.Seconds())
	claims["exp"] = expirationEpoch

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return AccessToken{}, err
	}

	// generate session id
	sessionID := GenerateUUID()

	return AccessToken{
		Key:       sessionID,
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: time.Unix(expirationEpoch, 0),
	}, nil
}
