package userland

import "time"

type Session struct {
	IsCurrent  bool
	IP         string
	ClientID   int
	ClientName string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Sessions []Session
