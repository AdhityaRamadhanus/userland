// +build integration

package postgres_test

import (
	"testing"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/postgres"
	"github.com/AdhityaRamadhanus/userland/pkg/userlandtest"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	Config         *config.Configuration
	DB             *sqlx.DB
	UserRepository userland.UserRepository
}

func NewUserRepositoryTestSuite(cfg *config.Configuration) *UserRepositoryTestSuite {
	return &UserRepositoryTestSuite{
		Config: cfg,
	}
}

func (suite *UserRepositoryTestSuite) Teardown() {
	suite.T().Log("Teardown UserRepositoryTestSuite")
	suite.DB.Close()
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.T().Log("Connecting to postgres at", suite.Config.Postgres)
	pgConn, err := postgres.CreateConnection(suite.Config.Postgres)
	if err != nil {
		suite.T().Fatalf("postgres.CreateConnection() err = %v; want nil", err)
	}

	suite.DB = pgConn
	suite.UserRepository = postgres.NewUserRepository(pgConn)
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	query := "DELETE FROM users"
	if _, err := suite.DB.Query(query); err != nil {
		suite.T().Fatalf("suite.DB.Query(%q) err = %v; want nil", query, err)
	}
}

func (suite *UserRepositoryTestSuite) TestInsert() {
	type args struct {
		user userland.User
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "inserted",
			args: args{
				user: userland.User{
					Email:    "adhitya.ramadhanus@gmail.com",
					Fullname: "Adhitya Ramadhanus",
					Password: "test123",
				},
			},
			wantErr: nil,
		},
		{
			name: "failed_duplicate",
			args: args{
				user: userland.User{
					Email:    "adhitya.ramadhanus@gmail.com",
					Fullname: "Adhitya Ramadhanus",
					Password: "test123",
				},
			},
			wantErr: userland.ErrDuplicateKey,
		},
	}
	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.UserRepository.Insert(&tc.args.user)
			if err != tc.wantErr {
				t.Fatalf("UserRepository.Insert(user) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestFindByEmail() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		email string
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "found",
			args: args{
				email: defaultUser.Email,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				email: "adhitya@gmail.com",
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.UserRepository.FindByEmail(tc.args.email)
			if err != tc.wantErr {
				t.Fatalf("UserRepository.FindByEmail(email) err = %v; want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && user.Email != tc.args.email {
				t.Fatalf("UserRepository.FindByEmail(email) User.Email = %s; want %s", user.Email, tc.args.email)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestFindByID() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		userID int
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "found",
			args: args{
				userID: defaultUser.ID,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				userID: defaultUser.ID + 1,
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			user, err := suite.UserRepository.Find(tc.args.userID)
			if err != tc.wantErr {
				t.Fatalf("UserRepository.Find(email) err = %v; want %v", err, tc.wantErr)
			}
			if tc.wantErr == nil && user.ID != tc.args.userID {
				t.Fatalf("UserRepository.Find(email) User.ID = %d; want %d", user.ID, tc.args.userID)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestUpdate() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		user userland.User
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "found",
			args: args{
				user: userland.User{
					ID:         defaultUser.ID,
					Fullname:   "Adhitya Ramadhanus",
					TFAEnabled: true,
					Email:      "adhitya.ramadhanus@gmail.com",
					Phone:      "08123456789",
					Bio:        "Test Update",
				},
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				user: userland.User{
					ID:         defaultUser.ID + 1,
					Fullname:   "Adhitya Ramadhanus",
					TFAEnabled: true,
					Email:      "adhitya@gmail.com",
					Phone:      "08123456789",
					Bio:        "Test Update",
				},
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.UserRepository.Update(tc.args.user)
			if err != tc.wantErr {
				t.Fatalf("UserRepository.Update(user) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestStoreBackupCodes() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		user userland.User
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "found",
			args: args{
				user: userland.User{
					ID:          defaultUser.ID,
					Email:       "adhitya.ramadhanus@gmail.com",
					BackupCodes: []string{"xxx", "xxx"},
				},
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				user: userland.User{
					ID:          defaultUser.ID + 1,
					Email:       "adhitya@gmail.com",
					BackupCodes: []string{"xxx", "xxx"},
				},
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.UserRepository.StoreBackupCodes(tc.args.user)
			if err != tc.wantErr {
				t.Fatalf("UserRepository.StoreBackupCodes(user) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestDelete() {
	defaultUser := userlandtest.TestCreateUser(suite.T(), suite.UserRepository)

	type args struct {
		userID int
	}

	testCases := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "found",
			args: args{
				userID: defaultUser.ID,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			args: args{
				userID: defaultUser.ID + 1,
			},
			wantErr: userland.ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.UserRepository.Delete(tc.args.userID)
			if err != tc.wantErr {
				t.Fatalf("UserRepository.Delete(userID) err = %v; want %v", err, tc.wantErr)
			}
		})
	}
}
