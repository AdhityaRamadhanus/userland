package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type KeyValueService struct {
	mock.Mock
}

func (m KeyValueService) Set(key string, value []byte) error {
	args := m.Called(key, value)

	return args.Get(0).(error)
}

func (m KeyValueService) SetEx(key string, value []byte, expirationInSeconds time.Duration) error {
	args := m.Called(key, value, expirationInSeconds)

	return args.Get(0).(error)
}

func (m KeyValueService) Delete(key string) error {
	args := m.Called(key)

	return args.Get(0).(error)
}

func (m KeyValueService) Get(key string) ([]byte, error) {
	args := m.Called(key)
	if args.Get(1) == nil {
		return args.Get(0).([]byte), nil
	}

	return nil, args.Get(1).(error)
}
