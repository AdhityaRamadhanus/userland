package userlandtest

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
)

var (
	DefaultUserEmail    = "adhitya.ramadhanus@gmail.com"
	DefaultUserPassword = "test123"
)

func WithUserEmail(email string) func(user *userland.User) {
	return func(user *userland.User) {
		user.Email = email
	}
}
func WithUserPassword(password string) func(user *userland.User) {
	return func(user *userland.User) {
		user.Email = password
	}
}
func Verified(verified bool) func(user *userland.User) {
	return func(user *userland.User) {
		user.Verified = verified
	}
}
func TestCreateUser(t *testing.T, ur userland.UserRepository, opts ...func(user *userland.User)) *userland.User {
	user := &userland.User{
		Email:    DefaultUserEmail,
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword(DefaultUserPassword),
	}

	for _, opt := range opts {
		opt(user)
	}
	if err := ur.Insert(user); err != nil {
		t.Fatalf("Failed to create default user Insert(&user) err = %v; want nil", err)
	}
	return user
}

func TestCreateTFAEnabledUser(t *testing.T, ur userland.UserRepository, opts ...func(user *userland.User)) *userland.User {
	user := &userland.User{
		Email:    DefaultUserEmail,
		Fullname: "Adhitya Ramadhanus",
		Password: security.HashPassword(DefaultUserPassword),
	}

	for _, opt := range opts {
		opt(user)
	}
	if err := ur.Insert(user); err != nil {
		t.Fatalf("Failed to create default user Insert(&user) err = %v; want nil", err)
	}

	user.TFAEnabled = true
	user.TFAEnabledAt = time.Now()

	if err := ur.Update(*user); err != nil {
		t.Fatalf("Failed to update default user Update(user) err = %v; want nil", err)
	}
	return user
}
