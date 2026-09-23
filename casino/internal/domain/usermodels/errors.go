package usermodels

import "errors"

var (
	ErrUsernameExists  = errors.New("username exists")
	ErrInvalidUsername = errors.New("invalid username")
	ErrUsernameTaken   = errors.New("username is taken")
	ErrInvalidBalance  = errors.New("invalid balance")
	ErrUserNotFound    = errors.New("user not found")
	ErrSessionExists   = errors.New("user session exists")
)
