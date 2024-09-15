package models

type User struct {
	UserID int
	Username string
	PasswordHash string
}

type Article struct {
	ArticleID int
	Title string
	Content string
	Author string
}