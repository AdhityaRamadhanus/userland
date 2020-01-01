package redis

import (
	"fmt"

	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/go-redis/redis"
	_redis "github.com/go-redis/redis"
	"github.com/pkg/errors"
)

// CreateClient from RedisConfig and DB number
func CreateClient(cfg config.RedisConfig, db int) (*redis.Client, error) {
	redisAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	redisClient := _redis.NewClient(&_redis.Options{
		Addr:     redisAddr,
		Password: cfg.Password, // no password set
		DB:       db,           // use default DB
	})
	if _, err := redisClient.Ping().Result(); err != nil {
		return nil, errors.Wrapf(err, "redisClient.Ping().Result() err;")
	}
	return redisClient, nil
}
