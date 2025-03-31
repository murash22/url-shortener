package models

type UrlShortener struct {
	Id    int64
	Alias string
	Url   string
}

type User struct {
	Id       int64
	Email    string
	Password []byte
}
