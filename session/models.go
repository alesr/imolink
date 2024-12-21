package session

import "time"

type Session struct {
	ThreadID       string
	UserID         string
	LastAccessedAt time.Time
}
