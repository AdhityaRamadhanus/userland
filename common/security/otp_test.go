// +build all common

package security_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland/common/security"
	"github.com/stretchr/testify/assert"
)

func TestGenerateOTP(t *testing.T) {
	testCases := []struct {
		Length      int
		ExpectError bool
	}{
		{
			Length:      6,
			ExpectError: false,
		},
		{
			Length:      2,
			ExpectError: false,
		},
		{
			Length:      4,
			ExpectError: false,
		},
		{
			Length:      -1,
			ExpectError: true,
		},
		{
			Length:      10,
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		code, err := security.GenerateOTP(testCase.Length)
		if testCase.ExpectError {
			assert.NotNil(t, err, "Should return error")
		} else {
			assert.Nil(t, err, "Should success create otp")
			assert.Equal(t, len(code), testCase.Length, "otp length is not correct")
		}
	}
}
