package models

type User struct {
	UserPublic
	Pass string `json:"password"`
}

type UserPublic struct {
	Id  int    `json:"id"`
	Log string `json:"login"`
}
