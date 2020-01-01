package security

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

	"github.com/pkg/errors"
)

func GenerateOTP(length int) (string, error) {
	if length < 0 {
		return "", errors.New("OTP Length Must be Positive")
	}
	if length > 8 {
		return "", errors.New("Max OTP Length 8")
	}
	max := int64(math.Pow10(length) - 1)
	bigIntCode, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", errors.Wrapf(err, "rand.Int() err")
	}
	code := fmt.Sprintf("%.*d", length, bigIntCode.Int64())
	return code, nil
}
