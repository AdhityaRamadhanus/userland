package profile

import (
	"io"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/stretchr/testify/mock"
)

type ProfileService struct {
	mock.Mock
}

func (m ProfileService) ProfileByEmail(email string) (userland.User, error) {
	args := m.Called(email)

	if args.Get(1) == nil {
		return args.Get(0).(userland.User), nil
	}

	return userland.User{}, args.Get(1).(error)
}

func (m ProfileService) Profile(userID int) (userland.User, error) {
	args := m.Called(userID)

	if args.Get(1) == nil {
		return args.Get(0).(userland.User), nil
	}

	return userland.User{}, args.Get(1).(error)
}

func (m ProfileService) SetProfile(user userland.User) error {
	args := m.Called(user)

	return args.Get(0).(error)
}

func (m ProfileService) SetProfilePicture(user userland.User, image io.Reader) error {
	args := m.Called(user)

	return args.Get(0).(error)
}

func (m ProfileService) RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error) {
	args := m.Called(user, newEmail)

	if args.Get(1) == nil {
		return args.Get(0).(string), nil
	}

	return "", args.Get(1).(error)
}

func (m ProfileService) ChangeEmail(user userland.User, verificationID string) error {
	args := m.Called(user, verificationID)

	return args.Get(0).(error)
}

func (m ProfileService) ChangePassword(user userland.User, oldPassword, newPassword string) error {
	args := m.Called(user, oldPassword, newPassword)

	return args.Get(0).(error)
}

func (m ProfileService) EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error) {
	args := m.Called(user)

	if args.Get(2) == nil {
		return args.Get(0).(string), args.Get(1).(string), nil
	}

	return "", "", args.Get(2).(error)
}

func (m ProfileService) ActivateTFA(user userland.User, secret string, code string) ([]string, error) {
	args := m.Called(user, secret, code)

	if args.Get(1) == nil {
		return args.Get(0).([]string), nil
	}

	return nil, args.Get(1).(error)
}

func (m ProfileService) RemoveTFA(user userland.User, currPassword string) error {
	args := m.Called(user)

	return args.Get(0).(error)
}

func (m ProfileService) DeleteAccount(user userland.User, currPassword string) error {
	args := m.Called(user)

	return args.Get(0).(error)
}

func (m ProfileService) ListEvents(user userland.User, pagingOptions userland.EventPagingOptions) (userland.Events, int, error) {
	args := m.Called(user, pagingOptions)

	if args.Get(1) == nil {
		return args.Get(0).(userland.Events), args.Get(1).(int), nil
	}

	return nil, 0, args.Get(2).(error)
}
