package authentication

import (
	"errors"
	"fmt"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/security"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotVerified = errors.New("User not verified")
	ErrWrongPassword   = errors.New("Wrong password")
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
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
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
		key := fmt.Sprintf("%s:%d:%s", "email-verify", user.ID, verificationID)
		s.keyValueService.SetEx(key, []byte(code), time.Second*60)
		// call mail service here
		return verificationID, nil
	default:
		return "", errors.New("Service not Implemented")
	}
}

func (s *service) VerifyAccount(verificationID string, email string, code string) error {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%d:%s", "email-verify", user.ID, verificationID)
	expectedCode, err := s.keyValueService.Get(key)
	if err != nil {
		return err
	}

	if string(expectedCode) != code {
		return errors.New("OTP doesn't match")
	}

	defer s.keyValueService.Delete(key)
	user.Verified = true
	return s.userRepository.Update(user)
}

func (s *service) Login(email, password string) (requireTFA bool, accessToken security.AccessToken, err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return false, security.AccessToken{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return false, security.AccessToken{}, ErrWrongPassword
	}

	// check if verified
	if !user.Verified {
		return false, security.AccessToken{}, ErrUserNotVerified
	}

	if user.TFAEnabled {
		// create code
		code, err := security.GenerateOTP(6)
		if err != nil {
			return false, security.AccessToken{}, err
		}

		accessToken, err = security.CreateAccessToken(user, security.AccessTokenOptions{
			Expiration: time.Second * 60 * 10,
			Scope:      []string{"tfa"},
		})
		if err != nil {
			return false, security.AccessToken{}, err
		}

		tfaKey := fmt.Sprintf("%s:%d:%s", "tfa-verify", user.ID, accessToken.Key)
		s.keyValueService.SetEx(tfaKey, []byte(code), time.Second*60)

		sessionKey := fmt.Sprintf("session:%d:%s", user.ID, accessToken.Key)
		s.keyValueService.SetEx(sessionKey, []byte(accessToken.Value), time.Second*60*10)
		return true, accessToken, nil
	}

	accessToken, err = security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: time.Second * 60 * 10,
		Scope:      []string{"user"},
	})
	if err != nil {
		return false, security.AccessToken{}, err
	}

	sessionKey := fmt.Sprintf("session:%d:%s", user.ID, accessToken.Key)
	s.keyValueService.SetEx(sessionKey, []byte(accessToken.Value), time.Second*60*10)
	return false, accessToken, nil
}

func (s *service) VerifyTFA(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	// find user
	tfaKey := fmt.Sprintf("%s:%d:%s", "tfa-verify", userID, tfaToken)
	tfaSessionKey := fmt.Sprintf("session:%d:%s", userID, tfaToken)
	expectedCode, err := s.keyValueService.Get(tfaKey)
	if err != nil {
		return security.AccessToken{}, err
	}

	// check code
	if string(expectedCode) != code {
		return security.AccessToken{}, errors.New("OTP doesn't match")
	}

	user, err := s.userRepository.Find(userID)
	if err != nil {
		return security.AccessToken{}, err
	}
	accessToken, err = security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: time.Second * 60 * 10,
		Scope:      []string{"user"},
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	defer s.keyValueService.Delete(tfaKey)
	defer s.keyValueService.Delete(tfaSessionKey)

	sessionKey := fmt.Sprintf("session:%d:%s", user.ID, accessToken.Key)
	s.keyValueService.SetEx(sessionKey, []byte(accessToken.Value), time.Second*60*10)
	return accessToken, nil
}

func (s *service) VerifyTFABypass(tfaToken string, userID int, code string) (accessToken security.AccessToken, err error) {
	// find user
	tfaKey := fmt.Sprintf("%s:%d:%s", "tfa-verify", userID, tfaToken)
	tfaSessionKey := fmt.Sprintf("session:%d:%s", userID, tfaToken)

	user, err := s.userRepository.Find(userID)
	if err != nil {
		return security.AccessToken{}, err
	}

	codeFound := false
	foundIdx := -1
	for idx, backupCode := range user.BackupCodes {
		err = bcrypt.CompareHashAndPassword([]byte(backupCode), []byte(code))
		if err == nil {
			codeFound = true
			foundIdx = idx
			break
		}
	}

	if !codeFound {
		return security.AccessToken{}, errors.New("code doesn't match any backup codes")
	}

	accessToken, err = security.CreateAccessToken(user, security.AccessTokenOptions{
		Expiration: time.Second * 60 * 10,
		Scope:      []string{"user"},
	})
	if err != nil {
		return security.AccessToken{}, err
	}

	user.BackupCodes = append(user.BackupCodes[:foundIdx], user.BackupCodes[foundIdx+1:]...)
	s.userRepository.StoreBackupCodes(user)

	defer s.keyValueService.Delete(tfaKey)
	defer s.keyValueService.Delete(tfaSessionKey)

	sessionKey := fmt.Sprintf("session:%d:%s", user.ID, accessToken.Key)
	s.keyValueService.SetEx(sessionKey, []byte(accessToken.Value), time.Second*60*10)
	return accessToken, nil
}

func (s *service) ForgotPassword(email string) (verificationID string, err error) {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return "", err
	}

	verificationID = security.GenerateUUID()
	forgotPassKey := fmt.Sprintf("%s:%s", "forgot-password-token", verificationID)
	s.keyValueService.SetEx(forgotPassKey, []byte(user.Email), time.Second*60)
	// call mail service
	return verificationID, nil
}

func (s *service) ResetPassword(forgotPassToken string, newPassword string) error {
	// verify token
	key := fmt.Sprintf("%s:%s", "forgot-password-token", forgotPassToken)
	email, err := s.keyValueService.Get(key)
	if err != nil {
		return err
	}

	user, err := s.userRepository.FindByEmail(string(email))
	if err != nil {
		return err
	}

	// update password
	user.Password = newPassword
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	defer s.keyValueService.Delete(key)
	user.Password = string(hash)
	return s.userRepository.Update(user)
}
