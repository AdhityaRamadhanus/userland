package redis

import (
	"time"

	"github.com/go-redis/redis"
)

//CacheService implements chronicle.CacheService interface using redis
type KeyValueService struct {
	redisClient *redis.Client
}

//NewCacheService construct a new CacheService from redis client
func NewKeyValueService(redisClient *redis.Client) *KeyValueService {
	return &KeyValueService{
		redisClient: redisClient,
	}
}

//Get a cache in bytes from a key
func (c KeyValueService) Get(key string) (result []byte, err error) {
	return c.redisClient.Get(key).Bytes()
}

//Set cache in bytes with key without expiration
func (c KeyValueService) Set(key string, value []byte) (err error) {
	return c.redisClient.Set(key, string(value), 0).Err()
}

//SetEx cache in bytes with key with expiration
func (c KeyValueService) SetEx(key string, value []byte, expiration time.Duration) (err error) {
	return c.redisClient.Set(key, value, expiration).Err()
}
