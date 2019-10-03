package keygenerator

import (
	"fmt"

	"github.com/AdhityaRamadhanus/userland"
)

func EmailVerificationKey(user userland.User, uuid string) string {
	return fmt.Sprintf("email-verification:%d:%s", user.ID, uuid)
}

func TFAVerificationKey(user userland.User, uuid string) string {
	return fmt.Sprintf("tfa-verification:%d:%s", user.ID, uuid)
}

func SessionKey(user userland.User, uuid string) string {
	return fmt.Sprintf("session:%d:%s", user.ID, uuid)
}

func ForgotPasswordKey(uuid string) string {
	return fmt.Sprintf("%s:%s", "forgot-password-token", uuid)
}
