package gamemodels

type ErrorCode string

const (
	ErrGameInProgress    ErrorCode = "ERR_GAME_IN_PROGRESS"
	ErrInvalidBet        ErrorCode = "ERR_INVALID_BET"
	ErrInsufficientFunds ErrorCode = "ERR_INSUFFICIENT_FUNDS"
	ErrGameNotFound      ErrorCode = "ERR_GAME_NOT_FOUND"
	ErrCannotCloseGame   ErrorCode = "ERR_CANNOT_CLOSE_GAME"
)

type Error struct {
	Code    ErrorCode
	Message string
}

func NewError(code ErrorCode, msg ...string) *Error {
	var m string
	if len(msg) == 1 {
		m = msg[0]
	}

	return &Error{
		Code:    code,
		Message: m,
	}
}

func (err *Error) Error() string {
	if err.Message == "" {
		return string(err.Code)
	}
	return string(err.Code) + ": " + err.Message
}

func IsGameError(err error) bool {
	_, ok := err.(*Error)
	return ok
}
