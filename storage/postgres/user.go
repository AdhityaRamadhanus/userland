package postgres

import (
	"database/sql"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type UserScanStruct struct {
	ID          int
	Email       string
	Fullname    string
	Phone       sql.NullString
	Location    sql.NullString
	Bio         sql.NullString
	WebURL      sql.NullString
	PictureURL  sql.NullString
	Password    string
	TFAEnabled  sql.NullBool
	Verified    sql.NullBool
	BackupCodes pq.StringArray
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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
	userScanStruct := UserScanStruct{}
	query := `SELECT
				id,
				email, 
				fullname, 
				phone, 
				location,
				bio,
				weburl,
				pictureurl,
				verified,
				tfaenabled,
				password,
				backupcodes,
				createdAt, 
				updatedAt
			FROM users 
			WHERE id=$1`
	err = s.db.Get(&userScanStruct, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return userland.User{}, userland.ErrUserNotFound
		}
		return userland.User{}, err
	}

	user = s.convertStructScanToEntity(userScanStruct)
	return user, nil
}

//FindByEmail User by email
func (s UserRepository) FindByEmail(email string) (user userland.User, err error) {
	userScanStruct := UserScanStruct{}
	query := `SELECT
				id,
				email, 
				fullname, 
				phone, 
				location,
				bio,
				weburl,
				pictureurl,
				verified,
				tfaenabled,
				password,
				backupcodes,
				createdAt, 
				updatedAt
			FROM users 
			WHERE email=$1`

	err = s.db.Get(&userScanStruct, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return userland.User{}, userland.ErrUserNotFound
		}
		return userland.User{}, err
	}

	user = s.convertStructScanToEntity(userScanStruct)
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
				:password,
				:bio,
				:weburl,
				:pictureurl,
				:tfaenabled,
				now(), 
				now()
			) RETURNING id`

	nstmt, err := s.db.PrepareNamed(query)
	_, err = nstmt.Query(user)
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
				password,
				pictureurl,
				verified,
				tfaenabled,
				updatedAt
			) = (
				:email, 
				:fullname,
				:phone,
				:location,
				:bio,
				:weburl,
				:password,
				:pictureurl,
				:verified,
				:tfaenabled,
				now()
			) WHERE id=:id`

	_, err := s.db.NamedQuery(query, user)
	return err
}

func (s UserRepository) StoreBackupCodes(user userland.User) error {
	query := `UPDATE users SET (backupcodes, updatedAt) = ($2, now()) WHERE id=$1`
	_, err := s.db.Query(query, user.ID, pq.Array(user.BackupCodes))
	return err
}

func (u UserRepository) convertStructScanToEntity(userScanStruct UserScanStruct) userland.User {
	user := userland.User{
		ID:          userScanStruct.ID,
		Fullname:    userScanStruct.Fullname,
		Email:       userScanStruct.Email,
		Password:    userScanStruct.Password,
		BackupCodes: []string(userScanStruct.BackupCodes),
	}

	if userScanStruct.Phone.Valid {
		user.Phone = userScanStruct.Phone.String
	}
	if userScanStruct.Location.Valid {
		user.Location = userScanStruct.Location.String
	}
	if userScanStruct.Bio.Valid {
		user.Bio = userScanStruct.Bio.String
	}
	if userScanStruct.WebURL.Valid {
		user.WebURL = userScanStruct.WebURL.String
	}
	if userScanStruct.PictureURL.Valid {
		user.PictureURL = userScanStruct.PictureURL.String
	}
	if userScanStruct.TFAEnabled.Valid {
		user.TFAEnabled = userScanStruct.TFAEnabled.Bool
	}
	if userScanStruct.Verified.Valid {
		user.Verified = userScanStruct.Verified.Bool
	}

	return user
}
