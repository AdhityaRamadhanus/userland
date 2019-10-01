package userland

import (
	"time"
)

//User is domain entity
type User struct {
	ID         int
	Email      string
	Fullname   string
	Phone      string
	Location   string
	Bio        string
	WebURL     string
	PictureURL string
	Password   string
	TFAEnabled bool
	Verified   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

//UserRepository provide an interface to get user entities
type UserRepository interface {
	Find(id int) (User, error)
	FindByEmail(email string) (User, error)
	FindByEmailAndPassword(email, password string) (User, error)
	Insert(user User) error
	Update(user User) error
	Delete(id int) error
}
