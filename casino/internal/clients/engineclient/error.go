package engineclient

type ErrorCode int8

const (
	ErrEngineError ErrorCode = iota
)

type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
