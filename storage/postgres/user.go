package postgres

import (
	"github.com/AdhityaRamadhanus/userland"
	"github.com/jmoiron/sqlx"
)

/*
UserRepository is implementation of UserRepository interface
of chronicle domain using postgre
*/
type UserRepository struct {
	db *sqlx.DB
}

//NewUserRepository is constructor to create story repository
func NewUserRepository(conn *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: conn,
	}
}

//Find User by id
func (s UserRepository) Find(id int) (user userland.User, err error) {
	user = userland.User{}
	query := `SELECT
				id,
				email, 
				fullname, 
				phone, 
				location,
				bio,
				weburl,
				pictureurl,
				createdAt, 
				updatedAt
			FROM users 
			WHERE id=$1`
	err = s.db.Get(user, query, id)
	if err != nil {
		return userland.User{}, err
	}

	return user, nil
}

//FindByEmail User by email
func (s UserRepository) FindByEmail(email string) (user userland.User, err error) {
	user = userland.User{}
	query := `SELECT
				id,
				email, 
				fullname, 
				phone, 
				location,
				bio,
				weburl,
				pictureurl,
				createdAt, 
				updatedAt
			FROM users 
			WHERE email=$1`
	err = s.db.Get(user, query, email)
	if err != nil {
		return userland.User{}, err
	}

	return user, nil
}

//Delete delete story by id
func (s UserRepository) Delete(id int) error {
	query := `DELETE FROM users where id=$1`

	deleteStatement, err := s.db.Prepare(query)
	if err != nil {
		return err
	}

	defer deleteStatement.Close()
	_, err = deleteStatement.Exec(id)
	return err
}

//Insert insert story to datastore
func (s UserRepository) Insert(user userland.User) error {
	query := `INSERT INTO users (
				email, 
				fullname, 
				phone,
				location,
				password,
				bio,
				weburl,
				pictureurl,
				tfaenabled, 
				createdAt, 
				updatedAt
			) VALUES (
				:email, 
				:fullname, 
				:phone, 
				:location,
				crypt(:password, gen_salt('bf')),
				:bio,
				:weburl,
				:pictureurl,
				:tfaenabled,
				now(), 
				now()
			) RETURNING id`

	_, err := s.db.NamedQuery(query, user)
	return err
}

//Update update story
func (s UserRepository) Update(user userland.User) error {
	query := `UPDATE users SET (
				email, 
				fullname,
				phone,
				location,
				bio,
				weburl,
				pictureurl,
				tfaenabled,
				updatedAt
			) = (
				:email, 
				:fullname,
				:phone,
				:location,
				:bio,
				:weburl,
				:pictureurl,
				:tfaenabled,
				now()
			) WHERE id=:id`

	_, err := s.db.NamedQuery(query, user)
	return err
}
