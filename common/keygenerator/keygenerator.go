package keygenerator

import (
	"fmt"
)

func EmailVerificationKey(userID int, uuid string) string {
	return fmt.Sprintf("email-verification:%d:%s", userID, uuid)
}

func TFAVerificationKey(userID int, uuid string) string {
	return fmt.Sprintf("tfa-verification:%d:%s", userID, uuid)
}

func TFAActivationKey(userID int, uuid string) string {
	return fmt.Sprintf("tfa-activation:%d:%s", userID, uuid)
}

func SessionKey(uuid string) string {
	return fmt.Sprintf("session:%s", uuid)
}

func SessionListKey(userID int) string {
	return fmt.Sprintf("sessions:%d", userID)
}

func ForgotPasswordKey(uuid string) string {
	return fmt.Sprintf("%s:%s", "forgot-password-token", uuid)
}
