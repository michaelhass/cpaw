package models

type Item struct {
	Id        string `json:"id"`
	CreatedAt int64  `json:"created_at"`
	Content   string `json:"content"`
}
