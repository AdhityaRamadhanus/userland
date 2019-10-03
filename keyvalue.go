package userland

import (
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("Key not found")
)

type KeyValueService interface {
	Set(key string, value []byte) error
	SetEx(key string, value []byte, expirationInSeconds time.Duration) error
	Get(key string) ([]byte, error)
	Delete(key string) error
}
