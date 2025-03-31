package storage

import "errors"

var (
	ErrUrlNotFound  = errors.New("url not found")
	ErrUrlExists    = errors.New("url already exists")
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)
