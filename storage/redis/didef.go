package redis

import (
	"fmt"
	"os"

	"github.com/sarulabs/di"

	_redis "github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

var (
	ConnectionBuilder = func(connName string, DB int) di.Def {
		return di.Def{
			Name:  connName,
			Scope: di.App,
			Build: func(ctn di.Container) (interface{}, error) {
				redisAddr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
				redisPass := os.Getenv("REDIS_PASSWORD")
				if os.Getenv("ENV") == "testing" {
					redisAddr = fmt.Sprintf("%s:%s", os.Getenv("TEST_REDIS_HOST"), os.Getenv("TEST_REDIS_PORT"))
					redisPass = os.Getenv("TEST_REDIS_PASSWORD")
				}

				redisClient := _redis.NewClient(&_redis.Options{
					Addr:     redisAddr,
					Password: redisPass, // no password set
					DB:       DB,        // use default DB
				})
				if _, err := redisClient.Ping().Result(); err != nil {
					log.WithError(err).Error("Failed to connect to redis")
					return nil, err
				}
				return redisClient, nil
			},
			Close: func(obj interface{}) error {
				return obj.(*_redis.Client).Close()
			},
		}
	}

	KeyValueServiceBuilder = di.Def{
		Name:  "keyvalue-service",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			redisConn := ctn.Get("redis-connection").(*_redis.Client)
			keyValueService := NewKeyValueService(redisConn)
			return keyValueService, nil
		},
	}

	SessionRepositoryBuilder = di.Def{
		Name:  "session-repository",
		Scope: di.App,
		Build: func(ctn di.Container) (interface{}, error) {
			redisConn := ctn.Get("redis-connection").(*_redis.Client)
			sessionRepository := NewSessionRepository(redisConn)
			return sessionRepository, nil
		},
	}
)
