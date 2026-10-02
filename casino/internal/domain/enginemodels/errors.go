package enginemodels

type Error string

const (
	ErrEngineRegistered Error = "ERR_ENGINE_REGISTERED"
)

func (err Error) Error() string {
	return string(err)
}
