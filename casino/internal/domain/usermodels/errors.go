package usermodels

type Error string

func (err Error) Error() string {
	return string(err)
}

const (
	ErrUsernameExists  = Error("username exists")
	ErrInvalidUsername = Error("invalid username")
	ErrUsernameTaken   = Error("username is taken")
	ErrInvalidBalance  = Error("invalid balance")
	ErrUserNotFound    = Error("user not found")
	ErrSessionExists   = Error("user session exists")
	ErrBadArgs         = Error("bad arguments")
)
