package profile

import (
	"encoding/base64"
	"errors"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
	qrcode "github.com/skip2/go-qrcode"
)

var (
	ErrEmailAlreadyUsed  = errors.New("Email is already used")
	ErrWrongPassword     = errors.New("Wrong password")
	ErrTFAAlreadyEnabled = errors.New("TFA already enabled")
	ErrWrongOTP          = errors.New("Wrong OTP")

	EmailVerificationExpiration = time.Second * 60 * 2
)

//Service provide an interface to story domain service
type Service interface {
	Profile(userID int) (userland.User, error)
	SetProfile(user userland.User) error
	RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error)
	ChangeEmail(user userland.User, verificationID string) error
	ChangePassword(user userland.User, oldPassword, newPassword string) error
	EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error)
	ActivateTFA(user userland.User, secret string, code string) error
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

func (s *service) Profile(userID int) (userland.User, error) {
	return s.userRepository.Find(userID)
}

func (s *service) SetProfile(user userland.User) error {
	return s.userRepository.Update(user)
}

func (s *service) RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error) {
	_, err = s.userRepository.FindByEmail(newEmail)
	if err == nil { // user present
		return "", ErrEmailAlreadyUsed
	}

	verificationID = security.GenerateUUID()
	s.keyValueService.SetEx(keygenerator.EmailVerificationKey(user, verificationID), []byte(newEmail), EmailVerificationExpiration)

	// call mail service here
	return verificationID, nil
}

func (s *service) ChangeEmail(user userland.User, verificationID string) error {
	verificationKey := keygenerator.EmailVerificationKey(user, verificationID)
	newEmail, err := s.keyValueService.Get(verificationKey)
	if err != nil {
		return err
	}

	_, err = s.userRepository.FindByEmail(string(newEmail))
	if err == nil { // user present
		return ErrEmailAlreadyUsed
	}

	defer s.keyValueService.Delete(verificationKey)
	user.Email = string(newEmail)
	return s.userRepository.Update(user)
}

func (s *service) ChangePassword(user userland.User, oldPassword string, newPassword string) error {
	if err := security.ComparePassword(user.Password, oldPassword); err != nil {
		return ErrWrongPassword
	}

	user.Password = security.HashPassword(newPassword)
	return s.userRepository.Update(user)
}

func (s *service) EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error) {
	if user.TFAEnabled {
		return "", "", ErrTFAAlreadyEnabled
	}

	secret = security.GenerateUUID()
	code, err := security.GenerateOTP(6)
	if err != nil {
		return "", "", err
	}

	var qrCodeImageBytes []byte
	qrCodeImageBytes, err = qrcode.Encode(code, qrcode.Medium, 256)
	if err != nil {
		return "", "", err
	}

	qrcodeImageBase64 = base64.StdEncoding.EncodeToString(qrCodeImageBytes)
	s.keyValueService.SetEx(keygenerator.TFAActivationKey(user, secret), []byte(code), time.Second*60*5)
	return secret, qrcodeImageBase64, nil
}

func (s *service) ActivateTFA(user userland.User, secret string, code string) error {
	if user.TFAEnabled {
		return ErrTFAAlreadyEnabled
	}

	tfaActivationKey := keygenerator.TFAActivationKey(user, secret)
	expectedCode, err := s.keyValueService.Get(tfaActivationKey)
	if err != nil {
		return err
	}

	if string(expectedCode) != code {
		return ErrWrongOTP
	}

	user.TFAEnabled = true
	user.TFAEnabledAt = time.Now()
	defer s.keyValueService.Delete(tfaActivationKey)
	return s.userRepository.Update(user)
}
