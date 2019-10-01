package story

import (
	"errors"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/security"
)

type AccessToken struct {
	Value     string
	Type      string
	ExpiredAt time.Time
}

//Service provide an interface to story domain service
type Service interface {
	Register(user userland.User) error
	RequestVerification(verificationType string, userID int) error
	Login(email, password string) (requireTFA bool, accessToken AccessToken, err error)
	VerifyTFA(userID int, code string, tfaSource string) (accessToken AccessToken, err error)
	ForgotPassword(email string) error
	ResetPassword(forgotPassToken string, newPassword string) error
}

func NewService(userRepository userland.UserRepository) Service {
	return &service{
		userRepository: userRepository,
	}
}

type service struct {
	userRepository userland.UserRepository
}

func (s *service) Register(user userland.User) error {
	return s.userRepository.Insert(user)
}

func (s *service) RequestVerification(verificationType string, userID int) error {
	switch verificationType {
	case "email.verify":
		// generate code
		_, err := security.GenerateOTP(6)
		if err != nil {
			return err
		}
		// create redis key verification
		// call mail service here
		return nil
	default:
		return errors.New("Service not Implemented")
	}
}

func (s *service) Login(email, password string) (requireTFA bool, accessToken AccessToken, err error) {
	user, err := s.userRepository.FindByEmailAndPassword(email, password)
	if err != nil {
		return false, AccessToken{}, err
	}

	if user.TFAEnabled {
		// create access token for TFA
		requireTFA = true
		// create code
		code, err := security.GenerateOTP(6)
		if err != nil {
			return false, AccessToken{}, err
		}
		return requireTFA, s.createTFAOnlyAccessToken(user, code), nil
	}
	return false, s.createAccessToken(user), nil
}

func (s *service) VerifyTFA(userID int, code string, tfaSource string) (accessToken AccessToken, err error) {
	return AccessToken{}, nil
}

func (s *service) ForgotPassword(email string) error {
	// create forgot password token
	// call mail service
	return nil
}

func (s *service) ResetPassword(forgotPassToken string, newPassword string) error {
	// verify token
	// update password

	var userID int
	user, err := s.userRepository.Find(userID)
	if err != nil {
		return err
	}

	user.Password = newPassword
	return s.userRepository.Update(user)
}

func (s *service) createAccessToken(user userland.User) AccessToken {
	return AccessToken{}
}

func (s *service) createTFAOnlyAccessToken(user userland.User, code string) AccessToken {
	return AccessToken{}
}
