package authentication

import (
	"fmt"

	"github.com/AdhityaRamadhanus/userland"
	mailing "github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	EventLogin          = "user.authentication.login"
	EventForgotPassword = "user.authentication.forgot_password"

	ErrUserRegistered        = errors.New("User already registered")
	ErrUserNotVerified       = errors.New("User not verified")
	ErrWrongPassword         = errors.New("Wrong password")
	ErrServiceNotImplemented = errors.New("Service not implemented")
	ErrWrongOTP              = errors.New("Wrong OTP")
	ErrWrongBackupCode       = errors.New("code doesn't match any backup codes")
	ErrOTPInvalid            = errors.New("OTP Invalid")
)

//Service provide an interface to story domain service
type Service interface {
	Register(user userland.User) error
	RequestVerification(verificationType string, email string) (verificationID string, err error)
	VerifyAccount(verificationID string, email string, code string) error
	Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error)
	VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error)
	VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error)
	ForgotPassword(email string) (verificationID string, err error)
	ResetPassword(forgotPassToken string, newPassword string) error
}

func WithUserRepository(userRepository userland.UserRepository) func(service *service) {
	return func(service *service) {
		service.userRepository = userRepository
	}
}

func WithKeyValueService(keyValueService userland.KeyValueService) func(service *service) {
	return func(service *service) {
		service.keyValueService = keyValueService
	}
}

func WithMailingClient(mailingClient mailing.Client) func(service *service) {
	return func(service *service) {
		service.mailingClient = mailingClient
	}
}

func WithConfiguration(cfg *config.Configuration) func(service *service) {
	return func(service *service) {
		service.config = cfg
	}
}

func NewService(options ...func(*service)) Service {
	service := &service{}
	for _, option := range options {
		option(service)
	}

	return service
}

type service struct {
	config          *config.Configuration
	mailingClient   mailing.Client
	userRepository  userland.UserRepository
	keyValueService userland.KeyValueService
}

func (s service) Register(user userland.User) (err error) {
	user.Password = security.HashPassword(user.Password)
	if err := s.userRepository.Insert(&user); err != nil {
		if err == userland.ErrDuplicateKey {
			return ErrUserRegistered
		}
		return err
	}

	return nil
}

func (s service) RequestVerification(verificationType string, email string) (verificationID string, err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return "", err
	}

	switch verificationType {
	case "email.verify":
		// generate code
		code, err := security.GenerateOTP(6)
		if err != nil {
			return "", err
		}
		// create redis key verification
		verificationID := security.GenerateUUID()
		emailVerificationKey := keygenerator.EmailVerificationKey(user.ID, verificationID)
		s.keyValueService.SetEx(emailVerificationKey, []byte(code), security.EmailVerificationExpiration)
		// call mail service here
		// TODO change verificationLink to use mail host via mailing client
		// TODO see if wee need to return error isntead of just logging
		verificationLink := fmt.Sprintf("http://localhost:8000/email_verification?code=%s&key=%s", code, verificationID)
		if err := s.mailingClient.SendVerificationEmail(user.Email, user.Fullname, verificationLink); err != nil {
			log.WithError(err).Error("Error sending email")
		}
		return verificationID, nil
	default:
		return "", ErrServiceNotImplemented
	}
}

func (s service) VerifyAccount(verificationID string, email string, code string) (err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return err
	}

	verificationKey := keygenerator.EmailVerificationKey(user.ID, verificationID)
	expectedCode, err := s.keyValueService.Get(verificationKey)
	if err != nil {
		return err
	}

	if string(expectedCode) != code {
		return ErrWrongOTP
	}

	defer s.keyValueService.Delete(verificationKey)
	user.Verified = true
	return s.userRepository.Update(user)
}

func (s service) loginWithTFA(user userland.User) (accessToken security.AccessToken, err error) {
	code, err := security.GenerateOTP(6)
	if err != nil {
		return security.AccessToken{}, err
	}

	accessToken, err = security.CreateAccessToken(user, s.config.JWTSecret, security.AccessTokenOptions{
		Expiration: security.TFATokenExpiration,
		Scope:      security.TFATokenScope,
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	tfaKey := keygenerator.TFAVerificationKey(user.ID, accessToken.Key)
	s.keyValueService.SetEx(tfaKey, []byte(code), security.TFATokenExpiration)

	tokenKey := keygenerator.TokenKey(accessToken.Key)
	s.keyValueService.SetEx(tokenKey, []byte(accessToken.Value), security.TFATokenExpiration)

	// TODO return error?
	if err := s.mailingClient.SendOTPEmail(user.Email, user.Fullname, "TFA Verification", code); err != nil {
		log.WithError(err).Error("Error sending email")
	}
	return accessToken, nil
}

func (s service) loginNormal(user userland.User) (accessToken security.AccessToken, err error) {
	accessToken, err = security.CreateAccessToken(user, s.config.JWTSecret, security.AccessTokenOptions{
		Expiration: security.UserAccessTokenExpiration,
		Scope:      security.UserTokenScope,
	})
	if err != nil {
		return security.AccessToken{}, err
	}
	return accessToken, nil
}

func (s service) Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return false, security.AccessToken{}, err
	}

	if err = security.ComparePassword(user.Password, password); err != nil {
		return false, security.AccessToken{}, ErrWrongPassword
	}

	// check if verified
	if !user.Verified {
		return false, security.AccessToken{}, ErrUserNotVerified
	}

	if user.TFAEnabled {
		accessToken, err := s.loginWithTFA(user)
		return true, accessToken, err
	}

	accessToken, err = s.loginNormal(user)
	return false, accessToken, err
}

func (s service) VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	// find user
	user, err := s.userRepository.Find(userID)
	if err != nil {
		return security.AccessToken{}, err
	}

	tfaVerificationID := keygenerator.TFAVerificationKey(user.ID, tfaToken)
	tfaTokenKey := keygenerator.TokenKey(tfaToken)
	expectedCode, err := s.keyValueService.Get(tfaVerificationID)
	if err != nil {
		return security.AccessToken{}, err
	}

	// check code
	if string(expectedCode) != code {
		return security.AccessToken{}, ErrWrongOTP
	}

	defer s.keyValueService.Delete(tfaVerificationID)
	defer s.keyValueService.Delete(tfaTokenKey)
	return s.loginNormal(user)
}

func (s service) VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	// find user
	user, err := s.userRepository.Find(userID)
	if err != nil {
		return security.AccessToken{}, err
	}

	codeFound := false
	foundIdx := -1
	for idx, backupCode := range user.BackupCodes {
		if err = security.ComparePassword(backupCode, code); err == nil {
			codeFound = true
			foundIdx = idx
			break
		}
	}

	if !codeFound {
		return security.AccessToken{}, ErrWrongBackupCode
	}

	user.BackupCodes = append(user.BackupCodes[:foundIdx], user.BackupCodes[foundIdx+1:]...)
	s.userRepository.StoreBackupCodes(user)

	tfaVerificationID := keygenerator.TFAVerificationKey(user.ID, tfaToken)
	tfaTokenKey := keygenerator.TokenKey(tfaVerificationID)
	defer s.keyValueService.Delete(tfaVerificationID)
	defer s.keyValueService.Delete(tfaTokenKey)
	return s.loginNormal(user)
}

func (s service) ForgotPassword(email string) (verificationID string, err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return "", err
	}

	verificationID = security.GenerateUUID()
	forgotPassKey := keygenerator.ForgotPasswordKey(verificationID)
	s.keyValueService.SetEx(forgotPassKey, []byte(user.Email), security.ForgotPassExpiration)
	// call mail service
	// TODO return error?
	if err := s.mailingClient.SendOTPEmail(user.Email, user.Fullname, "Forgot Password", verificationID); err != nil {
		log.WithError(err).Error("Error sending email")
	}
	return verificationID, nil
}

func (s service) ResetPassword(forgotPassToken string, newPassword string) (err error) {
	// verify token
	forgotPassKey := keygenerator.ForgotPasswordKey(forgotPassToken)
	email, err := s.keyValueService.Get(forgotPassKey)
	if err != nil {
		return ErrOTPInvalid
	}

	user, err := s.userRepository.FindByEmail(string(email))
	if err != nil {
		return err
	}

	// update password
	user.Password = security.HashPassword(newPassword)
	defer s.keyValueService.Delete(forgotPassKey)
	return s.userRepository.Update(user)
}
