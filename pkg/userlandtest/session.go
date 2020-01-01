package userlandtest

import (
	"testing"
	"time"

	"github.com/AdhityaRamadhanus/userland"
	"github.com/AdhityaRamadhanus/userland/pkg/common/security"
)

var (
	DefaultSessionUserID     = 1
	DefaultSessionExpiration = security.UserAccessTokenExpiration
)

type SessionCreateConfig struct {
	UserID           int
	NumberOfSessions int
	Expiration       time.Duration
}

func WithNumberOfSessions(num int) func(*SessionCreateConfig) {
	return func(cfg *SessionCreateConfig) {
		cfg.NumberOfSessions = num
	}
}

func WithUserID(userID int) func(*SessionCreateConfig) {
	return func(cfg *SessionCreateConfig) {
		cfg.UserID = userID
	}
}

func WithExpiration(exp time.Duration) func(*SessionCreateConfig) {
	return func(cfg *SessionCreateConfig) {
		cfg.Expiration = exp
	}
}

func TestCreateSession(t *testing.T, sr userland.SessionRepository, opts ...func(*SessionCreateConfig)) userland.Session {
	cfg := SessionCreateConfig{
		UserID:           DefaultSessionUserID,
		Expiration:       DefaultSessionExpiration,
		NumberOfSessions: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	session := userland.Session{
		ID:         security.GenerateUUID(),
		Token:      "test",
		IP:         "123.123.13.123",
		ClientID:   1,
		ClientName: "test",
		Expiration: cfg.Expiration,
	}

	if err := sr.Create(cfg.UserID, session); err != nil {
		t.Fatalf("Failed to create session for user %d err = %v; want nil", cfg.UserID, err)
	}
	return session
}

func TestCreateSessions(t *testing.T, sr userland.SessionRepository, opts ...func(*SessionCreateConfig)) []userland.Session {
	sessions := []userland.Session{}
	cfg := SessionCreateConfig{
		UserID:           DefaultSessionUserID,
		Expiration:       DefaultSessionExpiration,
		NumberOfSessions: 1,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	for i := 0; i < cfg.NumberOfSessions; i++ {
		session := userland.Session{
			ID:         security.GenerateUUID(),
			Token:      "test",
			IP:         "123.123.13.123",
			ClientID:   1,
			ClientName: "test",
			Expiration: cfg.Expiration,
		}

		if err := sr.Create(cfg.UserID, session); err != nil {
			t.Fatalf("Failed to create session for user %d err = %v; want nil", cfg.UserID, err)
		}
		sessions = append(sessions, session)
	}

	return sessions
}
