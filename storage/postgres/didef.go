package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/sarulabs/di"
)

var (
	ConnectionBuilder = di.Def{
		Name:  "postgres-connection",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			pgConnString := CreateConnectionString()
			return sqlx.Open("postgres", pgConnString)
		},
		Close: func(obj interface{}) error {
			return obj.(*sqlx.DB).Close()
		},
	}

	UserRepositoryBuilder = di.Def{
		Name:  "user-repository",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			pgConn := ctn.Get("postgres-connection").(*sqlx.DB)
			userRepository := NewUserRepository(pgConn)
			return userRepository, nil
		},
	}

	EventRepositoryBuilder = di.Def{
		Name:  "event-repository",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			pgConn := ctn.Get("postgres-connection").(*sqlx.DB)
			eventRepository := NewEventRepository(pgConn)
			return eventRepository, nil
		},
	}
)
