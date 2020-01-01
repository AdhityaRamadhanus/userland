package security

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}

func ComparePassword(hashedPassword string, plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return errors.Wrap(err, "different password")
	}
	return nil
}
