package profile

import (
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	mailing "github.com/AdhityaRamadhanus/userland/pkg/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/pkg/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
)

var (
	EventChangeEmailRequest = "user.profile.change_email_request"
	EventChangeEmail        = "user.profile.change_email"
	EventChangePassword     = "user.profile.change_password"
	EventEnableTFA          = "user.profile.enable_tfa"
	EventDisableTFA         = "user.profile.disable_tfa"

	ErrEmailAlreadyUsed  = errors.New("Email is already used")
	ErrWrongPassword     = errors.New("Wrong password")
	ErrTFAAlreadyEnabled = errors.New("TFA already enabled")
	ErrWrongOTP          = errors.New("Wrong OTP")
)

func WithMailingClient(mailingClient mailing.Client) func(service *service) {
	return func(service *service) {
		service.mailingClient = mailingClient
	}
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

func WithObjectStorageService(objectStorageService userland.ObjectStorageService) func(service *service) {
	return func(service *service) {
		service.objectStorageService = objectStorageService
	}
}

//Service provide an interface to story domain service
type Service interface {
	ProfileByEmail(email string) (userland.User, error)
	Profile(userID int) (userland.User, error)
	SetProfile(user userland.User) error
	SetProfilePicture(user userland.User, image io.Reader) error
	RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error)
	ChangeEmail(user userland.User, verificationID string) error
	ChangePassword(user userland.User, oldPassword, newPassword string) error
	EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error)
	ActivateTFA(user userland.User, secret string, code string) ([]string, error)
	RemoveTFA(user userland.User, currPassword string) error
	DeleteAccount(user userland.User, currPassword string) error
}

func NewService(options ...func(*service)) Service {
	service := &service{}
	for _, option := range options {
		option(service)
	}

	return service
}

type service struct {
	mailingClient        mailing.Client
	userRepository       userland.UserRepository
	keyValueService      userland.KeyValueService
	objectStorageService userland.ObjectStorageService
}

func (s service) ProfileByEmail(email string) (user userland.User, err error) {
	return s.userRepository.FindByEmail(email)
}

func (s service) Profile(userID int) (user userland.User, err error) {
	return s.userRepository.Find(userID)
}

func (s service) SetProfile(user userland.User) (err error) {
	return s.userRepository.Update(user)
}

func (s service) RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error) {
	if _, err = s.userRepository.FindByEmail(newEmail); err == nil { // user present
		return "", ErrEmailAlreadyUsed
	}

	verificationID = security.GenerateUUID()
	emailVerificationKey := keygenerator.EmailVerificationKey(user.ID, verificationID)
	if err := s.keyValueService.SetEx(emailVerificationKey, []byte(newEmail), security.EmailVerificationExpiration); err != nil {
		return "", errors.New(fmt.Sprintf("keyValueService.SetEx(%q, %s, exp) err", emailVerificationKey, newEmail))
	}

	// call mail service here
	// TODO should return error here?
	if err := s.mailingClient.SendVerificationEmail(newEmail, user.Fullname, verificationID); err != nil {
		log.WithError(err).Error("Error sending email")
	}
	return verificationID, nil
}

func (s service) ChangeEmail(user userland.User, verificationID string) (err error) {
	verificationKey := keygenerator.EmailVerificationKey(user.ID, verificationID)
	newEmail, err := s.keyValueService.Get(verificationKey)
	if err != nil {
		return err
	}

	defer s.keyValueService.Delete(verificationKey)
	user.Email = string(newEmail)
	return s.userRepository.Update(user)
}

func (s service) ChangePassword(user userland.User, oldPassword string, newPassword string) (err error) {
	if err := security.ComparePassword(user.Password, oldPassword); err != nil {
		return ErrWrongPassword
	}

	user.Password = security.HashPassword(newPassword)
	return s.userRepository.Update(user)
}

func (s service) EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error) {
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
		return "", "", errors.Wrap(err, "qrcode.Encode() err")
	}

	qrcodeImageBase64 = base64.StdEncoding.EncodeToString(qrCodeImageBytes)
	s.keyValueService.SetEx(keygenerator.TFAActivationKey(user.ID, secret), []byte(code), time.Second*60*5)
	return secret, qrcodeImageBase64, nil
}

func (s service) ActivateTFA(user userland.User, secret string, code string) (backupCodes []string, err error) {
	if user.TFAEnabled {
		return nil, ErrTFAAlreadyEnabled
	}

	tfaActivationKey := keygenerator.TFAActivationKey(user.ID, secret)
	expectedCode, err := s.keyValueService.Get(tfaActivationKey)
	if err != nil {
		return nil, err
	}

	if string(expectedCode) != code {
		return nil, ErrWrongOTP
	}

	// create 5 backup codes
	backupCodes = []string{}
	user.BackupCodes = []string{}
	for i := 0; i < 5; i++ {
		code, _ := security.GenerateOTP(6)
		backupCodes = append(backupCodes, code)
		user.BackupCodes = append(user.BackupCodes, security.HashPassword(code))
	}
	user.TFAEnabled = true
	user.TFAEnabledAt = time.Now()

	// TODO wrap in transaction
	err = s.userRepository.StoreBackupCodes(user)
	if err != nil {
		return nil, err
	}
	err = s.userRepository.Update(user)

	defer s.keyValueService.Delete(tfaActivationKey)
	return backupCodes, err
}

func (s service) RemoveTFA(user userland.User, currPassword string) error {
	if err := security.ComparePassword(user.Password, currPassword); err != nil {
		return ErrWrongPassword
	}

	// TODO wrap in transaction
	user.BackupCodes = []string{}
	s.userRepository.StoreBackupCodes(user)
	user.TFAEnabled = false
	return s.userRepository.Update(user)
}

func (s service) DeleteAccount(user userland.User, currPassword string) (err error) {
	if err := security.ComparePassword(user.Password, currPassword); err != nil {
		return ErrWrongPassword
	}

	// TODO remove event in handler
	return s.userRepository.Delete(user.ID)
}

func (s service) SetProfilePicture(user userland.User, image io.Reader) (err error) {
	link, err := s.objectStorageService.Write(image, userland.ObjectMetaData{
		CacheControl: "public, max-age=86400",
		ContentType:  "image/jpeg",
		Path:         fmt.Sprintf("userland_%d_profile.jpeg", user.ID),
	})
	if err != nil {
		return err
	}

	user.PictureURL = link
	return s.SetProfile(user)
}
