package profile

import (
	"errors"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/common/keygenerator"
	"github.com/AdhityaRamadhanus/userland/common/security"
)

var (
	ErrEmailAlreadyUsed = errors.New("Email is already used")

	EmailVerificationExpiration = time.Second * 60 * 2
)

//Service provide an interface to story domain service
type Service interface {
	Profile(userID int) (userland.User, error)
	SetProfile(user userland.User) error
	RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error)
	ChangeEmail(user userland.User, verificationID string) error
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
