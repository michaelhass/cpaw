package models

type Item struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	Content   string `json:"content"`
	UserId    string `json:"userId"`
}
