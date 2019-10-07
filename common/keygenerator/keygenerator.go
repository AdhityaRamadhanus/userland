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

func TFAActivationKey(user userland.User, uuid string) string {
	return fmt.Sprintf("tfa-activation:%d:%s", user.ID, uuid)
}

func SessionKey(uuid string) string {
	return fmt.Sprintf("session:%s", uuid)
}

func ForgotPasswordKey(uuid string) string {
	return fmt.Sprintf("%s:%s", "forgot-password-token", uuid)
}
