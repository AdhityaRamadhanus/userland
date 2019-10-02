package security

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

var (
	clientName = "userland-app"
)

type AccessToken struct {
	Key       string
	Value     string
	Type      string
	ExpiredAt time.Time
}

func generateSessionID() string {
	id := uuid.NewV4()
	return id.String()
}

type AccessTokenOptions struct {
	Expiration time.Duration
	Scope      []string
}

func CreateAccessToken(user userland.User, options AccessTokenOptions) (AccessToken, error) {
	nowInSeconds := time.Now().Unix()

	// generate value token
	claims := jwt.MapClaims{
		"iss":    "userland-api",
		"aud":    clientName,
		"sub":    fmt.Sprintf("userland-access-token|%s|%d", clientName, nowInSeconds),
		"iat":    nowInSeconds,
		"client": clientName,
		"scope":  options.Scope,
	}

	expirationEpoch := nowInSeconds + int64(options.Expiration.Seconds())
	claims["exp"] = expirationEpoch

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return AccessToken{}, err
	}

	log.Println(tokenString)

	// generate session id
	sessionID := generateSessionID()

	return AccessToken{
		Key:       sessionID,
		Value:     tokenString,
		Type:      "Bearer",
		ExpiredAt: time.Unix(expirationEpoch, 0),
	}, nil
}
