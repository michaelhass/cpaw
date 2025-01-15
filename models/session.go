package models

import "time"

type Session struct {
	token     string
	expiresAt time.Time
}
