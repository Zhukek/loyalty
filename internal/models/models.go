package models

type User struct {
	Id   int
	Log  string `json:"login"`
	Pass string `json:"password"`
}
