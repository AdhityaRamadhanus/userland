// +build integration

package redis_test

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/config"
	"github.com/AdhityaRamadhanus/userland/pkg/storage/redis"
	_redis "github.com/go-redis/redis"
	"github.com/stretchr/testify/suite"
)

type KeyValueServiceTestSuite struct {
	suite.Suite
	Config          *config.Configuration
	RedisClient     *_redis.Client
	KeyValueService userland.KeyValueService
}

func NewKeyValueServiceTestSuite(cfg *config.Configuration) *KeyValueServiceTestSuite {
	return &KeyValueServiceTestSuite{
		Config: cfg,
	}
}

func (suite *KeyValueServiceTestSuite) SetupTest() {
	if err := suite.RedisClient.FlushAll().Err(); err != nil {
		suite.T().Fatalf("RedisClient.FlushAll() err = %v; want nil", err)
	}
}

func (suite *KeyValueServiceTestSuite) SetupSuite() {
	suite.T().Logf("Connecting to redis at %v", suite.Config.Redis)
	redisClient, err := redis.CreateClient(suite.Config.Redis, 0)
	if err != nil {
		suite.T().Fatalf("redis.CreateClient() err = %v; want nil", err)
	}
	suite.RedisClient = redisClient
	suite.KeyValueService = redis.NewKeyValueService(redisClient)
}

func (suite *KeyValueServiceTestSuite) TestGet() {
	suite.RedisClient.Set("example1", "value", 0)

	type args struct {
		key string
	}
	testCases := []struct {
		name      string
		args      args
		wantValue string
		wantErr   bool
	}{
		{
			name: "success",
			args: args{
				key: "example",
			},
			wantErr: true,
		},
		{
			name: "return error",
			args: args{
				key: "example1",
			},
			wantValue: "value",
			wantErr:   false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			val, err := suite.KeyValueService.Get(tc.args.key)
			if err != nil && !tc.wantErr {
				t.Fatalf("KeyValueService.Get() err = %v; want nil", err)
			}
			if err == nil && tc.wantErr {
				t.Fatalf("KeyValueService.Get() err = nil; want not nil")
			}

			gotValue := string(val)
			wantValue := tc.wantValue
			if gotValue != wantValue {
				t.Errorf("KeyValueService.Get() = %s; want %s", gotValue, wantValue)
			}
		})
	}
}

func (suite *KeyValueServiceTestSuite) TestSet() {
	type args struct {
		key string
		val []byte
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				key: "example",
				val: []byte("value"),
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.KeyValueService.Set(tc.args.key, tc.args.val)
			if err != nil && !tc.wantErr {
				t.Fatalf("KeyValueService.Set() err = %v; want nil", err)
			}
			if err == nil && tc.wantErr {
				t.Fatalf("KeyValueService.Set() err = nil; want not nil")
			}
		})
	}
}

func (suite *KeyValueServiceTestSuite) TestSetEx() {
	type args struct {
		key string
		val []byte
		exp time.Duration
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				key: "example",
				val: []byte("value"),
				exp: 500 * time.Millisecond,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.KeyValueService.SetEx(tc.args.key, tc.args.val, tc.args.exp)
			if err != nil && !tc.wantErr {
				t.Fatalf("KeyValueService.SetEx() err = %v; want nil", err)
			}
			if err == nil && tc.wantErr {
				t.Fatalf("KeyValueService.SetEx() err = nil; want not nil")
			}

			if _, err := suite.KeyValueService.Get(tc.args.key); err != nil && err == userland.ErrKeyNotFound {
				t.Fatalf("KeyValueService.Get(notExpKey) err = %v; want nil", err)
			}

			time.Sleep(tc.args.exp + (100 * time.Millisecond))
			if _, err := suite.KeyValueService.Get(tc.args.key); err != userland.ErrKeyNotFound {
				t.Fatalf("KeyValueService.Get(expKey) err = nil; want %v", userland.ErrKeyNotFound.Error())
			}
		})
	}
}
func (suite *KeyValueServiceTestSuite) TestDelete() {
	suite.RedisClient.Set("example1", "value", 0)

	type args struct {
		key string
	}
	testCases := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				key: "example",
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			err := suite.KeyValueService.Delete(tc.args.key)
			if err != nil && !tc.wantErr {
				t.Fatalf("KeyValueService.Delete() err = %v; want nil", err)
			}
			if err == nil && tc.wantErr {
				t.Fatalf("KeyValueService.Delete() err = nil; want not nil")
			}

			if _, err := suite.KeyValueService.Get(tc.args.key); err != userland.ErrKeyNotFound {
				t.Fatalf("KeyValueService.Get(deletedKey) err = nil; want %v", userland.ErrKeyNotFound.Error())
			}
		})
	}
}
