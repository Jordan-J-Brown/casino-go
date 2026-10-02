package gamemodels

type ErrorCode string

const (
	ErrGameInProgress    ErrorCode = "ERR_GAME_IN_PROGRESS"
	ErrTooManyRequests   ErrorCode = "ERR_TOO_MANY_REQUESTS"
	ErrInvalidBet        ErrorCode = "ERR_INVALID_BET"
	ErrInsufficientFunds ErrorCode = "ERR_INSUFFICIENT_FUNDS"
	ErrGameNotFound      ErrorCode = "ERR_GAME_NOT_FOUND"
	ErrCannotCloseGame   ErrorCode = "ERR_CANNOT_CLOSE_GAME"
	ErrNextUnavailable   ErrorCode = "ERR_NEXT_UNAVAILABLE"
)

func (code ErrorCode) Error() string {
	return string(code)
}
