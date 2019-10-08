// +build all repository

package redis_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/storage/redis"
	"github.com/stretchr/testify/suite"

	_redis "github.com/go-redis/redis"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

type KeyValueServiceTestSuite struct {
	suite.Suite
	RedisClient     *_redis.Client
	KeyValueService userland.KeyValueService
}

func (suite *KeyValueServiceTestSuite) SetupTest() {
	err := suite.RedisClient.FlushAll().Err()
	if err != nil {
		log.Fatal("Cannot setup redis")
	}
}

func (suite *KeyValueServiceTestSuite) SetupSuite() {
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

	keyValueService := redis.NewKeyValueService(redisClient)

	suite.RedisClient = redisClient
	suite.KeyValueService = keyValueService
}

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

func TestKeyValueService(t *testing.T) {
	suiteTest := new(KeyValueServiceTestSuite)
	suite.Run(t, suiteTest)
}

func (suite *KeyValueServiceTestSuite) TestKeyValueGet() {
	suite.RedisClient.Set("example1", "value", 0)
	suite.RedisClient.Set("example2", "value", 0)
	suite.RedisClient.Set("example3", "value", 0)

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
		val, err := suite.KeyValueService.Get(testCase.Key)
		if testCase.ExpectError {
			suite.NotNil(err, "should return error")
		} else {
			suite.Nil(err, "should get key")
			suite.Equal(testCase.ExpectedValue, string(val))
		}
	}
}

func (suite *KeyValueServiceTestSuite) TestKeyValueSet() {
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
		err := suite.KeyValueService.Set(testCase.Key, []byte(testCase.Value))
		suite.Nil(err, "should set key")
	}
}

func (suite *KeyValueServiceTestSuite) TestKeyValueSetEx() {
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
		err := suite.KeyValueService.SetEx(testCase.Key, []byte(testCase.Value), testCase.Expiration)
		suite.Nil(err, "should set key")

		val, err := suite.KeyValueService.Get(testCase.Key)
		suite.Nil(err, "should get key")
		suite.Equal(string(val), testCase.Value, "should get key")

		// wait for duration + 1
		time.Sleep(testCase.Expiration + 2)
		_, err = suite.KeyValueService.Get(testCase.Key)
		suite.NotNil(err, "should not get key")
	}
}
