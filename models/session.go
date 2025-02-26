package models

type Session struct {
	ID        string
	ExpiresAt int64
	UserID    string
}
