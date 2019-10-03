// +build all repository

package redis_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	"github.com/stretchr/testify/assert"

	_redis "github.com/go-redis/redis"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

var (
	keyValueService userland.KeyValueService
)

func Setup(redisClient *_redis.Client) {
	err := redisClient.FlushAll().Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}

	err = redisClient.Set("example1", "value", 0).Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}

	err = redisClient.Set("example2", "value", 0).Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}

	err = redisClient.Set("example3", "value", 0).Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}
}

func TestMain(m *testing.M) {
	godotenv.Load("../../.env")

	redisClient := _redis.NewClient(&_redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.WithError(err).Error("Failed to connect to redis")
	}

	Setup(redisClient)
	keyValueService = redis.NewKeyValueService(redisClient)

	code := m.Run()
	os.Exit(code)
}

func TestKeyValueGet(t *testing.T) {
	testCases := []struct {
		Key           string
		ExpectedValue string
		ExpectError   bool
	}{
		{
			Key:         "example",
			ExpectError: true,
		},
		{
			Key:           "example1",
			ExpectedValue: "value",
			ExpectError:   false,
		},
		{
			Key:           "example2",
			ExpectedValue: "value",
			ExpectError:   false,
		},
		{
			Key:           "example3",
			ExpectedValue: "value",
			ExpectError:   false,
		},
	}

	for _, testCase := range testCases {
		val, err := keyValueService.Get(testCase.Key)
		if testCase.ExpectError {
			assert.NotNil(t, err, "should return error")
		} else {
			assert.Nil(t, err, "should get key")
			assert.Equal(t, testCase.ExpectedValue, string(val))
		}
	}
}

func TestKeyValueSet(t *testing.T) {
	testCases := []struct {
		Key   string
		Value string
	}{
		{
			Key:   "example",
			Value: "value",
		},
		{
			Key:   "example1",
			Value: "value",
		},
	}

	for _, testCase := range testCases {
		err := keyValueService.Set(testCase.Key, []byte(testCase.Value))
		assert.Nil(t, err, "should set key")
	}
}

func TestKeyValueSetEx(t *testing.T) {
	testCases := []struct {
		Key        string
		Value      string
		Expiration time.Duration
	}{
		{
			Key:        "example",
			Value:      "value",
			Expiration: time.Second,
		},
	}

	for _, testCase := range testCases {
		err := keyValueService.SetEx(testCase.Key, []byte(testCase.Value), testCase.Expiration)
		assert.Nil(t, err, "should set key")
		val, err := keyValueService.Get(testCase.Key)
		assert.Nil(t, err, "should get key")
		assert.Equal(t, string(val), testCase.Value, "should get key")

		// wait for duration + 1
		time.Sleep(testCase.Expiration + 1)
		_, err = keyValueService.Get(testCase.Key)
		assert.NotNil(t, err, "should not get key")
	}
}
