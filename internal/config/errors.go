package config

import "errors"

var (
	ErrOrderAlreadyUploadedByUser    = errors.New("order has been already created by current user")
	ErrOrderAlreadyUploadedByAnother = errors.New("order has been already created by other user")
	ErrNoOrders                      = errors.New("no orders found")
)
