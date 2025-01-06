package models

type User struct {
	Id           string
	CreatedAt    int64
	UserName     string
	PasswordHash string
}
