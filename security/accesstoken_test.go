package security_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/security"
	"github.com/stretchr/testify/assert"
)

func TestGenerateSessionID(t *testing.T) {
	sessionIDs := []string{}
	for i := 0; i < 1000; i++ {
		sessionID := security.GenerateSessionID()
		sessionIDs = append(sessionIDs, sessionID)
	}

	// make sure id unique
	sessionIDTable := map[string]bool{}
	for _, sessionID := range sessionIDs {
		_, duplicate := sessionIDTable[sessionID]
		assert.False(t, duplicate)
		sessionIDTable[sessionID] = true
	}
}

func TestCreateAccessToken(t *testing.T) {
	testCases := []struct {
		User    userland.User
		Options security.AccessTokenOptions
	}{
		{
			User: userland.User{
				Fullname: "Adhitya Ramadhanus",
				Email:    "adhitya.ramadhanus@gmail.com",
				ID:       1,
			},
			Options: security.AccessTokenOptions{
				Expiration: 60 * time.Second * 5,
				Scope: []string{
					"user",
				},
			},
		},
	}

	for _, testCase := range testCases {
		token, err := security.CreateAccessToken(testCase.User, testCase.Options)
		fmt.Println("token ", token.Key, token.Value, token.ExpiredAt)
		assert.Nil(t, err, "Should success create otp")
	}
}
