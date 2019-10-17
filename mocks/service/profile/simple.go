package profile

import (
	"io"

	"github.com/AdhityaRamadhanus/userland"
)

type SimpleProfileService struct {
	CalledMethods map[string]bool
}

func (m SimpleProfileService) ProfileByEmail(email string) (userland.User, error) {
	m.CalledMethods["ProfileByEmail"] = true
	return userland.User{}, nil
}

func (m SimpleProfileService) Profile(userID int) (userland.User, error) {
	m.CalledMethods["Profile"] = true
	return userland.User{}, nil
}

func (m SimpleProfileService) SetProfile(user userland.User) error {
	m.CalledMethods["SetProfile"] = true
	return nil
}

func (m SimpleProfileService) SetProfilePicture(user userland.User, image io.Reader) error {
	m.CalledMethods["SetProfilePicture"] = true
	return nil
}

func (m SimpleProfileService) RequestChangeEmail(user userland.User, newEmail string) (verificationID string, err error) {
	m.CalledMethods["RequestChangeEmail"] = true
	return "", nil
}

func (m SimpleProfileService) ChangeEmail(user userland.User, verificationID string) error {
	m.CalledMethods["ChangeEmail"] = true
	return nil
}

func (m SimpleProfileService) ChangePassword(user userland.User, oldPassword, newPassword string) error {
	m.CalledMethods["ChangePassword"] = true
	return nil
}

func (m SimpleProfileService) EnrollTFA(user userland.User) (secret string, qrcodeImageBase64 string, err error) {
	m.CalledMethods["EnrollTFA"] = true
	return "", "", nil
}

func (m SimpleProfileService) ActivateTFA(user userland.User, secret string, code string) ([]string, error) {
	m.CalledMethods["ActivateTFA"] = true
	return []string{}, nil
}

func (m SimpleProfileService) RemoveTFA(user userland.User, currPassword string) error {
	m.CalledMethods["RemoveTFA"] = true
	return nil
}

func (m SimpleProfileService) DeleteAccount(user userland.User, currPassword string) error {
	m.CalledMethods["DeleteAccount"] = true
	return nil
}

func (m SimpleProfileService) ListEvents(user userland.User, pagingOptions userland.EventPagingOptions) (userland.Events, int, error) {
	m.CalledMethods["ListEvents"] = true
	return userland.Events{}, 0, nil
}
