package profile

import (
	"encoding/base64"

	"github.com/go-errors/errors"

	"fmt"
	"io"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	mailing "github.com/AdhityaRamadhanus/userland/common/http/clients/mailing"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
	log "github.com/sirupsen/logrus"
	qrcode "github.com/skip2/go-qrcode"
)

var (
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

func WithEventRepository(eventRepository userland.EventRepository) func(service *service) {
	return func(service *service) {
		service.eventRepository = eventRepository
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
	ListEvents(user userland.User, pagingOptions userland.EventPagingOptions) (userland.Events, int, error)
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
	eventRepository      userland.EventRepository
	userRepository       userland.UserRepository
	keyValueService      userland.KeyValueService
	objectStorageService userland.ObjectStorageService
}

func (s service) ProfileByEmail(email string) (userland.User, error) {
	return s.userRepository.FindByEmail(email)
}

func (s service) Profile(userID int) (userland.User, error) {
	return s.userRepository.Find(userID)
}

func (s service) SetProfile(user userland.User) error {
	return s.userRepository.Update(user)
}

func (s service) RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error) {
	_, err = s.userRepository.FindByEmail(newEmail)
	if err == nil { // user present
		return "", ErrEmailAlreadyUsed
	}

	verificationID = security.GenerateUUID()
	emailVerificationKey := keygenerator.EmailVerificationKey(user.ID, verificationID)
	s.keyValueService.SetEx(emailVerificationKey, []byte(newEmail), security.EmailVerificationExpiration)

	// call mail service here
	if err := s.mailingClient.SendVerificationEmail(user.Email, user.Fullname, verificationID); err != nil {
		log.WithError(err).Error("Error sending email")
	}
	return verificationID, nil
}

func (s service) ChangeEmail(user userland.User, verificationID string) error {
	verificationKey := keygenerator.EmailVerificationKey(user.ID, verificationID)
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

func (s service) ChangePassword(user userland.User, oldPassword string, newPassword string) error {
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
		return "", "", err
	}

	qrcodeImageBase64 = base64.StdEncoding.EncodeToString(qrCodeImageBytes)
	s.keyValueService.SetEx(keygenerator.TFAActivationKey(user.ID, secret), []byte(code), time.Second*60*5)
	return secret, qrcodeImageBase64, nil
}

func (s service) ActivateTFA(user userland.User, secret string, code string) ([]string, error) {
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
	backupCodes := []string{}
	user.BackupCodes = []string{}
	for i := 0; i < 5; i++ {
		code, _ := security.GenerateOTP(6)
		backupCodes = append(backupCodes, code)
		user.BackupCodes = append(user.BackupCodes, security.HashPassword(code))
	}
	user.TFAEnabled = true
	user.TFAEnabledAt = time.Now()

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

	user.BackupCodes = []string{}
	s.userRepository.StoreBackupCodes(user)
	user.TFAEnabled = false
	return s.userRepository.Update(user)
}

func (s service) DeleteAccount(user userland.User, currPassword string) error {
	if err := security.ComparePassword(user.Password, currPassword); err != nil {
		return ErrWrongPassword
	}

	// should do in background
	// delete user events
	if err := s.eventRepository.DeleteAllByUserID(user.ID); err != nil {
		return err
	}

	return s.userRepository.Delete(user.ID)
}

func (s service) ListEvents(user userland.User, pagingOptions userland.EventPagingOptions) (userland.Events, int, error) {
	return s.eventRepository.FindAllByUserID(user.ID, pagingOptions)
}

func (s service) SetProfilePicture(user userland.User, image io.Reader) error {
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
