package model

type UserCredentials struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}
