package models

import "time"

type Session struct {
	Token     string
	ExpiresAt time.Time
	UserId    string
}
