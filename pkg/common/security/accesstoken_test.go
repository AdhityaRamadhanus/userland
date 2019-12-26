// +build all common unit

package security_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/stretchr/testify/assert"
)

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
				Scope:      security.UserTokenScope,
			},
		},
	}

	for _, testCase := range testCases {
		_, err := security.CreateAccessToken(testCase.User, testCase.Options)
		assert.Nil(t, err, "Should success create otp")
	}
}
