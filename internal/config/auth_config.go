package config

import (
	"gophermart/internal/model"
)

const UserIDKey model.ContextKey = "user_id"

var JWTSecret = []byte("my-super-secret-key-for-testing")
