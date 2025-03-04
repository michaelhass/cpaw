package models

type Session struct {
	Id        string `json:"id"`
	ExpiresAt int64  `json:"expiresAt"`
	UserId    string `json:"userId"`
}
