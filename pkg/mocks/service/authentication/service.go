package authentication

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/stretchr/testify/mock"
)

type AuthenticationService struct {
	mock.Mock
}

func (m AuthenticationService) Register(user userland.User) error {
	args := m.Called(user)

	return args.Get(0).(error)
}

func (m AuthenticationService) RequestVerification(verificationType string, email string) (verificationID string, err error) {
	args := m.Called(verificationType, email)

	if args.Get(1) == nil {
		return args.Get(0).(string), nil
	}

	return "", args.Get(1).(error)
}

func (m AuthenticationService) VerifyAccount(verificationID string, email string, code string) error {
	args := m.Called(verificationID, email, code)

	return args.Get(0).(error)
}

func (m AuthenticationService) Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error) {
	args := m.Called(email, password)

	if args.Get(2) == nil {
		return args.Get(0).(bool), args.Get(1).(security.AccessToken), nil
	}

	return true, security.AccessToken{}, args.Get(2).(error)
}

func (m AuthenticationService) VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	args := m.Called(tfaToken, userID, code)

	if args.Get(1) == nil {
		return args.Get(0).(security.AccessToken), nil
	}

	return security.AccessToken{}, args.Get(1).(error)
}

func (m AuthenticationService) VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	args := m.Called(tfaToken, userID, code)

	if args.Get(1) == nil {
		return args.Get(0).(security.AccessToken), nil
	}

	return security.AccessToken{}, args.Get(1).(error)
}

func (m AuthenticationService) ForgotPassword(email string) (verificationID string, err error) {
	args := m.Called(email)

	if args.Get(1) == nil {
		return args.Get(0).(string), nil
	}

	return "", args.Get(1).(error)
}

func (m AuthenticationService) ResetPassword(forgotPassToken string, newPassword string) error {
	args := m.Called(forgotPassToken, newPassword)

	return args.Get(0).(error)
}
