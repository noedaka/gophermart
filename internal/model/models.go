package model

type ContextKey string

type UserCredentials struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string `json:"number"`
	Status     string `json:"status"`
	Accrual    int    `json:"accrual,omitempty"`
	UploadedAt string `json:"uploaded_at"`
}
