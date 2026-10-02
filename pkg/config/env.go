package config

type Env string

const (
	Production Env = "prod"
	Local      Env = "local"
)

func (e Env) IsProd() bool {
	return e == Production
}
