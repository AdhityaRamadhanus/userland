package userland

import (
	"errors"
	"time"
)

//User is domain entity
type User struct {
	ID          int
	Email       string
	Fullname    string
	Phone       string
	Location    string
	Bio         string
	WebURL      string
	PictureURL  string
	Password    string
	TFAEnabled  bool
	Verified    bool
	BackupCodes []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var (
	ErrUserNotFound = errors.New("User not found")
)

//UserRepository provide an interface to get user entities
type UserRepository interface {
	Find(id int) (User, error)
	FindByEmail(email string) (User, error)
	Insert(user User) error
	Update(user User) error
	// problematic func here
	StoreBackupCodes(user User) error
	Delete(id int) error
}
