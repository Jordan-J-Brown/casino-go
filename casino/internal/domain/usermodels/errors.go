package usermodels

type Error string

func (err Error) Error() string {
	return string(err)
}

const (
	ErrUsernameExists   = Error("ERR_USERNAME_EXISTS")
	ErrInvalidUsername  = Error("ERR_INVALID_USERNAME")
	ErrInvalidBalance   = Error("ERR_INVALID_BALANCE")
	ErrUserNotFound     = Error("ERR_USER_NOT_FOUND")
	ErrSessionExists    = Error("ERR_SESSION_EXISTS")
	ErrBadArgs          = Error("ERR_BAD_ARGS")
	ErrInvalidSessionID = Error("ERR_INVALID_SESSION_ID")
)
