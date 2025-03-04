package models

type User struct {
	Id           string `json:"id"`
	CreatedAt    int64  `json:"createdAt"`
	UserName     string `json:"userName"`
	PasswordHash string `json:"-"`
	Role         Role   `json:"role"`
}
