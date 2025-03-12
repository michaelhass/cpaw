package models

type Session struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	UserId    string `json:"userId"`
}
