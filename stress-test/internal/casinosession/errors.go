package casinosession

type Error string

const (
	ErrGameIsClosed  = Error("ERR_GAME_IS_CLOSED")
	ErrGameNotClosed = Error("ERR_GAME_NOT_CLOSED")
)

func (err Error) Error() string {
	return string(err)
}
