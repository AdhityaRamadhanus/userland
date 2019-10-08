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
	AddToSortedSet(key string, value string, score float64) error
	DeleteFromSortedSet(key string, value string) error
	GetSortedSet(key string) ([]string, error)
}
