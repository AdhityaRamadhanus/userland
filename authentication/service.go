package authentication

import (
	"errors"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
)

var (
	ErrUserNotVerified       = errors.New("User not verified")
	ErrWrongPassword         = errors.New("Wrong password")
	ErrServiceNotImplemented = errors.New("Service not implemented")
	ErrWrongOTP              = errors.New("Wrong OTP")

	UserAccessTokenExpiration   = time.Second * 60 * 10 // 10 minutes
	TFATokenExpiration          = time.Second * 60 * 2  // 2 minutes
	ForgotPassExpiration        = time.Second * 60 * 5  // 5 minutes
	EmailVerificationExpiration = time.Second * 60 * 2
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

func NewService(userRepository userland.UserRepository, keyValueService userland.KeyValueService) Service {
	return &service{
		userRepository:  userRepository,
		keyValueService: keyValueService,
	}
}

type service struct {
	userRepository  userland.UserRepository
	keyValueService userland.KeyValueService
}

func (s *service) Register(user userland.User) error {
	user.Password = security.HashPassword(user.Password)
	return s.userRepository.Insert(user)
}

func (s *service) RequestVerification(verificationType string, email string) (verificationID string, err error) {
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
		s.keyValueService.SetEx(keygenerator.EmailVerificationKey(user, verificationID), []byte(code), EmailVerificationExpiration)
		// call mail service here
		return verificationID, nil
	default:
		return "", ErrServiceNotImplemented
	}
}

func (s *service) VerifyAccount(verificationID string, email string, code string) error {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return err
	}

	verificationKey := keygenerator.EmailVerificationKey(user, verificationID)
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

func (s *service) loginWithTFA(user userland.User) (accessToken security.AccessToken, err error) {
	code, err := security.GenerateOTP(6)
	if err != nil {
		return security.AccessToken{}, err
	}

	accessToken, err = security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: TFATokenExpiration,
		Scope:      security.TFATokenScope,
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	tfaKey := keygenerator.TFAVerificationKey(user, accessToken.Key)
	s.keyValueService.SetEx(tfaKey, []byte(code), TFATokenExpiration)

	sessionKey := keygenerator.SessionKey(user, accessToken.Key)
	s.keyValueService.SetEx(sessionKey, []byte(accessToken.Value), TFATokenExpiration)
	return accessToken, nil
}

func (s *service) loginNormal(user userland.User) (accessToken security.AccessToken, err error) {
	accessToken, err = security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: UserAccessTokenExpiration,
		Scope:      security.UserTokenScope,
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	sessionKey := keygenerator.SessionKey(user, accessToken.Key)
	s.keyValueService.SetEx(sessionKey, []byte(accessToken.Value), UserAccessTokenExpiration)
	return accessToken, nil
}

func (s *service) Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error) {
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

func (s *service) VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	// find user
	user, err := s.userRepository.Find(userID)
	if err != nil {
		return security.AccessToken{}, err
	}

	tfaKey := keygenerator.TFAVerificationKey(user, tfaToken)
	tfaSessionKey := keygenerator.SessionKey(user, tfaToken)
	expectedCode, err := s.keyValueService.Get(tfaKey)
	if err != nil {
		return security.AccessToken{}, err
	}

	// check code
	if string(expectedCode) != code {
		return security.AccessToken{}, ErrWrongOTP
	}

	defer s.keyValueService.Delete(tfaKey)
	defer s.keyValueService.Delete(tfaSessionKey)
	return s.loginNormal(user)
}

func (s *service) VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	// find user
	user, err := s.userRepository.Find(userID)
	if err != nil {
		return security.AccessToken{}, err
	}

	tfaKey := keygenerator.TFAVerificationKey(user, tfaToken)
	tfaSessionKey := keygenerator.SessionKey(user, tfaToken)
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
		return security.AccessToken{}, errors.New("code doesn't match any backup codes")
	}

	user.BackupCodes = append(user.BackupCodes[:foundIdx], user.BackupCodes[foundIdx+1:]...)
	s.userRepository.StoreBackupCodes(user)

	defer s.keyValueService.Delete(tfaKey)
	defer s.keyValueService.Delete(tfaSessionKey)
	return s.loginNormal(user)
}

func (s *service) ForgotPassword(email string) (verificationID string, err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return "", err
	}

	verificationID = security.GenerateUUID()
	forgotPassKey := keygenerator.ForgotPasswordKey(verificationID)
	s.keyValueService.SetEx(forgotPassKey, []byte(user.Email), ForgotPassExpiration)
	// call mail service
	return verificationID, nil
}

func (s *service) ResetPassword(forgotPassToken string, newPassword string) error {
	// verify token
	forgotPassKey := keygenerator.ForgotPasswordKey(forgotPassToken)
	email, err := s.keyValueService.Get(forgotPassKey)
	if err != nil {
		return err
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
