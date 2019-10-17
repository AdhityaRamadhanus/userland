package authentication

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/security"
)

type SimpleAuthenticationService struct {
	CalledMethods map[string]bool
}

func (m SimpleAuthenticationService) Register(user userland.User) error {
	m.CalledMethods["Register"] = true
	return nil
}

func (m SimpleAuthenticationService) RequestVerification(verificationType string, email string) (verificationID string, err error) {
	m.CalledMethods["RequestVerification"] = true
	return "", nil
}

func (m SimpleAuthenticationService) VerifyAccount(verificationID string, email string, code string) error {
	m.CalledMethods["VerifyAccount"] = true
	return nil
}

func (m SimpleAuthenticationService) Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error) {
	m.CalledMethods["Login"] = true
	return false, security.AccessToken{}, nil
}

func (m SimpleAuthenticationService) VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	m.CalledMethods["VerifyTFA"] = true
	return security.AccessToken{}, nil
}

func (m SimpleAuthenticationService) VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	m.CalledMethods["VerifyTFABypass"] = true
	return security.AccessToken{}, nil
}

func (m SimpleAuthenticationService) ForgotPassword(email string) (verificationID string, err error) {
	m.CalledMethods["ForgotPassword"] = true
	return "", nil
}

func (m SimpleAuthenticationService) ResetPassword(forgotPassToken string, newPassword string) error {
	m.CalledMethods["ResetPassword"] = true
	return nil
}
