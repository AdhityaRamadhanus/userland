package redis

import (
	"math"
	"time"

	"github.com/AdhityaRamadhanus/userland"

	"github.com/go-redis/redis"
)

//KeyValueService implements userland.KeyValueService interface using redis
type KeyValueService struct {
	redisClient *redis.Client
}

//NewKeyValueService construct a new KeyValueService from redis client
func NewKeyValueService(redisClient *redis.Client) *KeyValueService {
	return &KeyValueService{
		redisClient: redisClient,
	}
}

//Get a cache in bytes from a key
func (c KeyValueService) Get(key string) (result []byte, err error) {
	val, err := c.redisClient.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, userland.ErrKeyNotFound
		}
		return nil, err
	}

	return val, nil
}

//Set cache in bytes with key without expiration
func (c KeyValueService) Set(key string, value []byte) (err error) {
	return c.redisClient.Set(key, string(value), 0).Err()
}

//Set cache in bytes with key without expiration
func (c KeyValueService) Delete(key string) (err error) {
	return c.redisClient.Del(key).Err()
}

//SetEx cache in bytes with key with expiration
func (c KeyValueService) SetEx(key string, value []byte, expiration time.Duration) (err error) {
	return c.redisClient.Set(key, value, expiration).Err()
}

//SetEx cache in bytes with key with expiration
func (c KeyValueService) AddToSortedSet(key string, value string, score float64) (err error) {
	return c.redisClient.ZAdd(key, redis.Z{Score: score, Member: value}).Err()
}

func (c KeyValueService) DeleteFromSortedSet(key string, value string) (err error) {
	return c.redisClient.ZRem(key, value).Err()
}

//SetEx cache in bytes with key with expiration
func (c KeyValueService) GetSortedSet(key string) (result []string, err error) {
	stringSlice := c.redisClient.ZRange(key, math.MinInt64, math.MaxInt64)
	return stringSlice.Result()
}
